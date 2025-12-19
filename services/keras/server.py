import grpc
from concurrent import futures
import vertex_pb2
import vertex_pb2_grpc
from grpc_reflection.v1alpha import reflection

from kerasPredict import kerasRun
from datatypes import LifestyleQuestionare, LifestyleQuestionareFromIntermediary



class AiProompt(vertex_pb2_grpc.aiProomptServicer):
	def Proompt(self, request, context):
		print("recieved request",request.message)

		lifestyle: LifestyleQuestionare = LifestyleQuestionareFromIntermediary(request.message)
		riskScore: str = kerasRun(lifestyle)

		print(f"response {riskScore}")
		return vertex_pb2.ProomptReturn(message=riskScore)



def serve():
	server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
	vertex_pb2_grpc.add_aiProomptServicer_to_server(AiProompt(), server)

	SERVICE_NAMES = (
		vertex_pb2_grpc.aiProomptServicer.__name__,
		reflection.SERVICE_NAME,
	)
	reflection.enable_server_reflection(SERVICE_NAMES, server)

	port = "50053"
	server.add_insecure_port('[::]:'+port)
	server.start()
	print("Keras running on port "+port)
	server.wait_for_termination()


if __name__ == "__main__":
	serve()



