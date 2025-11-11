.PHONY: proto server1 server2 client clean all

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hello.proto

# Build all binaries
build:
	go build -o bin/server ./server
	go build -o bin/client ./client

# Run first server on port 8888
server1:
	go run ./server -port 8888

# Run second server on port 9999
server2:
	go run ./server -port 9999

# Run client
client:
	go run ./client

# Clean generated files and binaries
clean:
	rm -rf proto/*.pb.go
	rm -rf bin/
