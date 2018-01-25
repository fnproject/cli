all: dep build	
	./fn

build: 
	go build -o fn

install:
	go build -o ${GOPATH}/bin/fn

docker: 
	docker build -t fnproject/fn:latest .

dep:
	dep ensure --vendor-only

dep-up:
	dep ensure

test-cross-compile:
	GOOS=windows go build -o fn
	GOOS=linux GOARCH=arm64 go build -o fn
	GOOS=darwin go build -o fn
	GOOS=linux go build -o fn
	# clean up
	rm -fr fn

test: test-cross-compile build
	./test.sh

release:
	GOOS=linux go build -o fn_linux
	GOOS=darwin go build -o fn_mac
	GOOS=windows go build -o fn.exe
	GOOS=linux GOARCH=arm64 go build -o fn_linux_arm64
	# Uses fnproject/go:x.x-dev because golang:alpine has this issue: https://github.com/docker-library/golang/issues/155 and this https://github.com/docker-library/golang/issues/153
	docker run --rm -v ${PWD}:/go/src/github.com/fnproject/cli -w /go/src/github.com/fnproject/cli fnproject/go:1.9-dev go build -o fn_alpine

.PHONY: install test build
