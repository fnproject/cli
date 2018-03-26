package main

import (
	"context"
	"io"
	"time"

	"encoding/json"
	"github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

func myHandler(_ context.Context, _ io.Reader, out io.Writer) {
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		fdk.WriteStatus(out, 500)
		out.Write([]byte(err.Error()))
		return
	}
	fdk.WriteStatus(out, 200)
	json.NewEncoder(out).Encode(tz.String())
}
