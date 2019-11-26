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
		return nil, fmt.Errorf("no OCI compartment ID specified in config key %s ", CfgCompartmentID)
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

//TODO support OCI environment variables
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
