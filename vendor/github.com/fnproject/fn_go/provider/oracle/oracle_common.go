package oracle

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/common"
)

const (
	CfgTenancyID                          = "oracle.tenancy-id"
	CfgProfile                            = "oracle.profile"
	CfgCompartmentID                      = "oracle.compartment-id"
	CfgDisableCerts                       = "oracle.disable-certs"
	CompartmentMetadata                   = "http://169.254.169.254/opc/v1/instance/compartmentId"
	FunctionsAPIURLTmpl                   = "https://functions.%s.oraclecloud.com"
	requestHeaderOpcRequestID             = "Opc-Request-Id"
	requestHeaderOpcCompId                = "opc-compartment-id"
	OCI_CLI_PROFILE_ENV_VAR               = "OCI_CLI_PROFILE"
	OCI_CLI_REGION_ENV_VAR                = "OCI_CLI_REGION"
	OCI_CLI_TENANCY_ENV_VAR               = "OCI_CLI_TENANCY"
	OCI_CLI_CONFIG_FILE_ENV_VAR           = "OCI_CLI_CONFIG_FILE"
	OCI_CLI_DELEGATION_TOKEN_FILE_ENV_VAR = "OCI_CLI_DELEGATION_TOKEN_FILE"
)

type Response struct {
	Annotations Annotations `json:"annotations"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Name        string      `json:"name"`
}

type Annotations struct {
	CompartmentID string `json:"oracle.com/oci/compartmentId"`
	ShortCode     string `json:"oracle.com/oci/appCode"`
}

type OracleProvider struct {
	// FnApiUrl is the endpoint to use for API interactions
	FnApiUrl *url.URL

	//Signer performs auth operation
	Signer common.HTTPRequestSigner

	//A request interceptor can be used to customize the request before signing and dispatching
	Interceptor common.RequestInterceptor

	// DisableCerts indicates if server certificates should be ignored - TBD
	DisableCerts bool

	// CompartmentID is the ocid of the functions compartment ID for a given function
	CompartmentID string
}

//-- Provider interface impl ----------------------------------------------------------------------------------

type ociKeyProvider struct {
	ID  string
	key *rsa.PrivateKey
}

func (kp ociKeyProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return kp.key, nil
}

func (kp ociKeyProvider) KeyID() (string, error) {
	return kp.ID, nil
}

type ociSigningRoundTripper struct {
	signer        common.HTTPRequestSigner
	interceptor   common.RequestInterceptor
	transport     http.RoundTripper
	compartmentID string
}

func (t ociSigningRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	if request.Header.Get("Date") == "" {
		request.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	err = t.intercept(request)

	if err != nil {
		return
	}

	err = t.signRequest(request)

	if err != nil {
		return
	}

	response, err = t.transport.RoundTrip(request)

	return
}

// Add the necessary headers and sign the request
func (t ociSigningRoundTripper) signRequest(request *http.Request) (err error) {
	// Check that a Date header is set, otherwise authentication will fail
	if request.Header.Get("Date") == "" {
		return fmt.Errorf("Date header must be present and non-empty on request")
	}
	if request.Method == "POST" || request.Method == "PATCH" || request.Method == "PUT" {
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Length", fmt.Sprintf("%d", request.ContentLength))
	}

	err = t.signer.Sign(request)

	return err
}

// Add headers and call interceptor
func (t ociSigningRoundTripper) intercept(request *http.Request) (err error) {

	// set Opc-Request-Id header
	requestID := provider.GetRequestID(request.Context())
	if requestID != "" {
		request.Header.Set(requestHeaderOpcRequestID, requestID)

	}
	// set opc-compartment-id
	request.Header.Set(requestHeaderOpcCompId, t.compartmentID)

	// call interceptor
	if t.interceptor != nil {
		err = t.interceptor(request)
	}
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

//-- Provider interface impl ----------------------------------------------------------------------------------

func (op *OracleProvider) APIClientv2() *clientv2.Fn {
	runtime := openapi.New(op.FnApiUrl.Host, path.Join(op.FnApiUrl.Path, clientv2.DefaultBasePath),
		[]string{op.FnApiUrl.Scheme})
	runtime.Transport = op.WrapCallTransport(runtime.Transport)
	return clientv2.New(runtime, strfmt.Default)
}

func (op *OracleProvider) APIURL() *url.URL {
	return op.FnApiUrl
}

func (op *OracleProvider) UnavailableResources() []provider.FnResourceType {
	return []provider.FnResourceType{provider.TriggerResourceType}
}

func (op *OracleProvider) VersionClient() *version.Client {
	runtime := openapi.New(op.FnApiUrl.Host, op.FnApiUrl.Path, []string{op.FnApiUrl.Scheme})
	runtime.Transport = op.WrapCallTransport(runtime.Transport)
	return version.New(runtime, strfmt.Default)
}

func (op *OracleProvider) WrapCallTransport(roundTripper http.RoundTripper) http.RoundTripper {
	if op.DisableCerts {
		roundTripper = InsecureRoundTripper(roundTripper)
	}

	signingRoundTripper := ociSigningRoundTripper{
		transport:     roundTripper,
		signer:        op.Signer,
		interceptor:   op.Interceptor,
		compartmentID: op.CompartmentID,
	}

	return signingRoundTripper
}
