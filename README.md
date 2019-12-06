# Trackmate message producer golang microservice

A simple golang microservice to publish data to a message queue. 


## Usage 

```bash
# cd to project directory and build executable
$ go build -o microservice .

```

## Docker build

```bash
# I use the version of golang as a tag
docker build -t <your-registry-id>/trackmate-message-producer:1.13.1 .

```

## Curl timing usage
```
curl -w "@curl-timing.txt" -o /dev/null -s "http://site-to-test

```

## Executing tests
```bash
# clear the cache - this is optional
go clean -testcache
go test -v schema.go validate.go validate_test.go handlers.go handlers_test.go -coverprofile tests/results/cover.out
go tool cover -html=tests/results/cover.out -o tests/results/cover.html
# run sonarqube scanner (assuming sonarqube server is running)
# NB the SonarQube host and login will differ - please update it accordingly 
 ~/<path-to-sonarqube>/sonar-scanner-3.3.0.1492-linux/bin/sonar-scanner  -Dsonar.projectKey=trackmate-message-producer  -Dsonar.sources=.   -Dsonar.host.url=http://<url-to-server>   -Dsonar.login=<token> -Dsonar.go.coverage.reportPaths=tests/results/cover.out -Dsonar.exclusions=vendor/**,*_test.go,micorservice,connectors.go,tests/**

```
## Testing container 
```bash

# start the container
# curl the isalive endpoint
curl -k -H 'Token: xxxxx' -w '@curl-timing.txt'  http://127.0.0.1:9000/api/v1/sys/info/isalive

```
