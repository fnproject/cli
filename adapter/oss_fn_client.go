package adapter

import (
	oss "github.com/fnproject/fn_go/clientv2"
	"github.com/urfave/cli"
)

type OSSFnClient struct {
	Client *oss.Fn
}

func (a *OSSFnClient) CreateFn(c *cli.Context) error {
	//TODO: call OSS client
	// a.client.Fns.CreateFn()
	return nil
}

func (a *OSSFnClient) GetFn(c *cli.Context) error {
	//TODO: call OSS client
	return nil
}

func (a *OSSFnClient) UpdateFn(c *cli.Context) error {
	//TODO: call OSS client
	return nil
}

func (a *OSSFnClient) DeleteFn(c *cli.Context) error {
	//TODO: call OSS client
	return nil
}

func (a *OSSFnClient) ListFn(c *cli.Context) error {
	//TODO: call OSS client
	return nil
}
