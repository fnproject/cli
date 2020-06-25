package client

import (
	"github.com/fnproject/cli/adapter"
	"github.com/fnproject/cli/adapter/oci"
	"github.com/fnproject/cli/adapter/oss"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
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
		return &oci.Provider{FMCClient: &fnClient}, err
	case "oracle-ip":
		provider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
		fnClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(provider)
		return &oci.Provider{FMCClient: &fnClient}, err
	case "oracle-cs":
		provider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
		// TODO: The OBO token may be obtained by following the example code at:
		// https://github.com/fnproject/fn_go/blob/master/provider/oracle/cloudshell_provider.go#L56
		// https://github.com/fnproject/fn_go/blob/master/provider/oracle/cloudshell_provider.go#L155
		fnClient, err := functions.NewFunctionsManagementClientWithOboToken(provider, "Fill me in")
		return &oci.Provider{FMCClient: &fnClient}, err
	default:
		ossProvider, err := fn_go.DefaultProviders.ProviderFromConfig(ctxProvider, &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
		return &oss.Provider{OSSProvider: ossProvider}, err
	}
}
