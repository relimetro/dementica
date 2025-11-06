# protoc --include_imports --include_source_info -o proto/firebase.pb proto/proto.proto

GOOGLEAPIS_DIR="./googleapis"

# firebase
protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/firebase.pb \
	--proto_path=./proto/ \
	--go_out=./services/firebaseprototype/protoOut --go_opt=paths=source_relative \
	--go-grpc_out=./services/firebaseprototype/protoOut --go-grpc_opt=paths=source_relative \
	firebase.proto

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
	--go_out=./services/firebaseprototype/protoAI --go_opt=paths=source_relative \
	--go-grpc_out=./services/firebaseprototype/protoAI --go-grpc_opt=paths=source_relative \
	vertex.proto

# failed attempt to get firebase & dummy both working
protoc -I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/both.pb proto/firebase.proto proto/dummy.proto
