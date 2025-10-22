
python -m grpc_tools.protoc \
	--proto_path=. ./aiProompt.proto \
	--python_out=. --grpc_python_out=.

