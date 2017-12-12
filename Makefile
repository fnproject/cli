all: dep build	
	./fn

build: 
	go build -o fn

install:
	go build -o ${GOPATH}/bin/fn

docker: 
	docker build -t fnproject/fn:latest .

dep:
	dep ensure

dep-up:
	dep ensure --update

test:
	./test.sh

release:
	GOOS=linux go build -o fn_linux
	GOOS=darwin go build -o fn_mac
	GOOS=windows go build -o fn.exe
	docker run --rm -v ${PWD}:/go/src/github.com/fnproject/cli -w /go/src/github.com/fnproject/cli golang:alpine go build -o fn_alpine

.PHONY: install
