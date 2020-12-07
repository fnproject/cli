# build stage
FROM golang:1.14.13-alpine3.12 AS build-env
RUN apk add --no-cache gcc musl-dev
ARG D=/go/src/github.com/fnproject/cli
ARG GO111MODULE=on
ARG GOFLAGS=-mod=vendor
ADD . $D
RUN cd $D && go build -o fn-alpine && cp fn-alpine /tmp/

# final stage
FROM alpine:3.12
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=build-env /tmp/fn-alpine /app/fn
ENTRYPOINT ["./fn"]
