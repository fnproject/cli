package oracle

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fnproject/fn_go/provider/oracle/shim"
	"github.com/oracle/oci-go-sdk/v65/functions"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/v65/common"
)

const (
	CfgTenancyID                          = "oracle.tenancy-id"
	CfgProfile                            = "oracle.profile"
	CfgCompartmentID                      = "oracle.compartment-id"
	CfgImageCompartmentID                 = "oracle.image-compartment-id"
	CfgDisableCerts                       = "oracle.disable-certs"
	CompartmentMetadata                   = "http://169.254.169.254/opc/v2/instance/compartmentId"
	FunctionsAPIURLTmpl                   = "https://functions.%s.oci.%s"
	realmDomainMetadata                   = "http://169.254.169.254/opc/v2/instance/regionInfo/realmDomainComponent"
	requestHeaderOpcRequestID             = "Opc-Request-Id"
	requestHeaderOpcCompId                = "opc-compartment-id"
	requestHeaderOpcOboToken              = "opc-obo-token"
	OCI_CLI_PROFILE_ENV_VAR               = "OCI_CLI_PROFILE"
	OCI_CLI_REGION_ENV_VAR                = "OCI_CLI_REGION"
	OCI_CLI_TENANCY_ENV_VAR               = "OCI_CLI_TENANCY"
	OCI_CLI_CONFIG_FILE_ENV_VAR           = "OCI_CLI_CONFIG_FILE"
	OCI_CLI_DELEGATION_TOKEN_FILE_ENV_VAR = "OCI_CLI_DELEGATION_TOKEN_FILE"
	OCI_CLI_USER_ENV_VAR                  = "OCI_CLI_USER"
	OCI_CLI_FINGERPRINT_ENV_VAR           = "OCI_CLI_FINGERPRINT"
	OCI_CLI_KEY_FILE_ENV_VAR              = "OCI_CLI_KEY_FILE"

	userAgentPrefixUser = "fn_go-oracle"
	userAgentPrefixIp   = "fn_go-oracle-ip"
	userAgentPrefixCs   = "fn_go-oracle-cs"
)

type Response struct {
	Annotations Annotations `json:"annotations"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Name        string      `json:"name"`
	Shape       string      `json:shape`
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

	// ImageCompartmentID is the ocid of the functions compartment ID for a given function
	ImageCompartmentID string

	// ConfigurationProvider is the OCI configuration provider for signing requests
	ConfigurationProvider common.ConfigurationProvider

	ociClient functions.FunctionsManagementClient
}

//-- Provider interface impl ----------------------------------------------------------------------------------

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
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	if transport, ok := roundTripper.(*http.Transport); ok {
		if transport.TLSClientConfig != nil {
			transport.TLSClientConfig.InsecureSkipVerify = true
		} else {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		return transport
	}

	return nil
}

//-- Provider interface impl ----------------------------------------------------------------------------------

func (op *OracleProvider) APIClientv2() *clientv2.Fn {
	return &clientv2.Fn{
		Apps:     shim.NewAppsShim(op.ociClient, op.CompartmentID),
		Fns:      shim.NewFnsShim(op.ociClient),
		Triggers: shim.NewTriggersShim(),
	}
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

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}

// Retrieve second-level domain for the current realm from IMDS
func GetRealmDomain() (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", realmDomainMetadata, nil)
	if err != nil {
		return "", fmt.Errorf("problem fetching realm domain from metadata endpoint %s", err)
	}

	// IMDS v2 requires authorisation header for any request
	req.Header.Add("Authorization", "Bearer Oracle")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("problem fetching realm domain from metadata endpoint %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("problem reading body when fetching realm domain from metadata endpoint %s", err)
	}
	return string(body), nil
}
