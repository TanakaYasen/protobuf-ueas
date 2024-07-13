export PATH:=${PATH}:$(shell pwd)/protoc-gen-arpc

all:
	echo $(shell pwd)
	cd protoc-gen-as && go build
	cd protoc-gen-arpc && go build
	
	protoc --gogofast_out=./generated/ proto/game.proto
	protoc --arpc_out=generated proto/game.proto
	
	mv generated/*.go generated/game
	
	cd example/goserver && go build
	cd example/goclient && go build
	
test-wire:
	# protoc --cpp_out=generated --proto_path=proto proto/game.proto
	protoc --as_out=./generated/ proto/game.proto
	
netlibx:
	@echo ${PATH}
	cd netlib && go build