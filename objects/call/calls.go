package call

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	apps "github.com/fnproject/cli/objects/app"
	fns "github.com/fnproject/cli/objects/fn"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apicall "github.com/fnproject/fn_go/clientv2/call"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"
)

type callsCmd struct {
	client *fnclient.Fn
}

// getMarshalableCall returns a call struct that we can marshal to JSON and output
func getMarshalableCall(call *modelsv2.Call) interface{} {
	if call.Error != "" {
		return struct {
			ID          string          `json:"id"`
			AppID       string          `json:"appId"`
			FnID        string          `json:"fnId"`
			CreatedAt   strfmt.DateTime `json:"createdAt"`
			StartedAt   strfmt.DateTime `json:"startedAt"`
			CompletedAt strfmt.DateTime `json:"completedAt"`
			Status      string          `json:"status"`
			ErrorReason string          `json:"errorReason"`
		}{
			call.ID,
			call.AppID,
			call.FnID,
			call.CreatedAt,
			call.StartedAt,
			call.CompletedAt,
			call.Status,
			call.Error,
		}
	}

	return struct {
		ID          string          `json:"id"`
		AppID       string          `json:"appId"`
		FnID        string          `json:"fnId"`
		CreatedAt   strfmt.DateTime `json:"createdAt"`
		StartedAt   strfmt.DateTime `json:"startedAt"`
		CompletedAt strfmt.DateTime `json:"completedAt"`
		Status      string          `json:"status"`
	}{
		call.ID,
		call.AppID,
		call.FnID,
		call.CreatedAt,
		call.StartedAt,
		call.CompletedAt,
		call.Status,
	}
}

func printCalls(c *cli.Context, calls []*modelsv2.Call) error {
	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		var allCalls []interface{}
		for _, call := range calls {
			c := getMarshalableCall(call)
			allCalls = append(allCalls, c)
		}
		b, err := json.MarshalIndent(allCalls, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		for _, call := range calls {
			fmt.Println(fmt.Sprintf(
				"ID: %v\n"+
					"App Id: %v\n"+
					"Fn Id: %v\n"+
					"Created At: %v\n"+
					"Started At: %v\n"+
					"Completed At: %v\n"+
					"Status: %v\n",
				call.ID, call.AppID, call.FnID, call.CreatedAt,
				call.StartedAt, call.CompletedAt, call.Status))
			if call.Error != "" {
				fmt.Println(fmt.Sprintf("Error reason: %v\n", call.Error))
			}
		}
	}
	return nil
}

func (c *callsCmd) get(ctx *cli.Context) error {
	appName, fnName, callID := ctx.Args().Get(0), ctx.Args().Get(1), ctx.Args().Get(2)

	app, err := apps.GetAppByName(c.client, appName)
	if err != nil {
		return err
	}
	fn, err := fns.GetFnByName(c.client, app.ID, fnName)
	if err != nil {
		return err
	}
	params := apicall.GetFnsFnIDCallsCallIDParams{
		CallID:  callID,
		FnID:    fn.ID,
		Context: context.Background(),
	}
	resp, err := c.client.Call.GetFnsFnIDCallsCallID(&params)
	if err != nil {
		switch e := err.(type) {
		case *apicall.GetFnsFnIDCallsCallIDNotFound:
			return errors.New(e.Payload.Message)
		default:
			return err
		}
	}
	printCalls(ctx, []*modelsv2.Call{resp.Payload})
	return nil
}

func (c *callsCmd) list(ctx *cli.Context) error {
	appName, fnName := ctx.Args().Get(0), ctx.Args().Get(1)

	app, err := apps.GetAppByName(c.client, appName)
	if err != nil {
		return err
	}
	fn, err := fns.GetFnByName(c.client, app.ID, fnName)
	if err != nil {
		return err
	}
	params := apicall.GetFnsFnIDCallsParams{
		FnID:    fn.ID,
		Context: context.Background(),
	}
	if ctx.String("cursor") != "" {
		cursor := ctx.String("cursor")
		params.Cursor = &cursor
	}
	if ctx.String("from-time") != "" {
		fromTime := ctx.String("from-time")
		fromTimeInt64, err := time.Parse(time.RFC3339, fromTime)
		if err != nil {
			return err
		}
		res := fromTimeInt64.Unix()
		params.FromTime = &res

	}

	if ctx.String("to-time") != "" {
		toTime := ctx.String("to-time")
		toTimeInt64, err := time.Parse(time.RFC3339, toTime)
		if err != nil {
			return err
		}
		res := toTimeInt64.Unix()
		params.ToTime = &res
	}

	n := ctx.Int64("n")
	if n < 0 {
		return errors.New("Number of calls: negative value not allowed")
	}

	var resCalls []*modelsv2.Call
	for {
		resp, err := c.client.Call.GetFnsFnIDCalls(&params)
		if err != nil {
			switch e := err.(type) {
			case *apicall.GetFnsFnIDCallsNotFound:
				return errors.New(e.Payload.Message)
			default:
				return err
			}
		}

		resCalls = append(resCalls, resp.Payload.Items...)
		howManyMore := n - int64(len(resCalls)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	printCalls(ctx, resCalls)
	return nil
}
