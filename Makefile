BINARY_NAME := dist/projconf
PKG := github.com/train360-corp/projconf

ifndef VERSION
  $(error VERSION is required. Usage: make VERSION=x.y.z)
endif

LDFLAGS := -X $(PKG)/internal/version.Version=$(VERSION)

.PHONY: all build run clean test help

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

help:
	@echo "Usage:"
	@echo "  make VERSION=x.y.z        - Build with a specific version"
	@echo "  make run VERSION=x.y.z    - Build and run"
	@echo "  make clean                - Remove the binary"
	@echo "  make test                 - Run tests"
	@echo "  make help                 - Show this message"