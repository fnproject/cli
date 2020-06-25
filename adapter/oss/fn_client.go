package oss

import (
	"github.com/fnproject/cli/adapter"
	oss "github.com/fnproject/fn_go/clientv2"
)

type FnClient struct {
	Client *oss.Fn
}

func (f FnClient) CreateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (f FnClient) GetFn(appID string, fnName string) (*adapter.Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (f FnClient) UpdateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	//TODO: call OSS client
	return nil, nil
}

func (f FnClient) DeleteFn(fnID string) error {
	//TODO: call OSS client
	return nil
}

func (f FnClient) ListFn(appID string, limit int64) error {
	//TODO: call OSS client
	return nil
}
