package main

import (
	"context"
	"errors"
	"fmt"

	client "github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client"
	apicall "github.com/fnproject/fn_go/client/call"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

type callsCmd struct {
	client *fnclient.Functions
}

func calls() cli.Command {
	c := callsCmd{client: client.APIClient()}

	return cli.Command{
		Name:  "calls",
		Usage: "manage function calls for apps",
		Subcommands: []cli.Command{
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "get function call info per app",
				ArgsUsage: "<app> <call-id>",
				Action:    c.get,
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list all calls for specific app / route route is optional",
				ArgsUsage: "<app> [route]",
				Action:    c.list,
			},
		},
	}
}

func printCalls(calls []*models.Call) {
	for _, call := range calls {
		fmt.Println(fmt.Sprintf(
			"ID: %v\n"+
				"App: %v\n"+
				"Route: %v\n"+
				"Created At: %v\n"+
				"Started At: %v\n"+
				"Completed At: %v\n"+
				"Status: %v\n",
			call.ID, call.AppName, call.Path, call.CreatedAt,
			call.StartedAt, call.CompletedAt, call.Status))
	}
}

func (call *callsCmd) get(ctx *cli.Context) error {
	app, callID := ctx.Args().Get(0), ctx.Args().Get(1)
	params := apicall.GetAppsAppCallsCallParams{
		Call:    callID,
		App:     app,
		Context: context.Background(),
	}
	resp, err := call.client.Call.GetAppsAppCallsCall(&params)
	if err != nil {
		switch e := err.(type) {
		case *apicall.GetAppsAppCallsCallNotFound:
			return errors.New(e.Payload.Error.Message)
		default:
			return err
		}
	}
	printCalls([]*models.Call{resp.Payload.Call})
	return nil
}

func (call *callsCmd) list(ctx *cli.Context) error {
	app := ctx.Args().Get(0)
	params := apicall.GetAppsAppCallsParams{
		App:     app,
		Context: context.Background(),
	}
	if ctx.Args().Get(1) != "" {
		route := ctx.Args().Get(1)
		params.Route = &route
	}
	resp, err := call.client.Call.GetAppsAppCalls(&params)
	if err != nil {
		switch e := err.(type) {
		case *apicall.GetCallsCallNotFound:
			return errors.New(e.Payload.Error.Message)
		default:
			return err
		}
	}
	printCalls(resp.Payload.Calls)
	return nil
}
