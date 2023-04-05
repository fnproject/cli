package oracle

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/oracle/oci-go-sdk/v65/functions"

	"github.com/fnproject/fn_go/provider"
	homedir "github.com/mitchellh/go-homedir"
	oci "github.com/oracle/oci-go-sdk/v65/common"
)

const (
	CfgUserID      = "oracle.user-id"
	CfgFingerprint = "oracle.fingerprint"
	CfgKeyFile     = "oracle.key-file"
	CfgPassPhrase  = "oracle.pass-phrase"
)

func NewFromConfig(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {
	configProvider, err := loadOracleConfig(configSource, passphraseSource)
	if err != nil {
		return nil, err
	}

	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		return nil, fmt.Errorf("no OCI compartment OCID specified in config key %s ", CfgCompartmentID)
	}

	ociClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	disableCerts := configSource.GetBool(CfgDisableCerts)
	if disableCerts {
		c := ociClient.HTTPClient.(*http.Client)
		c.Transport = InsecureRoundTripper(c.Transport)
	}

	ociClient.UserAgent = fmt.Sprintf("%s %s", userAgentPrefixUser, ociClient.UserAgent)

	// If we have an explicit api-url configured then use that, otherwise let OCI client compute the url from the standard
	// production url template and the configured region from environment.
	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	var apiUrl *url.URL
	if cfgApiUrl != "" {
		apiUrl, err = provider.CanonicalFnAPIUrl(cfgApiUrl)
		if err != nil {
			return nil, err
		}
		ociClient.Host = apiUrl.String()
	} else {
		// Even if URL is computed by OCI SDK itself, we still populate FnApiUrl in the Provider for compatibility's sake
		apiUrl, err = provider.CanonicalFnAPIUrl(ociClient.Host)
		if err != nil {
			return nil, err
		}
	}

	return &OracleProvider{
		FnApiUrl:              apiUrl,
		Signer:                oci.DefaultRequestSigner(configProvider),
		Interceptor:           nil,
		DisableCerts:          disableCerts,
		CompartmentID:         compartmentID,
		ImageCompartmentID:    configSource.GetString(CfgImageCompartmentID),
		ConfigurationProvider: configProvider,
		ociClient:             ociClient,
	}, nil
}

func loadOracleConfig(config provider.ConfigSource, passphraseSource provider.PassPhraseSource) (oci.ConfigurationProvider, error) {
	var oracleProfile string
	var err error
	var cf oci.ConfigurationProvider

	oracleProfile = getEnv(OCI_CLI_PROFILE_ENV_VAR, config.GetString(CfgProfile))

	if oracleProfile == "" {
		oracleProfile = "DEFAULT"
	}

	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory %s", err)
	}

	path := getEnv(OCI_CLI_CONFIG_FILE_ENV_VAR, filepath.Join(home, ".oci", "config"))

	if _, err := os.Stat(path); err == nil {
		cf, err = oci.ConfigurationProviderFromFileWithProfile(path, oracleProfile, "")
		if err != nil {
			return nil, err
		}
	}

	var tenancyID string
	if tenancyID = getEnv(OCI_CLI_TENANCY_ENV_VAR, config.GetString(CfgTenancyID)); tenancyID == "" {
		if cf == nil {
			return nil, errors.New("unable to find tenancyID in environment or configuration.")
		}
		tenancyID, err = cf.TenancyOCID()
		if err != nil {
			return nil, err
		}

	}

	var userID string
	if userID = getEnv(OCI_CLI_USER_ENV_VAR, config.GetString(CfgUserID)); userID == "" {
		if cf == nil {
			return nil, errors.New("unable to find userID in environment or configuration.")
		}
		userID, err = cf.UserOCID()
		if err != nil {
			return nil, err
		}
	}

	var fingerprint string
	if fingerprint = getEnv(OCI_CLI_FINGERPRINT_ENV_VAR, config.GetString(CfgFingerprint)); fingerprint == "" {
		if cf == nil {
			return nil, errors.New("unable to find fingerprint in environment or configuration.")
		}
		fingerprint, err = cf.KeyFingerprint()
		if err != nil {
			return nil, err
		}
	}

	var keyFile string
	var passphrase *string
	if keyFile = getEnv(OCI_CLI_KEY_FILE_ENV_VAR, config.GetString(CfgKeyFile)); keyFile != "" {
		isEncrypted, err := isPrivateKeyEncrypted(keyFile)
		if err != nil {
			return nil, err
		}

		if isEncrypted {
			passphrase, err = getPrivateKeyPassphrase(config, passphraseSource, keyFile)
			if err != nil {
				return nil, err
			}
		}
	}

	// We need to ensure that, either api-url has been set in the context, or region is set in the OCI config file (from which endpoint will be constructed)
	// 'region' here is what gets passed to the override config provider. The cases are:
	//      - api-url is not set, region is not set -> return error
	//      - api-url is set, region is not set -> pass a string other than a blank string in the overrideConfigProvider - this is necessary as OCI client construction
	//          enforces a region value being present despite the fact we then override endpoint
	//      - api-url is not set, region is set -> we pass a blank string to the overrideConfigProvider such that it is considered unset, and the region from the main cf takes precedence
	region := "."
	if !config.IsSet(provider.CfgFnAPIURL) {
		msg := "unable to find api-url in context, or region in OCI config"
		if cf == nil {
			return nil, errors.New(msg)
		}
		_, err = cf.Region()
		if err != nil {
			return nil, errors.New(msg)
		}
		region = ""
	}

	overrideConfigProvider := oci.NewRawConfigurationProvider(tenancyID, userID, region, fingerprint, keyFile, passphrase)

	// We use a composing configuration provider, so that values set by env vars or Fn context take precedence over OCI config file
	return oci.ComposingConfigurationProvider([]oci.ConfigurationProvider{
		overrideConfigProvider,
		cf,
	})
}

func isPrivateKeyEncrypted(pkeyFilePath string) (bool, error) {
	keyBytes, err := ioutil.ReadFile(pkeyFilePath)
	if err != nil {
		return false, fmt.Errorf("Unable to read private key from file due to error: %s\n", err)
	}

	pemBlock, _ := pem.Decode(keyBytes)
	if pemBlock != nil {
		return false, fmt.Errorf("unable to decode private key file as PEM")
	}

	return x509.IsEncryptedPEMBlock(pemBlock), nil
}

func getPrivateKeyPassphrase(config provider.ConfigSource, passphraseSource provider.PassPhraseSource, pkeyFilePath string) (*string, error) {
	if config.IsSet(CfgPassPhrase) {
		passphrase := config.GetString(CfgPassPhrase)
		return &passphrase, nil
	}

	passphrase, err := passphraseSource.ChallengeForPassPhrase("oracle.privateKey", fmt.Sprintf("Enter passphrase for private key %s", pkeyFilePath))
	return &passphrase, err
}
