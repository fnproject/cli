package client

import (
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/fnproject/cli/adapter/oci"
	"github.com/fnproject/cli/adapter/oss"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/oracle/oci-go-sdk/v27/common"
	"github.com/oracle/oci-go-sdk/v27/common/auth"
	"github.com/oracle/oci-go-sdk/v27/functions"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
}

func CurrentProviderAdapter() (adapter.Provider, error) {
	ctxProvider := viper.GetString(config.ContextProvider)
	currentProfile := viper.GetString("oracle.profile")
	switch ctxProvider {
	case "oracle":
		fnClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(common.CustomProfileConfigProvider("", currentProfile))
		return &oci.Provider{FMClient: &fnClient}, err
	case "oracle-ip":
		common.EnableInstanceMetadataServiceLookup()
		provider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
		fnClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(provider)
		return &oci.Provider{FMClient: &fnClient}, err
	case "oracle-cs":
		common.EnableInstanceMetadataServiceLookup()
		provider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}

		var delegationToken string
		delegationTokenFile := os.Getenv(oracle.OCI_CLI_DELEGATION_TOKEN_FILE_ENV_VAR)
		if delegationTokenFile != "" {
			fileContent, err := ioutil.ReadFile(delegationTokenFile)
			if err != nil {
				return nil, fmt.Errorf("Could not load delegation token from file due to error: %s\n", err)
			}
			delegationToken = string(fileContent)
		}
		if delegationToken == "" {
			return nil, fmt.Errorf("Could not derive delegation token filepath from either config or environment.")
		}

		fnClient, err := functions.NewFunctionsManagementClientWithOboToken(provider, delegationToken)
		return &oci.Provider{FMClient: &fnClient}, err
	default:
		ossProvider, err := fn_go.DefaultProviders.ProviderFromConfig(ctxProvider, &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
		return &oss.Provider{OSSProvider: ossProvider}, err
	}
}
