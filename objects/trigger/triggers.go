package trigger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/fn_go/clientv2"
	apitriggers "github.com/fnproject/fn_go/clientv2/triggers"
	fnmodels "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/urfave/cli"
)

type triggersCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

var TriggerFlags = []cli.Flag{
	cli.StringSliceFlag{
		Name:  "source,s",
		Usage: "trigger source",
	},
	cli.StringFlag{
		Name:  "type, t",
		Usage: "Todo",
	},
	cli.StringSliceFlag{
		Name:  "annotation",
		Usage: "fn annotation (can be specified multiple times)",
	},
}

func (t *triggersCmd) create(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)

	app, err := app.GetAppByName(t.client, appName)
	if err != nil {
		return err
	}

	fn, err := fn.GetFnByName(t.client, app.ID, fnName)
	if err != nil {
		return nil
	}

	trigger := &fnmodels.Trigger{
		AppID: app.ID,
		FnID:  fn.ID,
	}

	trigger.Name = triggerName

	WithFlags(c, trigger)

	if trigger.Name == "" {
		return errors.New("triggerName path is missing")
	}

	return CreateTrigger(t.client, trigger)
}

func CreateTrigger(client *clientv2.Fn, trigger *fnmodels.Trigger) error {
	resp, err := client.Triggers.CreateTrigger(&apitriggers.CreateTriggerParams{
		Context: context.Background(),
		Body:    trigger,
	})

	if err != nil {
		switch e := err.(type) {
		case *apitriggers.CreateTriggerBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		case *apitriggers.CreateTriggerConflict:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	fmt.Println(resp.Payload.Name, "create with")
	return nil
}

func WithFlags(c *cli.Context, t *fnmodels.Trigger) {
	if len(c.StringSlice("annotation")) > 0 {
		t.Annotations = common.ExtractAnnotations(c)
	}
}

func (t *triggersCmd) list(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, err := app.GetAppByName(t.client, appName)
	fmt.Println("app.ID: ", app.ID)
	if err != nil {
		return err
	}

	fn, err := fn.GetFnByName(t.client, app.ID, fnName)
	if err != nil {
		return nil
	}

	params := &apitriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &app.ID,
		FnID:    &fn.ID,
	}

	var resTriggers []*fnmodels.Trigger
	for {
		resp, err := t.client.Triggers.ListTriggers(params)
		fmt.Println("Err 1: ", err)

		if err != nil {
			return err
		}
		n := c.Int64("n")
		if n < 0 {
			return errors.New("number of calls: negative value not allowed")
		}

		resTriggers = append(resTriggers, resp.Payload.Items...)
		howManyMore := n - int64(len(resTriggers)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	callURL := t.provider.CallURL()

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "name", "\t", "image", "\t", "endpoint", "\n")
	for _, trigger := range resTriggers {
		endpoint := path.Join(callURL.Host, "t", trigger.Name)
		fmt.Fprint(w, trigger.Name, "\t", trigger.Source, "\t", endpoint, "\n")
	}
	w.Flush()
	return nil
}

func (t *triggersCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)

	app, err := app.GetAppByName(t.client, appName)
	if err != nil {
		return err
	}

	fn, err := fn.GetFnByName(t.client, app.ID, fnName)
	if err != nil {
		return err
	}

	trigger, err := GetTriggerByName(t.client, app.ID, fn.ID, triggerName)
	if err != nil {
		return err
	}

	WithFlags(c, trigger)

	err = updateTrigger(t.client, trigger)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, triggerName, "updated")
	return nil
}

func updateTrigger(client *clientv2.Fn, trigger *fnmodels.Trigger) error {
	_, err := client.Triggers.UpdateTrigger(&apitriggers.UpdateTriggerParams{
		Context: context.Background(),
		Body:    trigger,
	})

	if err != nil {
		switch e := err.(type) {
		case *apitriggers.UpdateTriggerBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	return nil
}

func GetTriggerByName(client *clientv2.Fn, appId string, fnId string, triggerName string) (*fnmodels.Trigger, error) {
	triggerList, err := client.Triggers.ListTriggers(&apitriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &appId,
		FnID:    &fnId,
		Name:    &triggerName,
	})

	if err != nil {
		return nil, err
	}
	if len(triggerList.Payload.Items) == 0 {
		return nil, fmt.Errorf("Trigger %s not found", triggerName)
	}

	return triggerList.Payload.Items[0], nil
}

func (t *triggersCmd) delete(c *cli.Context) error {
	return nil
}
