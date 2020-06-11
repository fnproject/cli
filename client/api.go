package client

import (
	"github.com/fnproject/cli/adapter"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
)

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
}

func CurrentProviderAdapter() (adapter.ProviderAdapter, error) {

	ctxProvider := viper.GetString(config.ContextProvider)
	currentProfile := viper.GetString("oracle.profile")

	//For default provider route the call to fn_go
	//Anything else, send it to oci go sdk
	switch ctxProvider {
	case "default":
		ossProvider, err := fn_go.DefaultProviders.ProviderFromConfig(ctxProvider, &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
		return &adapter.OSSProviderAdapter{OSSProvider: ossProvider}, err
	default:
		fnClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(common.CustomProfileConfigProvider("", currentProfile))
		return &adapter.OCIProviderAdapter{FMCClient: &fnClient}, err
	}
}
