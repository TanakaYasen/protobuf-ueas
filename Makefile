all:
	echo $(shell pwd)
	cd protoc-gen-as && go build
	cd protoc-gen-arpc && go build
	
	protoc --gogofast_out=./generated/ proto/game.proto
	PATH=${PATH}:$(shell pwd)/protoc-gen-arpc/ protoc --arpc_out=generated --arpc_opt=xxxxxx proto/game.proto
	
	mv generated/*.go generated/game
	
	cd example/goserver && go build
	cd example/goclient && go build
	
test-wire:
	# protoc --cpp_out=generated --proto_path=proto proto/game.proto
	# PATH=${PATH}:$(shell pwd)/protoc-gen-as/ protoc --as_out=./generated/ proto/game.proto
	
w:
	cd example/goserver && go build