package commands

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go/provider/oracle"
	ociCommon "github.com/oracle/oci-go-sdk/v48/common"
	"github.com/spf13/viper"
	"net/url"
	"testing"
)

type mockConfigurationProvider struct {
	region string
}

func (mcp *mockConfigurationProvider) TenancyOCID() (string, error)            { return "", nil }
func (mcp *mockConfigurationProvider) UserOCID() (string, error)               { return "", nil }
func (mcp *mockConfigurationProvider) KeyID() (string, error)                  { return "", nil }
func (mcp *mockConfigurationProvider) KeyFingerprint() (string, error)         { return "", nil }
func (mcp *mockConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) { return nil, nil }
func (mcp *mockConfigurationProvider) AuthType() (ociCommon.AuthConfig, error) {
	return ociCommon.AuthConfig{}, nil
}
func (mcp *mockConfigurationProvider) Region() (string, error) { return mcp.region, nil }

var _ ociCommon.ConfigurationProvider = &mockConfigurationProvider{}

func TestIsSignatureConfigured(t *testing.T) {
	// signature configuration is valid if all 4 values are set
	signingDetails := common.SigningDetails{
		ImageCompartmentId: "ocid1.compartment.test",
		KmsKeyId:           "ocid1.kmskey.test",
		KmsKeyVersionId:    "ocid1.kmskeyversion.test",
		SigningAlgorithm:   "SHA_256_RSA_PKCS_PSS",
	}
	configured, err := isSignatureConfigured(signingDetails)
	if !configured {
		t.Fatal("expected true")
	}
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	// signature configuration is invalid if partial values are set
	signingDetails.ImageCompartmentId = ""
	configured, err = isSignatureConfigured(signingDetails)
	if err == nil {
		t.Fatal("expected error")
	}
	// signature configuration is valid if no values are set
	signingDetails = common.SigningDetails{}
	configured, err = isSignatureConfigured(signingDetails)
	if configured {
		t.Fatal("expected false")
	}
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
}

func TestGetRegion(t *testing.T) {
	// when FnApiUrl is set, retrieve region from the URL
	oracleProvider := &oracle.OracleProvider{
		FnApiUrl: &url.URL{Host: "functions.us-ashburn-1.oci.oraclecloud.com"},
	}
	region := getRegion(oracleProvider)
	expected := "us-ashburn-1"
	if region != expected {
		t.Fatalf("expected %s, but found %s", expected, region)
	}
	// when FnApiUrl is not set or cannot be parsed, retrieve region from OCI configuration provider
	oracleProvider.FnApiUrl.Host = "functions.com"
	oracleProvider.ConfigurationProvider = &mockConfigurationProvider{"us-phoenix-1"}
	region = getRegion(oracleProvider)
	expected = "us-phoenix-1"
	if region != expected {
		t.Fatalf("expected %s, but found %s", expected, region)
	}
}

func TestGetRepositoryName(t *testing.T) {
	viper.Set(config.EnvFnRegistry, "iad.ocir.io/test")
	ff := &common.FuncFileV20180708{
		Version: "1.0.0",
		Name:    "testfn",
	}
	repositoryName, err := getRepositoryName(ff)
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	expected := "testfn"
	if repositoryName != expected {
		t.Fatalf("expected %s, but found %s", expected, repositoryName)
	}
	viper.Set(config.EnvFnRegistry, "iad.ocir.io/test/test2")
	repositoryName, err = getRepositoryName(ff)
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	expected = "test2/testfn"
	if repositoryName != expected {
		t.Fatalf("expected %s, but found %s", expected, repositoryName)
	}
}

func TestCreateImageSignatureMessage(t *testing.T) {
	region, imageDigest, repositoryName := "us-ashburn-1", "sha256:digestvalue", "test/reponame"
	signingDetails := common.SigningDetails{
		ImageCompartmentId: "ocid1.compartment.test",
		KmsKeyId:           "ocid1.key.test",
		KmsKeyVersionId:    "ocid1.keyversion.test",
		SigningAlgorithm:   "SHA_256_RSA_PKCS_PSS",
	}
	message, err := createImageSignatureMessage(region, imageDigest, repositoryName, signingDetails)
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	messageBytes, _ := json.Marshal(&Message{
		Description:      "image signed by fn CLI",
		ImageDigest:      imageDigest,
		KmsKeyId:         signingDetails.KmsKeyId,
		KmsKeyVersionId:  signingDetails.KmsKeyVersionId,
		Metadata:         "{\"signedBy\":\"fn CLI\"}",
		Region:           region,
		RepositoryName:   repositoryName,
		SigningAlgorithm: signingDetails.SigningAlgorithm,
	})
	expectedMessage := base64.StdEncoding.EncodeToString(messageBytes)
	if message != expectedMessage {
		t.Fatalf("expected %s, but found %s", expectedMessage, message)
	}
}

func TestBuildCryptoEndpoint(t *testing.T) {
	// test old style regional endpoints
	region, keyId := "us-ashburn-1", "ocid1.key.oc1.iad.testvault.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefgh"
	endpoint, err := buildCryptoEndpoint(region, keyId)
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	expected := "https://testvault-crypto.kms.us-ashburn-1.oraclecloud.com"
	if endpoint != expected {
		t.Fatalf("expected %s, found %s", expected, endpoint)
	}
	// test new style regional endpoints
	region, keyId = "us-sanjose-1", "ocid1.key.oc1.us-sanjose-1.testvault.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefgh"
	endpoint, err = buildCryptoEndpoint(region, keyId)
	if err != nil {
		t.Fatalf("expected no error, but found %s", err)
	}
	expected = "https://testvault-crypto.kms.us-sanjose-1.oci.oraclecloud.com"
	if endpoint != expected {
		t.Fatalf("expected %s, found %s", expected, endpoint)
	}
	// invalid key results in error
	keyId = "invalidocid"
	endpoint, err = buildCryptoEndpoint(region, keyId)
	if err == nil {
		t.Fatalf("expected error, but found %s", endpoint)
	}
}

func TestFindMissingValues(t *testing.T) {
	tests := map[string]common.SigningDetails{
		"kms_key_id,kms_key_version_id,signing_algorithm": {ImageCompartmentId: "test"},
		"kms_key_version_id,signing_algorithm":            {ImageCompartmentId: "test", KmsKeyId: "test"},
		"signing_algorithm":                               {ImageCompartmentId: "test", KmsKeyId: "test", KmsKeyVersionId: "test"},
	}
	for expected, signingDetails := range tests {
		actual := findMissingValues(signingDetails)
		if actual != expected {
			t.Fatalf("input: %+v, expected: %s, actual: %s", signingDetails, expected, actual)
		}
	}
}
