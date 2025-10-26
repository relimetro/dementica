import grpc
from concurrent import futures
import aiProompt_pb2
import aiProompt_pb2_grpc
from grpc_reflection.v1alpha import reflection

port = "50052"
channel = grpc.insecure_channel('localhost:'+port)
stub = aiProompt_pb2_grpc.aiProomptStub(channel)
req = aiProompt_pb2.ProomptMsg(message="jeff") ############### WTF why no load
resp = stub.HealtcareProompt(req)
print(resp.message)



