BINARY_NAME=subping
CGO_ENABLED=0
BUILD_FLAGS="-w -s"

build:
	go build -o out/$(BINARY_NAME) ./cmd/subping/

build-linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME)-linux-amd64 ./cmd/subping/
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=386 go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME)-linux-i386 ./cmd/subping/
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME)-linux-arm64 ./cmd/subping/
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME)-linux-armhf ./cmd/subping/
