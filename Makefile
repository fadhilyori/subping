BINARY_NAME=subping
CGO_ENABLED=0
VERSION=$(shell git describe --tags --abbrev=0)
BUILD_FLAGS="-w -s -X main.subpingVersion=$(VERSION)"
PLATFORMS = linux/amd64 linux/386 linux/arm64 linux/arm windows/amd64 windows/386 windows/arm64 windows/arm

build:
	go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME) ./cmd/subping/

build-all:
	mkdir -p out
	$(foreach platform,$(PLATFORMS),\
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(word 1,$(subst /, ,$(platform))) GOARCH=$(word 2,$(subst /, ,$(platform))) go build -ldflags=$(BUILD_FLAGS) -o out/$(BINARY_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform))) ./cmd/subping/;)