.PHONY: all test build clean

all: test build

build: 
	mkdir -p build
	go build -o build ./...

test:
	go test -v -coverprofile=tests/results/cover.out ./...

cover:
	go tool cover -html=tests/results/cover.out -o tests/results/cover.html

clean:
	rm -rf build/microservice
	go clean ./...

container:
	podman build -t quay.io/luigizuccarelli/golang-message-producer:1.16.3 .

push:
	podman push --authfile=/home/lzuccarelli/config.json  quay.io/luigizuccarelli/golang-message-producer:1.16.3 
