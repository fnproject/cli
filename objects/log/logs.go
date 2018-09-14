package log

import (
	"context"
	"errors"
	"fmt"

	apps "github.com/fnproject/cli/objects/app"
	fns "github.com/fnproject/cli/objects/fn"
	fnclient "github.com/fnproject/fn_go/clientv2"
	ccall "github.com/fnproject/fn_go/clientv2/call"
	apicall "github.com/fnproject/fn_go/clientv2/operations"
	"github.com/urfave/cli"
)

type logsCmd struct {
	client *fnclient.Fn
}

func (l *logsCmd) get(ctx *cli.Context) error {
	appName, fnName, callID := ctx.Args().Get(0), ctx.Args().Get(1), ctx.Args().Get(2)

	app, err := apps.GetAppByName(l.client, appName)
	if err != nil {
		return err
	}
	fn, err := fns.GetFnByName(l.client, app.ID, fnName)
	if err != nil {
		return nil
	}

	if callID == "last" || callID == "l" {
		params := ccall.GetFnsFnIDCallsParams{
			FnID:    fn.ID,
			Context: context.Background(),
		}
		resp, err := l.client.Call.GetFnsFnIDCalls(&params)
		if err != nil {
			switch e := err.(type) {
			case *ccall.GetFnsFnIDCallsNotFound:
				return errors.New(e.Payload.Message)
			default:
				return err
			}
		}
		calls := resp.Payload.Items
		if len(calls) > 0 {
			callID = calls[0].ID
		} else {
			return errors.New("no previous calls found")
		}
	}
	params := apicall.GetCallLogsParams{
		CallID:  callID,
		FnID:    fn.ID,
		Context: context.Background(),
	}
	resp, err := l.client.Operations.GetCallLogs(&params)
	if err != nil {
		switch e := err.(type) {
		case *apicall.GetCallLogsNotFound:
			return fmt.Errorf("%v", e.Payload.Message)
		default:
			return err
		}
	}
	fmt.Print(resp.Payload.Log)
	return nil
}
