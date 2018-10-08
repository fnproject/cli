package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fnproject/fn_go/clientv2/fns"

	"github.com/jmoiron/jsonq"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/fn_go/clientv2"
	apiTriggers "github.com/fnproject/fn_go/clientv2/triggers"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/urfave/cli"
)

type triggersCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

// TriggerFlags used to create/update triggers
var TriggerFlags = []cli.Flag{
	cli.StringFlag{
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

	trigger := &models.Trigger{
		AppID: app.ID,
		FnID:  fn.ID,
	}

	trigger.Name = triggerName

	if triggerType := c.String("type"); triggerType != "" {
		trigger.Type = triggerType
	}

	if triggerSource := c.String("source"); triggerSource != "" {
		trigger.Source = validateTriggerSource(triggerSource)
	}

	WithFlags(c, trigger)

	if trigger.Name == "" {
		return errors.New("triggerName path is missing")
	}

	return CreateTrigger(t.client, trigger)
}

func validateTriggerSource(ts string) string {
	if !strings.HasPrefix(ts, "/") {
		ts = "/" + ts
	}
	return ts
}

// CreateTrigger request
func CreateTrigger(client *clientv2.Fn, trigger *models.Trigger) error {
	resp, err := client.Triggers.CreateTrigger(&apiTriggers.CreateTriggerParams{
		Context: context.Background(),
		Body:    trigger,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiTriggers.CreateTriggerBadRequest:
			fmt.Println(e)
			return fmt.Errorf("%s", e.Payload.Message)
		case *apiTriggers.CreateTriggerConflict:
			return fmt.Errorf("%s", e.Payload.Message)
		default:
			return err
		}
	}

	fmt.Println("Successfully created trigger:", resp.Payload.Name)
	return nil
}

func (t *triggersCmd) list(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	var params *apiTriggers.ListTriggersParams

	app, err := app.GetAppByName(t.client, appName)
	if err != nil {
		return err
	}

	if len(fnName) == 0 {
		params = &apiTriggers.ListTriggersParams{
			Context: context.Background(),
			AppID:   &app.ID,
		}

	} else {

		fn, err := fn.GetFnByName(t.client, app.ID, fnName)
		if err != nil {
			return err
		}
		params = &apiTriggers.ListTriggersParams{
			Context: context.Background(),
			AppID:   &app.ID,
			FnID:    &fn.ID,
		}
	}

	var resTriggers []*models.Trigger
	for {

		resp, err := t.client.Triggers.ListTriggers(params)
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

	if len(resTriggers) == 0 {
		if len(fnName) == 0 {
			fmt.Fprintf(os.Stderr, "No triggers found for app: %s\n", appName)
		} else {
			fmt.Fprintf(os.Stderr, "No triggers found for function: %s\n", fnName)
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if len(fnName) != 0 {

		fmt.Fprint(w, "NAME", "\t", "ID", "\t", "TYPE", "\t", "SOURCE", "\t", "ENDPOINT", "\n")
		for _, trigger := range resTriggers {
			endpoint := trigger.Annotations["fnproject.io/trigger/httpEndpoint"]
			fmt.Fprint(w, trigger.Name, "\t", trigger.ID, "\t", trigger.Type, "\t", trigger.Source, "\t", endpoint, "\n")
		}
	} else {
		fmt.Fprint(w, "FUNCTION", "\t", "NAME", "\t", "ID", "\t", "TYPE", "\t", "SOURCE", "\t", "ENDPOINT", "\n")
		for _, trigger := range resTriggers {
			endpoint := trigger.Annotations["fnproject.io/trigger/httpEndpoint"]

			resp, err := t.client.Fns.GetFn(&fns.GetFnParams{
				FnID:    trigger.FnID,
				Context: context.Background(),
			})
			if err != nil {
				return err
			}
			fnName = resp.Payload.Name
			fmt.Fprint(w, fnName, "\t", trigger.Name, "\t", trigger.ID, "\t", trigger.Type, "\t", trigger.Source, "\t", endpoint, "\n")
		}
	}
	w.Flush()
	return nil
}

func (t *triggersCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)

	trigger, err := GetTrigger(t.client, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	WithFlags(c, trigger)

	err = PutTrigger(t.client, trigger)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, triggerName, "updated")
	return nil
}

// PutTrigger updates the provided trigger with new values
func PutTrigger(t *clientv2.Fn, trigger *models.Trigger) error {
	_, err := t.Triggers.UpdateTrigger(&apiTriggers.UpdateTriggerParams{
		Context:   context.Background(),
		TriggerID: trigger.ID,
		Body:      trigger,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiTriggers.UpdateTriggerBadRequest:
			return fmt.Errorf("%s", e.Payload.Message)
		default:
			return err
		}
	}

	return nil
}

func (t *triggersCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)
	prop := c.Args().Get(3)

	trigger, err := GetTrigger(t.client, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(trigger)
		return nil
	}

	data, err := json.Marshal(trigger)
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %s", triggerName, err)
	}

	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %s", triggerName, err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return errors.New("failed to inspect %s field names")
	}
	enc.Encode(field)

	return nil
}

func (t *triggersCmd) delete(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)

	trigger, err := GetTrigger(t.client, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	params := apiTriggers.NewDeleteTriggerParams()
	params.TriggerID = trigger.ID

	_, err = t.client.Triggers.DeleteTrigger(params)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, triggerName, "deleted")
	return nil
}

// GetTrigger looks up a trigger using the provided client by app, function and trigger name
func GetTrigger(client *clientv2.Fn, appName, fnName, triggerName string) (*models.Trigger, error) {
	app, err := app.GetAppByName(client, appName)
	if err != nil {
		return nil, err
	}

	fn, err := fn.GetFnByName(client, app.ID, fnName)
	if err != nil {
		return nil, err
	}

	trigger, err := GetTriggerByName(client, app.ID, fn.ID, triggerName)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

// GetTriggerByName looks up a trigger using the provided client by app and function ID and trigger name
func GetTriggerByName(client *clientv2.Fn, appID string, fnID string, triggerName string) (*models.Trigger, error) {
	triggerList, err := client.Triggers.ListTriggers(&apiTriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &appID,
		FnID:    &fnID,
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

// WithFlags returns a trigger with the specified flags
func WithFlags(c *cli.Context, t *models.Trigger) {
	if len(c.StringSlice("annotation")) > 0 {
		t.Annotations = common.ExtractAnnotations(c)
	}
}
