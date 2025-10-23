import grpc
from concurrent import futures
import aiProompt_pb2
import aiProompt_pb2_grpc
from grpc_reflection.v1alpha import reflection
import vertexAI

class AiProompt(aiProompt_pb2_grpc.aiProomptServicer):
    def HealtcareProompt(self, request, context):
        print(f"recieved {request}")
        return aiProompt_pb2.ProomptReturn(Message=f"Hello, {request.name}!") # dummy
		# return aiProompt_pb2.Proompt(Message=vertexAI.FTproompt(request.name)) # vertexAI

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    aiProompt_pb2_grpc.add_aiProomptServicer_to_server(AiProompt(), server)
    # dummy_pb2_grpc.add_HelloServiceServicer_to_server(HelloService(), server)

    SERVICE_NAMES = (
		aiProompt_pb2_grpc.aiProomptServicer.__name__,
        # dummy_pb2_grpc.HelloServiceServicer.__name__,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)

    port = "50052"
    server.add_insecure_port('localhost:'+port)
    server.start()
    print("VertexAI running on port "+port)
    server.wait_for_termination()

if __name__ == "__main__":
    serve()



