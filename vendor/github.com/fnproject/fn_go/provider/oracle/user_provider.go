package oracle

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fnproject/fn_go/provider"
	homedir "github.com/mitchellh/go-homedir"
	oci "github.com/oracle/oci-go-sdk/common"
)

const (
	CfgUserID      = "oracle.user-id"
	CfgFingerprint = "oracle.fingerprint"
	CfgKeyFile     = "oracle.key-file"
	CfgPassPhrase  = "oracle.pass-phrase"
)

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
		return nil, fmt.Errorf("no OCI compartment OCID specified in config key %s ", CfgCompartmentID)
	}

	kp := &ociKeyProvider{
		ID:  keyID,
		key: pKey,
	}

	return &OracleProvider{
		FnApiUrl:      apiUrl,
		Signer:        oci.DefaultRequestSigner(kp),
		Interceptor:   nil,
		DisableCerts:  configSource.GetBool(CfgDisableCerts),
		CompartmentID: compartmentID,
	}, nil
}

func loadOracleConfig(config provider.ConfigSource, passphrase provider.PassPhraseSource) (string, *rsa.PrivateKey, error) {
	var oracleProfile string
	var err error
	var cf oci.ConfigurationProvider

	oracleProfile = getEnv(OCI_CLI_PROFILE_ENV_VAR, config.GetString(CfgProfile))

	if oracleProfile == "" {
		oracleProfile = "DEFAULT"
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", nil, fmt.Errorf("error getting home directory %s", err)
	}

	path := getEnv(OCI_CLI_CONFIG_FILE_ENV_VAR, filepath.Join(home, ".oci", "config"))

	if _, err := os.Stat(path); err == nil {
		cf, err = oci.ConfigurationProviderFromFileWithProfile(path, oracleProfile, "")
		if err != nil {
			return "", nil, err
		}
	}

	var tenancyID string
	if tenancyID = getEnv(OCI_CLI_TENANCY_ENV_VAR, config.GetString(CfgTenancyID)); tenancyID == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find tenancyID in environment or configuration.")
		}
		tenancyID, err = cf.TenancyOCID()
		if err != nil {
			return "", nil, err
		}

	}

	var userID string
	if userID = getEnv(OCI_CLI_USER_ENV_VAR, config.GetString(CfgUserID)); userID == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find userID in environment or configuration.")
		}
		userID, err = cf.UserOCID()
		if err != nil {
			return "", nil, err
		}
	}

	var fingerprint string
	if fingerprint = getEnv(OCI_CLI_FINGERPRINT_ENV_VAR, config.GetString(CfgFingerprint)); fingerprint == "" {
		if cf == nil {
			return "", nil, errors.New("unable to find fingerprint in environment or configuration.")
		}
		fingerprint, err = cf.KeyFingerprint()
		if err != nil {
			return "", nil, err
		}
	}

	keyID := tenancyID + "/" + userID + "/" + fingerprint
	var pKey *rsa.PrivateKey

	if keyFile := getEnv(OCI_CLI_USER_ENV_VAR, config.GetString(CfgKeyFile)); keyFile != "" {
		pKey, err = privateKey(config, passphrase, keyFile)
		if err != nil {
			return "", nil, err
		}
		return keyID, pKey, nil
	}

	if cf == nil {
		return "", nil, errors.New("unable to find private key in environment or configuration.")
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
		return nil, fmt.Errorf("Unable to read private key from file due to error: %s\n", err)
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
			return nil, fmt.Errorf("Unable to load private key from file due to error: %s\n", err)
		}
	}

	return key, nil
}
