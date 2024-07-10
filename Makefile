all:
	echo $(shell pwd)
	cd protoc-gen-as && go build
	cd protoc-gen-host && go build
	
	protoc --gogofast_out=./generated/ proto/game.proto
	# protoc --cpp_out=generated --proto_path=proto proto/game.proto
	# PATH=${PATH}:$(shell pwd)/protoc-gen-as/ protoc --as_out=./generated/ proto/game.proto
	PATH=${PATH}:$(shell pwd)/protoc-gen-host/ protoc --host_out=param=myValue,otherParam=anotherValue:generated proto/game.proto
	
	mv generated/*.go generated/game
	
	cd example/goserver && go build
	cd example/goclient && go build
	
test-wire:
	
	
w:
	cd example/goserver && go build