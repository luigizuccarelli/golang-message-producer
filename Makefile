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

docker:
	docker build -t nexus-registry-nexus.apps.aws2-dev.ocp.14west.io/message-consumer-couchbase:1.13.1 .

push:
	docker push nexus-registry-nexus.apps.aws2-dev.ocp.14west.io/message-consumer-couchbase:1.13.1 
