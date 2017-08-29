FROM golang:latest
RUN mkdir /app
ADd . /app
WORKDIR /app
RUN go build -o hello .
CMD ["/app/hello"]

