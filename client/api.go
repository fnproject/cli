package client

import (
	"github.com/fnproject/cli/config"
	"github.com/spf13/viper"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
)

type viperConfigSource struct {
}

func (*viperConfigSource) GetString(key string) string {
	return viper.GetString(key)
}

func (*viperConfigSource) GetBool(key string) bool {
	return viper.GetBool(key)
}
func (*viperConfigSource) IsSet(key string) bool {
	return viper.IsSet(key)
}

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &viperConfigSource{}, &provider.TerminalPassPhraseSource{})
}
