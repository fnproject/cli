package client

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type debugTransport struct {
	dest http.RoundTripper
}

func (dt *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))

	res, err := dt.dest.RoundTrip(req)
	if err != nil {
		fmt.Printf("HTTP error %s", err.Error())
		return res, err
	}
	b, err = httputil.DumpResponse(res, true)

	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return res, err
}

func WrapDebugTransport(tripper http.RoundTripper) http.RoundTripper {
	return &debugTransport{
		dest: tripper,
	}
}
