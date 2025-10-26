import grpc
from concurrent import futures
import vertex_pb2
import vertex_pb2_grpc
from grpc_reflection.v1alpha import reflection
from google import genai
from google.genai import types
# import vertexAI
import os

os.environ["GOOGLE_APPLICATION_CREDENTIALS"]="./copper-actor-475117-i7-92a1502a7bf4.json"

class AiProompt(vertex_pb2_grpc.aiProomptServicer):
    def HealtcareProompt(self, request, context):
        req = request.message
        print(f"recieved {req}")
        print("send request")

        client = genai.Client(
                vertexai=True,
                api_key=os.environ.get("GOOGLE_CLOUD_API_KEY"),
                project='copper-actor-475117-i7', location='us-central1'
                )

        model = "projects/585981786057/locations/us-central1/endpoints/3968791674062110720"
        contents = [
                types.Content(
                    role="user",
                    parts=[ types.Part.from_text(text=req) ]
                    )
                ]

        generate_content_config = types.GenerateContentConfig(
                temperature = 1,
                top_p = 0.95,
                max_output_tokens = 65535,
                safety_settings = [types.SafetySetting(
                    category="HARM_CATEGORY_HATE_SPEECH",
                    threshold="OFF"
                    ),types.SafetySetting(
                        category="HARM_CATEGORY_DANGEROUS_CONTENT",
                        threshold="OFF"
                        ),types.SafetySetting(
                            category="HARM_CATEGORY_SEXUALLY_EXPLICIT",
                            threshold="OFF"
                            ),types.SafetySetting(
                                category="HARM_CATEGORY_HARASSMENT",
                                threshold="OFF"
                                )],
                            thinking_config=types.ThinkingConfig( thinking_budget=0, ),
                            )

        out = ""
        for chunk in client.models.generate_content_stream(
                model = model,
                contents = contents,
                config = generate_content_config,
                ):
            txt = chunk.text
            out += txt



        print("recieved AI response")
        print(f"response {out}")
        return vertex_pb2.ProomptReturn(message=out)
# return vertex_pb2.ProomptReturn(Message=f"Hello, {request.name}!") # dummy

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    vertex_pb2_grpc.add_aiProomptServicer_to_server(AiProompt(), server)
    # dummy_pb2_grpc.add_HelloServiceServicer_to_server(HelloService(), server)

    SERVICE_NAMES = (
        vertex_pb2_grpc.aiProomptServicer.__name__,
        # dummy_pb2_grpc.HelloServiceServicer.__name__,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)

    port = "50052"
    server.add_insecure_port('[::]:'+port)
    server.start()
    print("VertexAI running on port "+port)
    server.wait_for_termination()

if __name__ == "__main__":
    serve()



