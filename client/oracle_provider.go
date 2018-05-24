package client

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fnproject/cli/config"
	"github.com/oracle/oci-go-sdk/common"
	oci "github.com/oracle/oci-go-sdk/common"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	oracleTenancyID     = "oracle.tenancy-id"
	oracleUserID        = "oracle.user-id"
	oracleFingerprint   = "oracle.fingerprint"
	oracleKeyFile       = "oracle.key-file"
	oraclePassPhrase    = "oracle.pass-phrase"
	oracleProfile       = "oracle.profile"
	oracleCompartmentID = "oracle.compartment-id"
	oracleDisableCerts  = "oracle.disable-certs"
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
	return common.RequestSigner(provider, []string{"host", "date", "(request-target)"}, []string{"content-length", "content-type", "x-content-sha256"})
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

func OracleConfigFile() (string, *rsa.PrivateKey, error) {
	var oracleProfile string
	var err error
	var cf oci.ConfigurationProvider

	if oracleProfile = viper.GetString(oracleProfile); oracleProfile == "" {
		oracleProfile = "DEFAULT"
	}

	path := filepath.Join(config.GetHomeDir(), ".oci", "config")
	if _, err := os.Stat(path); err == nil {
		cf, err = oci.ConfigurationProviderFromFileWithProfile(path, oracleProfile, "")
		if err != nil {
			return "", nil, err
		}
	}

	var tenancyID string
	if tenancyID = viper.GetString(oracleTenancyID); tenancyID == "" {
		if cf == nil {
			return "", nil, errors.New("oracle.tenancy-id is missing from current-context file")
		}
		tenancyID, err = cf.TenancyOCID()
		if err != nil {
			return "", nil, err
		}

	}

	var userID string
	if userID = viper.GetString(oracleUserID); userID == "" {
		if cf == nil {
			return "", nil, errors.New("oracle.user-id is missing from current-context file")
		}
		userID, err = cf.UserOCID()
		if err != nil {
			return "", nil, err
		}
	}

	var fingerprint string
	if fingerprint = viper.GetString(oracleFingerprint); fingerprint == "" {
		if cf == nil {
			return "", nil, errors.New("oracle.fingerprint is missing from current-context file")
		}
		fingerprint, err = cf.KeyFingerprint()
		if err != nil {
			return "", nil, err
		}
	}

	keyID := tenancyID + "/" + userID + "/" + fingerprint
	var pKey *rsa.PrivateKey
	if keyFile := viper.GetString(oracleKeyFile); keyFile != "" {
		pKey, err = privateKey(keyFile)
		if err != nil {
			return "", nil, err
		}
		return keyID, pKey, nil
	}

	if cf == nil {
		return "", nil, errors.New("oracle.key-file is missing from current-context file")
	}
	// Read private key for .oci file
	pKey, err = cf.PrivateRSAKey()
	if err != nil {
		return "", nil, err
	}

	return keyID, pKey, nil

}

func challengeForPKeyPassword() string {
	fmt.Print("Private Key Phrase: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
	password := string(bytePassword)
	fmt.Println()

	return password
}

func privateKey(pkeyFilePath string) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(pkeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to load private key from file: %s. Error: %s \n", pkeyFilePath, err)
	}

	var key *rsa.PrivateKey
	pKeyPword := viper.GetString(oraclePassPhrase)

	if !viper.IsSet(oraclePassPhrase) {
		if key, err = getPrivateKey(keyBytes, pKeyPword, pkeyFilePath); key == nil {
			pKeyPword = challengeForPKeyPassword()
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
