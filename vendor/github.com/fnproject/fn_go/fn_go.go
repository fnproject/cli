package fn_go

import (
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/defaultprovider"
	"github.com/fnproject/fn_go/provider/oracle"
)

const (
	// provider names
	DefaultProvider  = "default"
	OracleProvider   = "oracle"
	OracleIPProvider = "oracle-ip"
	OracleCSProvider = "oracle-cs"
)

// DefaultProviders includes the bundled providers available in the client
var DefaultProviders = provider.Providers{
	Providers: map[string]provider.ProviderFunc{
		"":               defaultprovider.NewFromConfig,
		DefaultProvider:  defaultprovider.NewFromConfig,
		OracleProvider:   oracle.NewFromConfig,
		OracleIPProvider: oracle.NewIPProvider,
		OracleCSProvider: oracle.NewCSProvider,
	},
}
