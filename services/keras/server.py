import grpc
from concurrent import futures
import vertex_pb2
import vertex_pb2_grpc
from grpc_reflection.v1alpha import reflection


class AiProompt(vertex_pb2_grpc.aiProomptServicer):
	def HealtcareProompt(self, request, context):
		out = "1"
		print("recieved AI response")
		print(f"response {out}")
		return vertex_pb2.ProomptReturn(message=out)


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

	port = "50053"
	server.add_insecure_port('[::]:'+port)
	server.start()
	print("Keras running on port "+port)
	server.wait_for_termination()


if __name__ == "__main__":
	serve()



