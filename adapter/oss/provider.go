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
	return &APIClient{
		fnClient: &FnClient{client: v2Client},
		appClient: &AppClient{client: v2Client},
		triggerClient: &TriggerClient{client: v2Client},
	}
}