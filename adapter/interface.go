package adapter

import "github.com/urfave/cli"

type Client interface {
	getFnClient() 		FnClient
	getAppsClient() 	AppClient
	getTriggerClient() 	TriggerClient
}

type FnClient interface {
	create(c *cli.Context) error
	update(c *cli.Context) error
	list(c *cli.Context) error
	delete(c *cli.Context) error
}

type AppClient interface {
	create(c *cli.Context) error
	get(c *cli.Context) error
	update(c *cli.Context) error
	list(c *cli.Context) error
	delete(c *cli.Context) error
}

type TriggerClient interface {

}

//TODO: InvokeFunction Client

//TODO: Version Client