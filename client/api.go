package client

import (
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/spf13/viper"
)

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
}
