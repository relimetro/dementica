import os
import grpc
from concurrent import futures
import requests
from grpc_reflection.v1alpha import reflection

import user_service_pb2
import user_service_pb2_grpc

import firebase_admin
from firebase_admin import credentials, auth, initialize_app, firestore

cred_path = os.getenv("FIREBASE_CREDENTIALS", "firebase.json")
web_api_key = os.getenv("FIREBASE_WEB_API_KEY")

cred = credentials.Certificate(cred_path)
initialize_app(cred)
db = firestore.client()

SIGNIN_URL = f"https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={web_api_key}"

def verify_token(id_token):
        try:
            decoded = auth.verify_id_token(id_token)
            return decoded["uid"]
        except Exception as e:
            raise ValueError(f"Invalid token: {str(e)}")


class UserService(user_service_pb2_grpc.UserServiceServicer):
    def SignUp(self, request, context):
        try:
            user = auth.create_user(
                email=request.email,
                password=request.password
            )
            return user_service_pb2.AuthReply(
                uid=user.uid,
                message=f"User {request.email} created successfully."
            )
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.AuthReply(message="Signup failed.")

    def Login(self, request, context):
        try:
            payload = {
                "email": request.email,
                "password": request.password,
                "returnSecureToken": True,
            }
            r = requests.post(SIGNIN_URL, json=payload)
            r.raise_for_status()
            data = r.json()
            return user_service_pb2.AuthReply(
                uid=data.get("localId"),
                id_token=data.get("idToken"),
                message=f"User {request.email} logged in successfully."
            )
        except requests.exceptions.HTTPError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.AuthReply(message="Invalid email or password.")
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.AuthReply(message="Login failed.")

    def LinkUser(self, request, context):
        """Patient links themselves to a doctor using their token."""
        try:
            patient_uid = verify_token(request.patient_token)
            doctor_uid = request.doctor_uid
            relation_type = request.relation_type or "doctor"

            # Store in Firestore
            doc_ref = db.collection("user_relations").document(patient_uid)
            doc = doc_ref.get()
            relations = doc.to_dict().get("relations", []) if doc.exists else []

            # Prevent duplicates
            if any(r["related_uid"] == doctor_uid for r in relations):
                return user_service_pb2.LinkUserReply(
                    message="Doctor already linked to this patient."
                )

            relations.append({
                "related_uid": doctor_uid,
                "relation_type": relation_type
            })

            doc_ref.set({"relations": relations})
            return user_service_pb2.LinkUserReply(
                message=f"Linked patient {patient_uid} to doctor {doctor_uid}."
            )

        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.LinkUserReply(message="Invalid token.")
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.LinkUserReply(message="Failed to link users.")

    def GetLinkedUsers(self, request, context):
        try:
            user_uid = verify_token(request.id_token)
            relation_type = request.relation_type

            # Fetch direct relations (same as before)
            doc_ref = db.collection("user_relations").document(user_uid)
            doc = doc_ref.get()
            direct_relations = doc.to_dict().get("relations", []) if doc.exists else []

            # Fetch reverse relations: find docs where this user is listed in "relations"
            reverse_docs = db.collection("user_relations").stream()
            reverse_relations = []
            for d in reverse_docs:
                rels = d.to_dict().get("relations", [])
                for r in rels:
                    if r["related_uid"] == user_uid:
                        reverse_relations.append({
                            "related_uid": d.id,
                            "relation_type": r["relation_type"]
                        })

            all_relations = direct_relations + reverse_relations

            if relation_type:
                all_relations = [r for r in all_relations if r["relation_type"] == relation_type]

            related_users = [
                user_service_pb2.RelatedUser(
                    uid=r["related_uid"],
                    relation_type=r["relation_type"]
                )
                for r in all_relations
            ]

            return user_service_pb2.GetLinkedUsersReply(related_users=related_users)

        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.GetLinkedUsersReply()
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.GetLinkedUsersReply()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    user_service_pb2_grpc.add_UserServiceServicer_to_server(UserService(), server)

    SERVICE_NAMES = (
        user_service_pb2.DESCRIPTOR.services_by_name["UserService"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)

    server.add_insecure_port("[::]:50061")
    server.start()
    print("AuthService running on port 50061")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()

