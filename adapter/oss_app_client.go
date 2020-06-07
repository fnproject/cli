package adapter

import (
	oss "github.com/fnproject/fn_go/clientv2"
	"github.com/urfave/cli"
)

type ossAppClient struct{
	client	*oss.Fn
}

func (a *ossAppClient) create(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossAppClient) get(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossAppClient) update(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossAppClient) delete(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *ossAppClient) list(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}
