package oci

import (
	"github.com/fnproject/cli/adapter"
	"github.com/oracle/oci-go-sdk/functions"
)

type FnClient struct {
	client *functions.FunctionsManagementClient
}

func (f FnClient) CreateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (f FnClient) GetFn(appID string, fnName string) (*adapter.Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (f FnClient) UpdateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (f FnClient) DeleteFn(fnID string) error {
	//TODO: call OCI client
	return nil
}

func (f FnClient) ListFn(appID string, limit int64) ([]*adapter.Fn, error) {
	//TODO: call OCI client
	return nil, nil
}
