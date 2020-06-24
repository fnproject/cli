package adapter

import "github.com/oracle/oci-go-sdk/functions"

type OCIProviderAdapter struct {
	FMCClient *functions.FunctionsManagementClient
}

func (o OCIProviderAdapter) APIClientAdapter() APIClientAdapter {

	return &OCIClient{ociFn: &OCIFnClient{client: o.FMCClient}, ociApp: &OCIAppClient{client: o.FMCClient},}
}

func (o OCIProviderAdapter) VersionClientAdapter() VersionClientAdapter {
	// TODO: implement
	return nil
}

func (o OCIProviderAdapter) FunctionInvokeClientAdapter() FunctionInvokeClientAdapter {
	// TODO: implement
	return nil
}