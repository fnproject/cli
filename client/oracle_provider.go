package client

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/oracle/oci-go-sdk/common"
)

type OCIKeyProvider struct {
	ID  string
	key *rsa.PrivateKey
}

func (kp OCIKeyProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return kp.key, nil
}

func (kp OCIKeyProvider) KeyID() (string, error) {
	return kp.ID, nil
}

type OCISigningRoundTripper struct {
	ociClient common.HTTPRequestSigner
	transport http.RoundTripper
}

func NewOCISigningRoundTripper(keyID string, key *rsa.PrivateKey, transport http.RoundTripper) http.RoundTripper {
	ociClient := initializeClient(keyID, key)
	return OCISigningRoundTripper{
		transport: transport,
		ociClient: ociClient,
	}
}

func (t OCISigningRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	if request.Header.Get("Date") == "" {
		request.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	request, err = signRequest(t.ociClient, request)

	if err != nil {
		return
	}

	response, err = t.transport.RoundTrip(request)

	return
}

func initializeClient(keyID string, key *rsa.PrivateKey) common.HTTPRequestSigner {
	provider := OCIKeyProvider{
		ID:  keyID,
		key: key,
	}
	return common.RequestSigner(provider, []string{"date", "(request-target)"}, []string{"content-length", "content-type", "x-content-sha256"})
}

// Add the necessary headers and sign the request
func signRequest(signer common.HTTPRequestSigner, request *http.Request) (signedRequest *http.Request, err error) {
	// Check that a Date header is set, otherwise authentication will fail
	if request.Header.Get("Date") == "" {
		return nil, fmt.Errorf("Date header must be present and non-empty on request")
	}
	if request.Method == "POST" || request.Method == "PATCH" || request.Method == "PUT" {
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Length", fmt.Sprintf("%d", request.ContentLength))
	}

	err = signer.Sign(request)

	return request, err
}

//  http.RoundTripper middleware that adds an opc-compartment-id header to all requests
type CompartmentIDRoundTripper struct {
	transport     http.RoundTripper
	compartmentID string
}

func NewCompartmentIDRoundTripper(compartmentID string, transport http.RoundTripper) CompartmentIDRoundTripper {
	return CompartmentIDRoundTripper{
		transport:     transport,
		compartmentID: compartmentID,
	}
}

func (t CompartmentIDRoundTripper) RoundTrip(request *http.Request) (response *http.Response, e error) {
	request.Header.Set("opc-compartment-id", t.compartmentID)
	response, e = t.transport.RoundTrip(request)
	return
}

// Skip verification of insecure certs
func InsecureRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	transport := roundTripper.(*http.Transport)
	if transport != nil {
		if transport.TLSClientConfig != nil {
			transport.TLSClientConfig.InsecureSkipVerify = true
		} else {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	return transport
}
