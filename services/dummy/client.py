import grpc
from concurrent import futures
import dummy_pb2
import dummy_pb2_grpc
from grpc_reflection.v1alpha import reflection

channel = grpc.insecure_channel("localhost:50051")
stub = dummy_pb2_grpc.HelloServiceStub(channel)
req = dummy_pb2.HelloRequest(name="jeff")
resp = stub.SayHello(req)
print(resp.message)
