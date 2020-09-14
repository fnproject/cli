package trigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jmoiron/jsonq"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

type triggersCmd struct {
	providerAdapter  adapter.Provider
	apiClientAdapter adapter.APIClient
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

	app, err := t.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	fn, err := t.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return err
	}

	trigger := &adapter.Trigger{
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

	_, err = t.apiClientAdapter.TriggerClient().CreateTrigger(trigger)
	return err
}

func validateTriggerSource(ts string) string {
	if !strings.HasPrefix(ts, "/") {
		ts = "/" + ts
	}
	return ts
}

func (t *triggersCmd) list(c *cli.Context) error {
	resTriggers, err := getTriggers(c, t.apiClientAdapter)
	if err != nil {
		return err
	}
	fnName := c.Args().Get(1)
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

			fn, err := t.apiClientAdapter.FnClient().GetFnByFnID(trigger.FnID)
			if err != nil {
				return err
			}
			fnName = fn.Name
			fmt.Fprint(w, fnName, "\t", trigger.Name, "\t", trigger.ID, "\t", trigger.Type, "\t", trigger.Source, "\t", endpoint, "\n")
		}
	}
	w.Flush()
	return nil
}

func getTriggers(c *cli.Context, apiClient adapter.APIClient) ([]*adapter.Trigger, error) {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	limit := c.Int64("n")

	app, err := apiClient.AppClient().GetApp(appName)
	if err != nil {
		return nil, err
	}

	var fnID string
	if len(fnName) != 0 {
		fn, err := apiClient.FnClient().GetFn(app.ID, fnName)
		if err != nil {
			return nil, err
		}
		fnID = fn.ID
	}

	resTriggers, err := apiClient.TriggerClient().ListTrigger(app.ID, fnID, limit)
	if len(resTriggers) == 0 {
		if len(fnID) == 0 {
			return nil, fmt.Errorf("no triggers found for app: %s", appName)
		}
		return nil, fmt.Errorf("no triggers found for function: %s", fnName)
	}
	return resTriggers, nil
}

// BashCompleteTriggers can be called from a BashComplete function
// to provide function completion suggestions (Assumes the
// current context already contains an app name and a function name
// as the first 2 arguments. This should be confirmed before calling this)
func BashCompleteTriggers(c *cli.Context) {
	providerAdapter, err := client.CurrentProviderAdapter()
	if err != nil {
		return
	}
	resp, err := getTriggers(c, providerAdapter.APIClient())
	if err != nil {
		return
	}
	for _, t := range resp {
		fmt.Println(t.Name)
	}
}

func (t *triggersCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)

	trigger, err := GetTrigger(t.apiClientAdapter, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	WithFlags(c, trigger)

	_, err = t.apiClientAdapter.TriggerClient().UpdateTrigger(trigger)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, triggerName, "updated")
	return nil
}

func (t *triggersCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	triggerName := c.Args().Get(2)
	prop := c.Args().Get(3)

	trigger, err := GetTrigger(t.apiClientAdapter, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	if c.Bool("endpoint") {
		endpoint, ok := trigger.Annotations["fnproject.io/trigger/httpEndpoint"].(string)
		if !ok {
			return errors.New("missing or invalid http endpoint on trigger")
		}
		fmt.Println(endpoint)
		return nil
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

	trigger, err := GetTrigger(t.apiClientAdapter, appName, fnName, triggerName)
	if err != nil {
		return err
	}

	err = t.apiClientAdapter.TriggerClient().DeleteTrigger(trigger.ID)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, triggerName, "deleted")
	return nil
}

// GetTrigger looks up a trigger using the provided client by app, function and trigger name
func GetTrigger(apiClient adapter.APIClient, appName, fnName, triggerName string) (*adapter.Trigger, error) {
	app, err := apiClient.AppClient().GetApp(appName)
	if err != nil {
		return nil, err
	}

	fn, err := apiClient.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return nil, err
	}

	trigger, err := apiClient.TriggerClient().GetTrigger(app.ID, fn.ID, triggerName)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

// GetTriggerByAppFnAndTriggerNames looks up a trigger using app, fn and trigger names
func GetTriggerByAppFnAndTriggerNames(appName, fnName, triggerName string) (*adapter.Trigger, error) {
	providerAdapter, err := client.CurrentProviderAdapter()
	if err != nil {
		return nil, err
	}
	client := providerAdapter.APIClient()
	return GetTrigger(client, appName, fnName, triggerName)
}

// WithFlags returns a trigger with the specified flags
func WithFlags(c *cli.Context, t *adapter.Trigger) {
	if len(c.StringSlice("annotation")) > 0 {
		t.Annotations = common.ExtractAnnotations(c)
	}
}
