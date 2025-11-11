
GOOGLEAPIS_DIR="./googleapis"



# firestore
protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/firestore.pb \
	--proto_path=./proto/ \
	--go_out=./services/firestore/protoOut --go_opt=paths=source_relative \
	--go-grpc_out=./services/firestore/protoOut --go-grpc_opt=paths=source_relative \
	firestore.proto



# vertexAI (python)
python -m grpc_tools.protoc \
	--proto_path=./proto/ ./proto/vertex.proto \
	--python_out=./services/vertexai --grpc_python_out=./services/vertexai

# vertexAI (keras)
python -m grpc_tools.protoc \
	--proto_path=./proto/ ./proto/vertex.proto \
	--python_out=./services/keras --grpc_python_out=./services/keras

# vertexAI (firebaseprototype)
protoc \
	--proto_path=./proto \
	--go_out=./services/firestore/protoAI --go_opt=paths=source_relative \
	--go-grpc_out=./services/firestore/protoAI --go-grpc_opt=paths=source_relative \
	vertex.proto

# todo descriptor
#protoc -I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/descriptor.pb proto/firestore.proto proto/user_service.proto
protoc -I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/both.pb proto/firestore.proto proto/user_service.proto
