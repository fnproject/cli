package main

import (
	"context"
	"io"
	"github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(func(_ context.Context, in io.Reader, out io.Writer) {
		out.Write([]byte("hello world"))
	}))
}
