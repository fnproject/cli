package oci

import (
	"github.com/fnproject/cli/adapter"
	"github.com/oracle/oci-go-sdk/functions"
)

type Provider struct {
	FMCClient *functions.FunctionsManagementClient
}

func (p Provider) APIClient() adapter.APIClient {
	return &APIClient{fnClient: &FnClient{client: p.FMCClient}, appClient: &AppClient{client: p.FMCClient}, triggerClient: &TriggerClient{}}
}

func (p Provider) VersionClient() adapter.VersionClient {
	// TODO: implement
	return nil
}

func (p Provider) FunctionInvokeClient() adapter.FunctionInvokeClient {
	// TODO: implement
	return nil
}
