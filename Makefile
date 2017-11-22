all: vendor build	
	./fn

build: 
	go build -o fn

install:
	go build -o ${GOPATH}/bin/fn

docker: 
	docker build -t fnproject/fn:latest .

dep:
	glide install -v

dep-up:
	glide up -v

test:
	GOOS=windows go build -o fn
	GOOS=linux GOARCH=amd64 go build -o fn
	GOOS=darwin go build -o fn
	GOOS=linux go build -o fn
	go build -o fn
	./test.sh

release:
	GOOS=linux go build -o fn_linux
	GOOS=darwin go build -o fn_mac
	GOOS=windows go build -o fn.exe
	GOOS=linux GOARCH=amd64 go build -o fn_linux_arm64
	docker run --rm -v ${PWD}:/go/src/github.com/fnproject/cli -w /go/src/github.com/fnproject/cli golang:alpine go build -o fn_alpine

.PHONY: install
