VERSION := $(shell git describe --tags --dirty)

RELEASE_DIR := releases/$(VERSION)
RELEASE_BIN := releases/$(VERSION)/bin

build:
	go fmt
	go build -o bin/mcp-alkira

fmt:
	go fmt

vendor: GOPRIVATE=github.com/alkiranet
vendor:
	go mod tidy
	go mod vendor

superclean:
	git clean -x -d -f
