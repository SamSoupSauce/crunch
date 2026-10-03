.PHONY: all build build-all build-darwin build-linux build-windows run test clean

all: build

build:
	go build -o server ./cmd/server

build-all: build-darwin build-linux build-windows

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/server-darwin-x64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o bin/server-darwin-arm64 ./cmd/server

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/server-linux-x64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -o bin/server-linux-arm64 ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/server-windows-x64.exe ./cmd/server
	GOOS=windows GOARCH=arm64 go build -o bin/server-windows-arm64.exe ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf server bin

