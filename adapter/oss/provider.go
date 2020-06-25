package oss

import (
	"github.com/fnproject/cli/adapter"
	"github.com/fnproject/fn_go/provider"
)

type Provider struct {
	OSSProvider provider.Provider
}

func (p Provider) APIClient() adapter.APIClient {
	v2Client := p.OSSProvider.APIClientv2()
	return &APIClient{fnClient: &FnClient{Client: v2Client}, appClient: &AppClient{client: v2Client},}
}

func (p Provider) VersionClient() adapter.VersionClient {
	// TODO: implement
	return nil
}

func (p Provider) FunctionInvokeClient() adapter.FunctionInvokeClient {
	// TODO: implement
	return nil
}
