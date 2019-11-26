package oracle

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/oracle/oci-go-sdk/common/auth"

	"path"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/oracle/oci-go-sdk/common"
	oci "github.com/oracle/oci-go-sdk/common"
)

const (
	CfgTenancyID        = "oracle.tenancy-id"
	CfgUserID           = "oracle.user-id"
	CfgFingerprint      = "oracle.fingerprint"
	CfgKeyFile          = "oracle.key-file"
	CfgPassPhrase       = "oracle.pass-phrase"
	CfgProfile          = "oracle.profile"
	CfgCompartmentID    = "oracle.compartment-id"
	CfgDisableCerts     = "oracle.disable-certs"
	CompartmentMetadata = "http://169.254.169.254/opc/v1/instance/compartmentId"
)

// Provider : Oracle Authentication provider
type OracleProvider struct {
	// FnApiUrl is the endpoint to use for API interactions
	FnApiUrl *url.URL

	// The key provider can be a user or instance-principal-based one
	KP oci.KeyProvider

	//DisableCerts indicates if server certificates should be ignored
	DisableCerts bool

	//CompartmentID is the ocid of the functions compartment ID for a given function
	CompartmentID string
}

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

func NewFromConfig(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {

	apiUrl, err := provider.CanonicalFnAPIUrl(configSource.GetString(provider.CfgFnAPIURL))
	if err != nil {
		return nil, err
	}

	keyID, pKey, err := loadOracleConfig(configSource, passphraseSource)

	if err != nil {
		return nil, err
	}

	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		return nil, fmt.Errorf("no OCI compartment ID specified in config key %s ", CfgCompartmentID)
	}

	return &OracleProvider{
		FnApiUrl: apiUrl,
		KP: &ociKeyProvider{
			ID:  keyID,
			key: pKey,
		},
		DisableCerts:  configSource.GetBool(CfgDisableCerts),
		CompartmentID: compartmentID,
	}, nil
}

func NewIPProvider(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {
	ip, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}

	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	if cfgApiUrl == "" {
		region, err := ip.Region()
		if err != nil {
			return nil, err
		}
		// Construct the API endpoint from the "nearby" endpoint
		cfgApiUrl = fmt.Sprintf("https://functions.%s.oraclecloud.com", region)
	}
	apiUrl, err := provider.CanonicalFnAPIUrl(cfgApiUrl)
	if err != nil {
		return nil, err
	}

	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		// Get the local compartment ID from the metadata endpoint
		resp, err := http.DefaultClient.Get(CompartmentMetadata)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment Id from metadata endpoint %s", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment Id from metadata endpoint %s", err)
		}
		compartmentID = string(body)
	}
	return &OracleProvider{
		FnApiUrl:      apiUrl,
		KP:            ip,
		DisableCerts:  configSource.GetBool(CfgDisableCerts),
		CompartmentID: compartmentID,
	}, nil
}

//-- Provider interface impl ----------------------------------------------------------------------------------

// func (op *OracleProvider) APIClient() *clientv2.Fn {
// 	runtime := openapi.New(op.FnApiUrl.Host, path.Join(op.FnApiUrl.Path, clientv2.DefaultBasePath), []string{op.FnApiUrl.Scheme})
// 	runtime.Transport = op.WrapCallTransport(runtime.Transport)
// 	return clientv2.New(runtime, strfmt.Default)
// }

func (op *OracleProvider) APIClientv2() *clientv2.Fn {
	runtime := openapi.New(op.FnApiUrl.Host, path.Join(op.FnApiUrl.Path, clientv2.DefaultBasePath), []string{op.FnApiUrl.Scheme})
	runtime.Transport = op.WrapCallTransport(runtime.Transport)
	return clientv2.New(runtime, strfmt.Default)
}

func (op *OracleProvider) APIURL() *url.URL {
	return op.FnApiUrl
}

func (p *OracleProvider) UnavailableResources() []provider.FnResourceType {
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

	ociClient := common.RequestSigner(op.KP, []string{"host", "date", "(request-target)"}, []string{"content-length", "content-type", "x-content-sha256"})

	signingRoundTripper := ociSigningRoundTripper{
		transport: roundTripper,
		ociClient: ociClient,
	}

	roundTripper = compartmentIDRoundTripper{
		transport:     signingRoundTripper,
		compartmentID: op.CompartmentID,
	}

	roundTripper = requestIdRoundTripper{
		transport: roundTripper,
	}

	return roundTripper
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
	ociClient common.HTTPRequestSigner
	transport http.RoundTripper
}

