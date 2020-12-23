package client

import (
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnxproject/cli/config"
	"github.com/spf13/viper"
)

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
}
