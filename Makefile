BINARY_NAME=subping
CGO_ENABLED=0
BUILD_FLAGS="-w -s"

build:
	go build -o out/$(BINARY_NAME) ./cmd/subping/

build-linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME) ./cmd/subping/
