package oracle

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fnproject/fn_go/provider"
	oci "github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
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
	oraProfile := configSource.GetString(CfgProfile)
	envOraProfile := os.Getenv(OCI_CLI_PROFILE_ENV_VAR)
	if envOraProfile != "" {
		oraProfile = envOraProfile
	}
	// If the oracle.profile in env or context isn't DEFAULT then derive config from OCI profile
	if oraProfile != "" {
		csConfig, err = loadCSOracleConfig(oraProfile, passphraseSource)
		if err != nil {
			return nil, err
		}
	} else {
		csConfig = &CloudShellConfig{tenancyID: "", region: "", delegationToken: ""}
	}

	// Now we read config from environment to either override base config, or instead of if config existed.
	region := os.Getenv(OCI_CLI_REGION_ENV_VAR)
	if region != "" {
		csConfig.region = region
	}
	if csConfig.region == "" {
		return nil, fmt.Errorf("Could not derive region from eiher config or environment.")
	}

	tenancyID := os.Getenv(OCI_CLI_TENANCY_ENV_VAR)
	if tenancyID != "" {
		csConfig.tenancyID = tenancyID
	}
	if csConfig.tenancyID == "" {
		return nil, fmt.Errorf("Could not derive tenancy ID from eiher config or environment.")
	}

	delegationTokenFile := os.Getenv(OCI_CLI_DELEGATION_TOKEN_FILE_ENV_VAR)
	if delegationTokenFile != "" {
		fileContent, err := ioutil.ReadFile(delegationTokenFile)
		if err != nil {
			return nil, fmt.Errorf("Could not load delegation token from file: %s. Error: %s \n", delegationTokenFile, err)
		}
		csConfig.delegationToken = string(fileContent)
	}
	if csConfig.delegationToken == "" {
		return nil, fmt.Errorf("Could not derive delegation token filepath from eiher config or environment.")
	}

	// If we have an explicit api-url configured then use that, otherwise compute the url from the standard
	// production url form and the configured region from environment.
	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	if cfgApiUrl == "" {
		cfgApiUrl = fmt.Sprintf(FunctionsAPIURLTmpl, csConfig.region)
	}
	apiUrl, err := provider.CanonicalFnAPIUrl(cfgApiUrl)
	if err != nil {
		return nil, err
	}
	//os.Stdout.WriteString("apiUrl:" + apiUrl.String())

	// If the compartment ID wasn't specified in the context, we default to the root compartment by using
	// the tenancy ID.
	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		compartmentID = csConfig.tenancyID
	}

	provider, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}

	client, err := oci.NewClientWithOboToken(provider, csConfig.delegationToken)
	if err != nil {
		return nil, err
	}

	return &OracleProvider{
		FnApiUrl:      apiUrl,
		Signer:        client.Signer,
		Interceptor:   client.Interceptor,
		DisableCerts:  configSource.GetBool(CfgDisableCerts),
		CompartmentID: compartmentID,
	}, nil
}

func loadCSOracleConfig(profileName string, passphrase provider.PassPhraseSource) (*CloudShellConfig, error) {
	var err error
	var cf oci.CloudShellConfigurationProvider

	path := os.Getenv(OCI_CLI_CONFIG_FILE_ENV_VAR)
	if _, err := os.Stat(path); err == nil {
		cf, err = oci.CloudshellConfigurationProviderFromFileWithProfile(path, profileName)
		if err != nil {
			return nil, err
		}
	}

	region, err := cf.Region()
	if err != nil {
		return nil, err
	}

	tenancyOCID, err := cf.TenancyOCID()
	if err != nil {
		return nil, err
	}

	delegationToken, err := cf.DelegationToken()
	if err != nil {
		return nil, err
	}

	return &CloudShellConfig{tenancyID: tenancyOCID, region: region, delegationToken: delegationToken}, nil
}
