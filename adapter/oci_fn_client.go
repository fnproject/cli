package adapter

import (
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/urfave/cli"
)

type OCIFnClient struct {
	client *functions.FunctionsManagementClient
}

func (a *OCIFnClient) CreateFn(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *OCIFnClient) GetFn(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *OCIFnClient) UpdateFn(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *OCIFnClient) DeleteFn(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}

func (a *OCIFnClient) ListFn(c *cli.Context) error {
	//TODO: call OCI client
	return nil
}
