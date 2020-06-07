package adapter

import (
	oss "github.com/fnproject/fn_go/clientv2"
	"github.com/urfave/cli"
)

type ossFnClient struct{
	client	*oss.Fn
}

func (a *ossFnClient) create(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossFnClient) get(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossFnClient) update(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossFnClient) delete(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossFnClient) list(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}
