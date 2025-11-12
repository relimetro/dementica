
GOOGLEAPIS_DIR="./googleapis"



# firestore
protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/firestore.pb \
	--proto_path=./proto/ \
	--go_out=./services/firestore/protoOut --go_opt=paths=source_relative \
	--go-grpc_out=./services/firestore/protoOut --go-grpc_opt=paths=source_relative \
	firestore.proto

# firestore (UserService)
protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/user_service.pb \
	--proto_path=./proto/ \
	--go_out=./services/firestore/UserService --go_opt=paths=source_relative \
	--go-grpc_out=./services/firestore/UserService --go-grpc_opt=paths=source_relative \
	user_service.proto



# user_service
python -m grpc_tools.protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/user_service.pb \
	--proto_path=./proto/ \
	--python_out=./services/user_service \
	--grpc_python_out=./services/user_service \
	user_service.proto
 # can delete proto/user_service.pb afterward, just need so does not complain about annotation.proto



# aiProompt (vertexAI) (with two o's lmao)
python -m grpc_tools.protoc \
	--proto_path=./proto/ ./proto/vertex.proto \
	--python_out=./services/vertexai --grpc_python_out=./services/vertexai

# aiProompt (keras)
python -m grpc_tools.protoc \
	--proto_path=./proto/ ./proto/vertex.proto \
	--python_out=./services/keras --grpc_python_out=./services/keras

# aiProompt (firebaseprototype)
protoc \
	--proto_path=./proto \
	--go_out=./services/firestore/protoAI --go_opt=paths=source_relative \
	--go-grpc_out=./services/firestore/protoAI --go-grpc_opt=paths=source_relative \
	vertex.proto



# descriptor
protoc -I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/descriptors.pb proto/firestore.proto proto/user_service.proto
