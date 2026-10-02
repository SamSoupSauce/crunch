.PHONY: all build run test clean

all: build

build:
	go build -o server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -f server
