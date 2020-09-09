package oci

import (
	"github.com/fnproject/cli/adapter"
	"github.com/oracle/oci-go-sdk/functions"
)

type Provider struct {
	FMClient *functions.FunctionsManagementClient
}

func (p Provider) APIClient() adapter.APIClient {
	return &APIClient{fnClient: &FnClient{client: p.FMClient}, appClient: &AppClient{client: p.FMClient}, triggerClient: &TriggerClient{}}
}