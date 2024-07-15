export PATH:=${PATH}:$(shell pwd)/protoc-gen-arpc

go:
	cd protoc-gen-arpc && go build
	
	protoc --gogofast_out=generated proto/game.proto
	protoc --arpc_out=generated --arpc_opt=go proto/game.proto
	
	cd example/goserver && go build
	cd example/goclient && go build
	
as:
	# protoc --cpp_out=generated --proto_path=proto proto/game.proto
	protoc --as_out=generated proto/game.proto
	
net:
	@echo ${PATH}
	cd netlib && go build

cpp:
	cd protoc-gen-arpc && go build
	protoc --cpp_out=generated --proto_path=proto proto/game.proto
	protoc --arpc_out=generated --proto_path=proto --arpc_opt=cpp proto/game.proto
	cd example/cppclient && g++ -g *.cpp ../../generated/*.cc ../../generated/*.cpp ../../netlib/*.cc -lprotobuf