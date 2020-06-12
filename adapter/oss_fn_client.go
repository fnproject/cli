package adapter

import (
	oss "github.com/fnproject/fn_go/clientv2"
)

type OSSFnClient struct {
	Client *oss.Fn
}

func (a *OSSFnClient) CreateFn(fn *Fn) (*Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (a *OSSFnClient) GetFn(appID string, fnName string) (*Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (a *OSSFnClient) UpdateFn(fn *Fn) (*Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (a *OSSFnClient) DeleteFn(fnID string) error {
	//TODO: call OSS client
	return nil
}

func (a *OSSFnClient) ListFn(appID string, limit int64) error {
	//TODO: call OSS client
	return nil
}
