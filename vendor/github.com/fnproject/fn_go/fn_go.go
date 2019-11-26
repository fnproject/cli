package fn_go

import (
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/defaultprovider"
	"github.com/fnproject/fn_go/provider/oracle"
)

// DefaultProviders includes the bundled providers available in the client
var DefaultProviders = provider.Providers{
	Providers: map[string]provider.ProviderFunc{
		"":          defaultprovider.NewFromConfig,
		"default":   defaultprovider.NewFromConfig,
		"oracle":    oracle.NewFromConfig,
		"oracle-ip": oracle.NewIPProvider,
		"oracle-cs": oracle.NewCSProvider,
	},
}
