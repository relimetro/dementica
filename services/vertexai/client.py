import grpc
from concurrent import futures
import vertex_pb2
import vertex_pb2_grpc
from grpc_reflection.v1alpha import reflection

port = "50052"
channel = grpc.insecure_channel('localhost:'+port)
stub = vertex_pb2_grpc.aiProomptStub(channel)
req = vertex_pb2.ProomptMsg(message="jeff") ############### WTF why no load
resp = stub.Proompt(req)
print(resp.message)



