import grpc
from concurrent import futures
import vertex_pb2
import vertex_pb2_grpc
from grpc_reflection.v1alpha import reflection
from google import genai
from google.genai import types
import os

os.environ["GOOGLE_APPLICATION_CREDENTIALS"]="./project-faf28b74-121d-4ea8-b17-fdd4d4d791e5.json"

class AiProompt(vertex_pb2_grpc.aiProomptServicer):
	def Proompt(self, request, context):
		req = request.message
		print(f"recieved {req}")

		client = genai.Client(
				vertexai=True,
				api_key=os.environ.get("GOOGLE_CLOUD_API_KEY"),
				project="project-faf28b74-121d-4ea8-b17",
				location="us-central1",
				)


		msg1_text1 = types.Part.from_text(text=req)
		si_text1 = """you are an accurate predictor of dementia. given a users input you will output a number between 0 and 30."""
		model = "projects/192856652580/locations/us-central1/endpoints/669363987292356608"

		contents = [
			types.Content(
			role="user",
			parts=[ msg1_text1 ]
			),
		]
		generate_content_config = types.GenerateContentConfig(
				temperature = 1,
				top_p = 1,
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
							system_instruction=[types.Part.from_text(text=si_text1)],
							thinking_config=types.ThinkingConfig(
								thinking_budget=0,
								),
							)

		out = ""
		for chunk in client.models.generate_content_stream(
				model = model,
				contents = contents,
				config = generate_content_config,
				):
			if not chunk.candidates or not chunk.candidates[0].content or not chunk.candidates[0].content.parts:
				continue
			out += chunk.text


		print("recieved AI response")
		print(f"response {out}")
		out2 = str(1.0-(int(out)/30))
		print(f"convert to percent {out2}")
		return vertex_pb2.ProomptReturn(message=out2)



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