func (t ociSigningRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
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

// http.RoundTripper middleware that injects an opc-request-id header
type requestIdRoundTripper struct {
	transport http.RoundTripper
}

func (t requestIdRoundTripper) RoundTrip(request *http.Request) (response *http.Response, e error) {
	requestID := provider.GetRequestID(request.Context())
	if requestID != "" {
		request.Header.Set("Opc-Request-Id", requestID)

	}
	response, e = t.transport.RoundTrip(request)
	return
}

//  http.RoundTripper middleware that adds an opc-compartment-id header to all requests
type compartmentIDRoundTripper struct {
	transport     http.RoundTripper
	compartmentID string
}

func (t compartmentIDRoundTripper) RoundTrip(request *http.Request) (response *http.Response, e error) {
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

func loadOracleConfig(config provider.ConfigSource, passphrase provider.PassPhraseSource) (string, *rsa.PrivateKey, error) {
	var oracleProfile string
	var err error
	var cf oci.ConfigurationProvider

	if oracleProfile = config.GetString(CfgProfile); oracleProfile == "" {
		oracleProfile = "DEFAULT"
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", nil, fmt.Errorf("error getting home directory %v", err)
	}

	path := filepath.Join(home, ".oci", "config")
	if _, err := os.Stat(path); err == nil {
		cf, err = oci.ConfigurationProviderFromFileWithProfile(path, oracleProfile, "")
		if err != nil {
			return "", nil, err
		}
	}

	var tenancyID string
	if tenancyID = config.GetString(CfgTenancyID); tenancyID == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find tenancyID in configuration: oracle.tenancy-id is missing from current-context file and tenancy-id wasn't found in ~/.oci/config.")
		}
		tenancyID, err = cf.TenancyOCID()
		if err != nil {
			return "", nil, err
		}

	}

	var userID string
	if userID = config.GetString(CfgUserID); userID == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find userID in configuration: oracle.tenancy-id is missing from current-context file and tenancy-id wasn't found in ~/.oci/config.")
		}
		userID, err = cf.UserOCID()
		if err != nil {
			return "", nil, err
		}
	}

	var fingerprint string
	if fingerprint = config.GetString(CfgFingerprint); fingerprint == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find fingerprint in configuration: oracle.tenancy-id is missing from current-context file and tenancy-id wasn't found in ~/.oci/config.")
		}
		fingerprint, err = cf.KeyFingerprint()
		if err != nil {
			return "", nil, err
		}
	}

	keyID := tenancyID + "/" + userID + "/" + fingerprint
	var pKey *rsa.PrivateKey

	if keyFile := config.GetString(CfgKeyFile); keyFile != "" {
		pKey, err = privateKey(config, passphrase, keyFile)
		if err != nil {
			return "", nil, err
		}
		return keyID, pKey, nil
	}

	if cf == nil {
		return "", nil, errors.New("unable to find private key in configuration: oracle.tenancy-id is missing from current-context file and tenancy-id wasn't found in ~/.oci/config.")
	}
	// Read private key for .oci file
	pKey, err = cf.PrivateRSAKey()
	if err != nil {
		return "", nil, err
	}

	return keyID, pKey, nil

}

func privateKey(config provider.ConfigSource, passphrase provider.PassPhraseSource, pkeyFilePath string) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(pkeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to load private key from file: %s. Error: %s \n", pkeyFilePath, err)
	}

	var key *rsa.PrivateKey
	pKeyPword := config.GetString(CfgPassPhrase)
	if !config.IsSet(CfgPassPhrase) {
		if key, err = getPrivateKey(keyBytes, pKeyPword, pkeyFilePath); key == nil {
			pKeyPword, err = passphrase.ChallengeForPassPhrase("oracle.privateKey", fmt.Sprintf("Enter passphrase for private key %s", pkeyFilePath))
			if err != nil {
				return nil, err
			}
		}
	}
	key, err = getPrivateKey(keyBytes, pKeyPword, pkeyFilePath)
	return key, err
}

func getPrivateKey(keyBytes []byte, pKeyPword, pkeyFilePath string) (*rsa.PrivateKey, error) {
	key, err := oci.PrivateKeyFromBytes(keyBytes, oci.String(pKeyPword))
	if err != nil {
		if pKeyPword != "" {
			return nil, fmt.Errorf("Unable to load private key from file bytes: %s. Error: %s \n", pkeyFilePath, err)
		}
	}

	return key, nil
}
