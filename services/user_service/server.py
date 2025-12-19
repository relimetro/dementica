from datetime import datetime
import os
import grpc
from concurrent import futures
import requests
from grpc_reflection.v1alpha import reflection

import user_service_pb2
import user_service_pb2_grpc

import firebase_admin
from firebase_admin import credentials, auth, initialize_app, firestore

import logging

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

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
    def AddTestResult(self, request, context):
        self.log_request("AddTestResult", request)
        try:
            # Verify token → get user ID
            uid = verify_token(request.id_token)

            # Prepare document
            doc = {
                "user_id": uid,
                "data": request.data,
                "risk_score": request.risk_score,
                "date": firestore.SERVER_TIMESTAMP
            }

            # Insert into TestResults collection
            db.collection("TestResults").add(doc)

            return user_service_pb2.AddTestResultReply(
                message="Test result stored successfully."
            )

        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.AddTestResultReply(
                message="Invalid token."
            )

        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.AddTestResultReply(
                message="Failed to store test result."
            )

    def log_request(self, method_name, request):
        try:
            from google.protobuf.json_format import MessageToDict
            payload = MessageToDict(request, preserving_proto_field_name=True)
        except Exception:
            payload = str(request)
        logging.info(f"Received gRPC request: {method_name} | payload: {payload}")

    def VerifyTokenRemote(self, request, context): # cathal added this :), for firestore to get user id from id_token
        self.log_request("VerifyTokenRemote", request)
        try:
            uid = verify_token(request.id_token)
            return user_service_pb2.VerifyTokenResponse(
                res=True,
                uid=uid
            )
        except ValueError as e:
            return user_service_pb2.VerifyTokenResponse(
                res=False,
                uid="__INVALID__"
            )

    def SignUp(self, request, context):
        self.log_request("SignUp", request)
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
            return user_service_pb2.AuthReply(message="Signup failed.") # note: firebase expects this exact message if signup if not successful

    def Login(self, request, context):
        self.log_request("Login", request)
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
            print(e)
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.AuthReply(message="Invalid email or password.")
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.AuthReply(message="Login failed.")

    def LinkUser(self, request, context):
        self.log_request("LinkUser", request)
        try:
            patient_uid = verify_token(request.patient_token)
            doctor_uid = request.doctor_uid
            relation_type = request.relation_type or "doctor"

            doc_ref = db.collection("user_relations").document(patient_uid)
            doc = doc_ref.get()
            relations = doc.to_dict().get("relations", []) if doc.exists else []

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
        self.log_request("GetLinkedUsers", request)
        try:
            requester_uid = verify_token(request.id_token)
            relation_type = request.relation_type  # optional filter

            requester_doc = db.collection("Users").document(requester_uid).get()
            if not requester_doc.exists:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("Requester not found.")
                return user_service_pb2.GetLinkedUsersReply()

            requester = requester_doc.to_dict()
            requester_type = requester.get("Type")

            related_users = []

            if requester_type == "Admin":
                query = db.collection("Users")
                if relation_type:
                    query = query.where("Type", "==", relation_type)

                for doc in query.stream():
                    related_users.append(
                        user_service_pb2.RelatedUser(
                            uid=doc.id,
                            relation_type=doc.to_dict().get("Type", "unknown").lower()
                        )
                    )

            elif requester_type == "Doctor":
                query = db.collection("Users").where("DoctorID", "==", requester_uid)

                if relation_type:
                    query = query.where("Type", "==", relation_type)

                for doc in query.stream():
                    related_users.append(
                        user_service_pb2.RelatedUser(
                            uid=doc.id,
                            relation_type="patient"
                        )
                    )

            elif requester_type == "Patient":
                doctor_uid = requester.get("DoctorID")
                if doctor_uid:
                    related_users.append(
                        user_service_pb2.RelatedUser(
                            uid=doctor_uid,
                            relation_type="doctor"
                        )
                    )

            else:
                context.set_code(grpc.StatusCode.PERMISSION_DENIED)
                context.set_details("Unknown user role.")
                return user_service_pb2.GetLinkedUsersReply()

            return user_service_pb2.GetLinkedUsersReply(
                related_users=related_users
            )

        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.GetLinkedUsersReply()

        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.GetLinkedUsersReply()

    def AddUserDetails(self, request, context):
        self.log_request("AddUserDetails", request)
        try:
            uid = verify_token(request.id_token)
            data = dict(request.details)

            # Replace or create the user's details document
            db.collection("user_details").document(uid).set(data)

            return user_service_pb2.AddUserDetailsReply(
                message=f"User details saved for {uid}."
            )
        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.AddUserDetailsReply(message="Invalid token.")
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.AddUserDetailsReply(message="Failed to save user details.")

    def GetUserDetails(self, request, context):
        self.log_request("GetUserDetails", request)
        try:
            requester_uid = verify_token(request.id_token)
            target_uid = request.target_uid or requester_uid

            # Check access if requester != target
            if target_uid != requester_uid:
                rel_doc = db.collection("user_relations").document(target_uid).get()
                is_related = False
                if rel_doc.exists:
                    relations = rel_doc.to_dict().get("relations", [])
                    is_related = any(r["related_uid"] == requester_uid for r in relations)
                # Also check reverse relationships
                if not is_related:
                    rev_doc = db.collection("user_relations").document(requester_uid).get()
                    if rev_doc.exists:
                        rels = rev_doc.to_dict().get("relations", [])
                        is_related = any(r["related_uid"] == target_uid for r in rels)
                if not is_related:
                    context.set_code(grpc.StatusCode.PERMISSION_DENIED)
                    context.set_details("Requester not related to target user.")
                    return user_service_pb2.GetUserDetailsReply()

            doc = db.collection("user_details").document(target_uid).get()
            if not doc.exists:
                return user_service_pb2.GetUserDetailsReply(details={})

            return user_service_pb2.GetUserDetailsReply(details=doc.to_dict())
        except ValueError as e:
            context.set_code(grpc.StatusCode.UNAUTHENTICATED)
            context.set_details(str(e))
            return user_service_pb2.GetUserDetailsReply()
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return user_service_pb2.GetUserDetailsReply()


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

