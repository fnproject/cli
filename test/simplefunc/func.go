package main

import (
	"context"
	"github.com/fnproject/fdk-go"
	"io"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(func(_ context.Context, in io.Reader, out io.Writer) {
		out.Write([]byte("hello world"))
	}))
}
