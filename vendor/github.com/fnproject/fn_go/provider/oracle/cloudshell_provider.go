package oracle

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/oracle/oci-go-sdk/v48/functions"

	"github.com/fnproject/fn_go/provider"
	oci "github.com/oracle/oci-go-sdk/v48/common"
	"github.com/oracle/oci-go-sdk/v48/common/auth"
)

const (
	DelegationTokenFileLocation = "/etc/oci/delegation_token"
)

// Holds the three required config values in a CS env.
type CloudShellConfig struct {
	tenancyID       string
	region          string
	delegationToken string
}

// Creates a new "oracle-cs" provider instance for use when Fn is deployed in an OCI CloudShell environment.
func NewCSProvider(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {

	var csConfig *CloudShellConfig
	var err error

	// Derive oracle.profile from context or environment
	oraProfile := getEnv(OCI_CLI_PROFILE_ENV_VAR, configSource.GetString(CfgProfile))

	// If the oracle.profile in env or context isn't empty then derive config from profile in OCI config
	if oraProfile != "" {
		csConfig, err = loadCSOracleConfig(oraProfile, passphraseSource)
		if err != nil {
			return nil, err
		}
	} else {
		csConfig = &CloudShellConfig{tenancyID: "", region: "", delegationToken: ""}
	}

	// Now we read config from environment to either override base config, or instead of if config existed.
	csConfig.region = getEnv(OCI_CLI_REGION_ENV_VAR, csConfig.region)

	if csConfig.region == "" {
		return nil, fmt.Errorf("Could not derive region from either config or environment.")
	}

	csConfig.tenancyID = getEnv(OCI_CLI_TENANCY_ENV_VAR, csConfig.tenancyID)
	if csConfig.tenancyID == "" {
		return nil, fmt.Errorf("Could not derive tenancy ID from either config or environment.")
	}

	delegationTokenFile := os.Getenv(OCI_CLI_DELEGATION_TOKEN_FILE_ENV_VAR)
	if delegationTokenFile != "" {
		fileContent, err := ioutil.ReadFile(delegationTokenFile)
		if err != nil {
			return nil, fmt.Errorf("Could not load delegation token from file due to error: %s\n", err)
		}
		csConfig.delegationToken = string(fileContent)
	}
	if csConfig.delegationToken == "" {
		return nil, fmt.Errorf("Could not derive delegation token filepath from either config or environment.")
	}

	// If we have an explicit api-url configured then use that, otherwise compute the url from the standard
	// production url form and the configured region from environment.
	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	if cfgApiUrl == "" {
		domain, err := GetRealmDomain()
		if err != nil {
			return nil, err
		}
		cfgApiUrl = fmt.Sprintf(FunctionsAPIURLTmpl, csConfig.region, domain)
	}
	apiUrl, err := provider.CanonicalFnAPIUrl(cfgApiUrl)
	if err != nil {
		return nil, err
	}

	// If the compartment ID wasn't specified in the context, we default to the root compartment by using
	// the tenancy ID.
	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		compartmentID = csConfig.tenancyID
	}

	// Set OCI SDK to use IMDS to fetch region info (second-level domain etc.)
	oci.EnableInstanceMetadataServiceLookup()

	configProvider, err := auth.InstancePrincipalDelegationTokenConfigurationProvider(&csConfig.delegationToken)
	if err != nil {
		return nil, err
	}

	// Interceptor to add obo token header
	interceptor := func(request *http.Request) error {
		request.Header.Add(requestHeaderOpcOboToken, csConfig.delegationToken)
		return nil
	}

	// Obo token will also be signed
	defaultHeaders := append(oci.DefaultGenericHeaders(), requestHeaderOpcOboToken)
	signer := oci.RequestSigner(configProvider, defaultHeaders, oci.DefaultBodyHeaders())

	ociClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	ociClient.UserAgent = fmt.Sprintf("%s %s", userAgentPrefixCs, ociClient.UserAgent)

	disableCerts := configSource.GetBool(CfgDisableCerts)
	if disableCerts {
		c := ociClient.HTTPClient.(*http.Client)
		c.Transport = InsecureRoundTripper(c.Transport)
	}

	ociClient.Host = apiUrl.String()

	return &OracleProvider{
		FnApiUrl:              apiUrl,
		Signer:                signer,
		Interceptor:           interceptor,
		DisableCerts:          disableCerts,
		CompartmentID:         compartmentID,
		ImageCompartmentID:    configSource.GetString(CfgImageCompartmentID),
		ConfigurationProvider: configProvider,
		ociClient:             ociClient,
	}, nil
}

func GetOCIRegionTenancy() (region string, tenancy string, err error) {
	var csConfig *CloudShellConfig
	oraProfile := os.Getenv(OCI_CLI_PROFILE_ENV_VAR)
	if oraProfile != "" {
		csConfig, err = loadCSOracleConfig(oraProfile, &provider.TerminalPassPhraseSource{})
		if err != nil {
			return "", "", err
		}
		if csConfig.region == "" {
			csConfig.region = oraProfile
		}
	} else {
		csConfig = &CloudShellConfig{tenancyID: "", region: "", delegationToken: ""}
	}

	return csConfig.region, csConfig.tenancyID, nil
}

func loadCSOracleConfig(profileName string, passphrase provider.PassPhraseSource) (*CloudShellConfig, error) {
	var err error
	var cf oci.ConfigurationProvider

	path := os.Getenv(OCI_CLI_CONFIG_FILE_ENV_VAR)
	if _, err := os.Stat(path); err == nil {
		cf, err = oci.ConfigurationProviderFromFileWithProfile(path, profileName, "")
		if err != nil {
			return nil, err
		}
	}

	// if oci config file does not exist, proceed and use environment variables.
	if os.IsNotExist(err) {
		return &CloudShellConfig{tenancyID: "", region: "", delegationToken: ""}, nil
	}

	region, err := cf.Region()
	if err != nil {
		return nil, err
	}

	tenancyOCID, err := cf.TenancyOCID()
	if err != nil {
		return nil, err
	}

	fileContent, err := ioutil.ReadFile(DelegationTokenFileLocation)
	if err != nil {
		return nil, fmt.Errorf("can not load delegation_token from file due to error: %s \n", err)
	}

	delegationToken := string(fileContent)

	return &CloudShellConfig{tenancyID: tenancyOCID, region: region, delegationToken: delegationToken}, nil
}
