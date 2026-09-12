.PHONY: build run test

build:
	@go build -o bin/fs ./cmd/api
run: build
	./bin/fs
test:
	@go test ./test