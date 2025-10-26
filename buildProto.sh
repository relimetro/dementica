# protoc --include_imports --include_source_info -o proto/firebase.pb proto/proto.proto

GOOGLEAPIS_DIR="./googleapis"

protoc \
	-I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto/firebase.pb \
	--proto_path=./proto/ \
	--go_out=./services/firebaseprototype/protoOut --go_opt=paths=source_relative \
	--go-grpc_out=./services/firebaseprototype/protoOut --go-grpc_opt=paths=source_relative \
	firebase.proto

# protoc -I${GOOGLEAPIS_DIR} -I. --include_imports --include_source_info --descriptor_set_out=proto.pb proto/firebase.proto # ../../proto/dummy.proto
