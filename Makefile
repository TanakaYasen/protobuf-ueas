CD := `pwd`

all:
	echo ${CD}
	go build
	PATH=${PATH}:/home/pxf_god/repo/protobuf-as protoc --as_out=./ addressbook.proto