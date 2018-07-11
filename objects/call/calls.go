package call

import (
	"context"
	"errors"
	"fmt"
	"time"

	fnclient "github.com/fnproject/fn_go/client"
	apicall "github.com/fnproject/fn_go/client/call"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

type callsCmd struct {
	client *fnclient.Fn
}

func printCalls(calls []*models.Call) {
	for _, call := range calls {
		fmt.Println(fmt.Sprintf(
			"ID: %v\n"+
				"App Id: %v\n"+
				"Route: %v\n"+
				"Created At: %v\n"+
				"Started At: %v\n"+
				"Completed At: %v\n"+
				"Status: %v\n",
			call.ID, call.AppID, call.Path, call.CreatedAt,
			call.StartedAt, call.CompletedAt, call.Status))
		if call.Error != "" {
			fmt.Println(fmt.Sprintf("Error reason: %v\n", call.Error))
		}
	}
}

func (c *callsCmd) get(ctx *cli.Context) error {
	app, callID := ctx.Args().Get(0), ctx.Args().Get(1)
	params := apicall.GetAppsAppCallsCallParams{
		Call:    callID,
		App:     app,
		Context: context.Background(),
	}
	resp, err := c.client.Call.GetAppsAppCallsCall(&params)
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

func (c *callsCmd) list(ctx *cli.Context) error {
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

	n := ctx.Int64("n")
	if n < 0 {
		return errors.New("Number of calls: negative value not allowed")
	}

	var resCalls []*models.Call
	for {
		resp, err := c.client.Call.GetAppsAppCalls(&params)
		if err != nil {
			switch e := err.(type) {
			case *apicall.GetAppsAppCallsNotFound:
				return errors.New(e.Payload.Error.Message)
			default:
				return err
			}
		}

		resCalls = append(resCalls, resp.Payload.Calls...)
		howManyMore := n - int64(len(resCalls)+len(resp.Payload.Calls))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	printCalls(resCalls)
	return nil
}
