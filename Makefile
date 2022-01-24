BINARY = "messier"

run:
	@go fmt ./
	@go run ./

build:
	@go fmt ./
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY}

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

protoc:
	@protoc --proto_path=./proto \
	--go_out=./proto/pb --go_opt=paths=source_relative \
	--go-grpc_out=./proto/pb --go-grpc_opt=paths=source_relative ./proto/*.proto

help:
	@echo "make run - format these code and use go run to running"
	@echo "make build - compile go code, create runable binary file"
	@echo "make clean - delete binary file and clean the work directory"
	@echo "make protoc - use protobuf files to generate go grpc package"