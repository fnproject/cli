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

test:
	./test.sh

release:
	GOOS=linux go build -o fn_linux
	GOOS=darwin go build -o fn_mac
	GOOS=windows go build -o fn.exe
	# Uses fnproject/go:x.x-dev because golang:alpine has this issue: https://github.com/docker-library/golang/issues/155 and this https://github.com/docker-library/golang/issues/153
	docker run --rm -v ${PWD}:/go/src/github.com/fnproject/cli -w /go/src/github.com/fnproject/cli fnproject/go:1.9-dev go build -o fn_alpine

.PHONY: install
