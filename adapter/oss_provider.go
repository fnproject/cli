package adapter

import "github.com/fnproject/fn_go/provider"

type OSSProviderAdapter struct {
	OSSProvider provider.Provider
}

func (o OSSProviderAdapter) APIClientAdapter() APIClientAdapter {

	v2Client := o.OSSProvider.APIClientv2()

	return &OSSClient{ossFn: &OSSFnClient{Client: v2Client}, ossApp: &OSSAppClient{client: v2Client},}
}

func (o OSSProviderAdapter) VersionClientAdapter() VersionClientAdapter {
	// TODO: implement
	return nil
}

func (o OSSProviderAdapter) FunctionInvokeClientAdapter() FunctionInvokeClientAdapter {
	// TODO: implement
	return nil
}