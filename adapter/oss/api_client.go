package oss

import "github.com/fnproject/cli/adapter"

type APIClient struct {
	fnClient      *FnClient
	appClient     *AppClient
	triggerClient *TriggerClient
}

func (api *APIClient) FnClient() adapter.FnClient {
	return api.fnClient
}

func (api *APIClient) AppClient() adapter.AppClient {
	return api.appClient
}

func (api *APIClient) TriggerClient() adapter.TriggerClient {
	return api.triggerClient
}
