package adapter

import (
	"github.com/oracle/oci-go-sdk/functions"
)

type OCIFnClient struct {
	client *functions.FunctionsManagementClient
}

func (a *OCIFnClient) CreateFn(fn *Fn) (*Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a *OCIFnClient) GetFn(appID string, fnName string) (*Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a *OCIFnClient) UpdateFn(fn *Fn) (*Fn, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a *OCIFnClient) DeleteFn(fnID string) error {
	//TODO: call OCI client
	return nil
}

func (a *OCIFnClient) ListFn(appID string, limit int64) error {
	//TODO: call OCI client
	return nil
}
