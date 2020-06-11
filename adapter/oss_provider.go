package adapter

import "github.com/fnproject/fn_go/provider"

type OSSProviderAdapter struct {
	context     string
	OSSProvider provider.Provider
}

func (o *OSSProviderAdapter) GetClientAdapter() ClientAdapter {

	v2Client := o.OSSProvider.APIClientv2()

	return &OSSClient{ossFn: &OSSFnClient{Client: v2Client}, ossApp: &OSSAppClient{client: v2Client},}
}
