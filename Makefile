all: vendor build	
	./fn

build: 
	go build -o fn

docker: 
	docker build -t fnproject/fn:latest .

dep:
	glide install -v

dep-up:
	glide up -v

test:
	./test.sh

release:
	GOOS=linux go build -o fn_linux
	GOOS=darwin go build -o fn_mac
	GOOS=windows go build -o fn.exe
	docker run --rm -v ${PWD}:/go/src/github.com/fnproject/cli -w /go/src/github.com/fnproject/cli funcy/go:dev go build -o fn_alpine
