import os
import grpc
from concurrent import futures
import requests
from grpc_reflection.v1alpha import reflection

import user_service_pb2
import user_service_pb2_grpc

import firebase_admin
from firebase_admin import credentials, auth, initialize_app

cred_path = os.getenv("FIREBASE_CREDENTIALS", "firebase.json")
web_api_key = os.getenv("FIREBASE_WEB_API_KEY")

cred = credentials.Certificate(cred_path)
initialize_app(cred)

SIGNIN_URL = f"https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={web_api_key}"

class AuthService(user_service_pb2_grpc.AuthServiceServicer):
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
            return user_serive_pb2.AuthReply(
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


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    user_service_pb2_grpc.add_AuthServiceServicer_to_server(AuthService(), server)

    SERVICE_NAMES = (
        user_service_pb2.DESCRIPTOR.services_by_name["AuthService"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)

    server.add_insecure_port("[::]:50061")
    server.start()
    print("AuthService running on port 50061")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()

