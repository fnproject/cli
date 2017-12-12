package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client"
	apicall "github.com/fnproject/fn_go/client/call"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

type callsCmd struct {
	client *fnclient.Fn
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
				Usage:     "list all calls for the specific app. Route is optional",
				ArgsUsage: "<app>",
				Action:    c.list,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "path",
						Usage: "function's path",
					},
					cli.StringFlag{
						Name:  "cursor",
						Usage: "pagination cursor",
					},
					cli.StringFlag{
						Name:  "from-time",
						Usage: "'start' timestamp",
					},
					cli.StringFlag{
						Name:  "to-time",
						Usage: "'stop' timestamp",
					},
					cli.Int64Flag{
						Name:  "per-page",
						Usage: "number of calls to return",
						Value: int64(30),
					},
				},
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
		if call.Error != "" {
			fmt.Println(fmt.Sprintf("Error reason: %v\n", call.Error))
		}
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
	if ctx.String("cursor") != "" {
		cursor := ctx.String("cursor")
		params.Cursor = &cursor
	}
	if ctx.String("path") != "" {
		route := ctx.String("path")
		params.Path = &route
	}
	if ctx.String("from-time") != "" {
		fromTime := ctx.String("from-time")
		fromTime_int64, err := time.Parse(time.RFC3339, fromTime)
		if err != nil {
			return err
		}
		res := fromTime_int64.Unix()
		params.FromTime = &res

	} else {
		lastHour := time.Now().Add(-time.Hour).Unix()
		params.FromTime = &lastHour
	}

	if ctx.String("to-time") != "" {
		toTime := ctx.String("to-time")
		toTime_int64, err := time.Parse(time.RFC3339, toTime)
		if err != nil {
			return err
		}
		res := toTime_int64.Unix()
		params.ToTime = &res
	}
	if ctx.Int64("per-page") > 0 {
		per_page := ctx.Int64("per-page")
		params.PerPage = &per_page
	}

	for {
		resp, err := call.client.Call.GetAppsAppCalls(&params)
		if err != nil {
			switch e := err.(type) {
			case *apicall.GetAppsAppCallsNotFound:
				return errors.New(e.Payload.Error.Message)
			default:
				return err
			}
		}

		printCalls(resp.Payload.Calls)

		if resp.Payload.NextCursor == "" {
			break
		}
		params.Cursor = &resp.Payload.NextCursor
	}

	return nil
}
