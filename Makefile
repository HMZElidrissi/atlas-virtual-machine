.PHONY: build run test proto clean

## build: compile all packages
build:
	go build ./...

## run: build and run the full pipeline (parse → compile → VM → consensus)
run:
	go run ./cmd/atlasvm/

## test: run the full test suite with verbose output
test:
	go test -v ./...

## proto: regenerate Go code from proto/atlas.proto
##        requires protoc, protoc-gen-go and protoc-gen-go-grpc
proto:
	protoc --go_out=. --go-grpc_out=. proto/atlas.proto

## clean: remove compiled binaries and test caches
clean:
	go clean -testcache
	rm -f atlasvm

help:
	@grep -E '^##' Makefile | sed 's/## //'
