package client

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fnproject/cli/config"
	fnclient "github.com/fnproject/fn_go/client"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	oci "github.com/oracle/oci-go-sdk/common"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

func Host() string {
	hostURL := HostURL()
	return hostURL.Host
}

func HostURL() *url.URL {
	return hostURL(viper.GetString(config.EnvFnAPIURL))
}

func hostURL(urlStr string) *url.URL {
	if !strings.Contains(urlStr, "://") {
		urlStr = fmt.Sprint("http://", urlStr)
	}

	url, err := url.Parse(urlStr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unparsable FN API Url: %s. Error: %s \n", urlStr, err)
		os.Exit(1)
	}

	if url.Port() == "" {
		if url.Scheme == "http" {
			url.Host = fmt.Sprint(url.Host, ":80")
		}
		if url.Scheme == "https" {
			url.Host = fmt.Sprint(url.Host, ":443")
		}
	}

	//maintain backwards compatibility with first version FN_API_URL env vars
	if url.Path == "" || url.Path == "/" {
		url.Path = "/v1"
	}

	return url
}

func defaultProvider(transport *openapi.Runtime) {
	if token := viper.GetString(config.EnvFnToken); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}
}

func challengeForPKeyPassword() string {
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
	pKeyPword := viper.GetString(config.OraclePassPhrase)

	if !viper.IsSet(config.OraclePassPhrase) {
		if key = getPrivateKey(keyBytes, pKeyPword, pkeyFilePath); key == nil {
			pKeyPword = challengeForPKeyPassword()
		}
	}
	key = getPrivateKey(keyBytes, pKeyPword, pkeyFilePath)
	return key, nil
}

func getPrivateKey(keyBytes []byte, pKeyPword, pkeyFilePath string) *rsa.PrivateKey {
	key, err := oci.PrivateKeyFromBytes(keyBytes, oci.String(pKeyPword))
	if err != nil {
		if pKeyPword != "" {
			fmt.Fprintf(os.Stderr, "Unable to load private key from file bytes: %s. Error: %s \n", pkeyFilePath, err)
			os.Exit(1)
		}
	}

	return key
}

func oracleProvider(transport *openapi.Runtime) {
	keyID, pKey, err := oracleConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s \n", err)
		os.Exit(1)
	}
	compartmentID := viper.GetString(config.OracleCompartmentID)

	if viper.GetBool(config.OracleDisableCerts) {
		transport.Transport = InsecureRoundTripper(transport.Transport)
	}

	transport.Transport =
		NewCompartmentIDRoundTripper(
			compartmentID,
			NewOCISigningRoundTripper(
				keyID,
				pKey,
				transport.Transport))

}

func oracleConfigFile() (string, *rsa.PrivateKey, error) {
	var oracleProfile string
	if oracleProfile = viper.GetString(config.OracleProfile); oracleProfile == "" {
		oracleProfile = "DEFAULT"
	}

	cf, err := oci.ConfigurationProviderFromFileWithProfile(filepath.Join(config.GetHomeDir(), ".oci", "config"), oracleProfile, "")
	if err != nil {
		return "", nil, err
	}

	var tenancyID string
	if tenancyID = viper.GetString(config.OracleTenancyID); tenancyID == "" {
		tenancyID, err = cf.TenancyOCID()
		if err != nil {
			return "", nil, err
		}
	}

	var userID string
	if userID = viper.GetString(config.OracleUserID); userID == "" {
		userID, err = cf.UserOCID()
		if err != nil {
			return "", nil, err
		}
	}

	var fingerprint string
	if fingerprint = viper.GetString(config.OracleFingerprint); fingerprint == "" {
		fingerprint, err = cf.KeyFingerprint()
		if err != nil {
			return "", nil, err
		}
	}

	keyID := tenancyID + "/" + userID + "/" + fingerprint

	var pKey *rsa.PrivateKey
	if keyFile := viper.GetString(config.OracleKeyFile); keyFile != "" {
		pKey, err = privateKey(keyFile)
		if err != nil {
			return "", nil, err
		}
		return keyID, pKey, nil
	}

	// Read private key for .oci file
	pKey, err = cf.PrivateRSAKey()
	if err != nil {
		return "", nil, err
	}

	return keyID, pKey, nil
}

func GetTransportAndRegistry() (*openapi.Runtime, strfmt.Registry) {
	hostURL := HostURL()
	transport := openapi.New(hostURL.Host, hostURL.Path, []string{hostURL.Scheme})

	switch viper.GetString(config.ContextProvider) {
	case "default":
		defaultProvider(transport)
	case "oracle":
		oracleProvider(transport)
	default:
		defaultProvider(transport)
	}

	return transport, strfmt.Default
}

func APIClient() *fnclient.Fn {
	return fnclient.New(GetTransportAndRegistry())
}
