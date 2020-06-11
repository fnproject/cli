package adapter

import "github.com/oracle/oci-go-sdk/functions"

type OCIProviderAdapter struct {
	FMCClient *functions.FunctionsManagementClient
}

func (o *OCIProviderAdapter) GetClientAdapter() ClientAdapter {

	return &OCIClient{ociFn: &OCIFnClient{client: o.FMCClient}, ociApp: &OCIAppClient{client: o.FMCClient},}
}
