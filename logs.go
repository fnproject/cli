package main

import (
	"context"
	"errors"
	"fmt"

	client "github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client"
	ccall "github.com/fnproject/fn_go/client/call"
	apicall "github.com/fnproject/fn_go/client/operations"
	"github.com/urfave/cli"
)

type logsCmd struct {
	client *fnclient.Fn
}

func logs() cli.Command {
	c := logsCmd{}

	return cli.Command{
		Name:  "logs",
		Usage: "manage function calls for apps",
		Before: func(cxt *cli.Context) error {
			c.client = client.APIClient()
			return nil
		},
		Subcommands: []cli.Command{
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "get logs for a call. Must provide call_id or last to retrieve the most recent calls logs.",
				ArgsUsage: "<app> <call-id>",
				Action:    c.get,
			},
		},
	}
}

func (log *logsCmd) get(ctx *cli.Context) error {
	app, callID := ctx.Args().Get(0), ctx.Args().Get(1)
	if callID == "last" {
		params := ccall.GetAppsAppCallsParams{
			App:     app,
			Context: context.Background(),
		}
		resp, err := log.client.Call.GetAppsAppCalls(&params)
		if err != nil {
			switch e := err.(type) {
			case *ccall.GetAppsAppCallsNotFound:
				return errors.New(e.Payload.Error.Message)
			default:
				return err
			}
		}
		calls := resp.Payload.Calls
		if len(calls) > 0 {
			callID = calls[0].ID
		} else {
			return errors.New("No previous calls found.")
		}
	}
	params := apicall.GetAppsAppCallsCallLogParams{
		Call:    callID,
		App:     app,
		Context: context.Background(),
	}
	resp, err := log.client.Operations.GetAppsAppCallsCallLog(&params)
	if err != nil {
		switch e := err.(type) {
		case *apicall.GetAppsAppCallsCallLogNotFound:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		default:
			return err
		}
	}
	fmt.Print(resp.Payload.Log.Log)
	return nil
}
