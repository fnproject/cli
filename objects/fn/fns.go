package fn

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type fnsCmd struct {
	providerAdapter  adapter.Provider
	apiClientAdapter adapter.APIClient
}

// FnFlags used to create/update functions
var FnFlags = []cli.Flag{
	cli.Uint64Flag{
		Name:  "memory,m",
		Usage: "Memory in MiB",
	},
	cli.StringSliceFlag{
		Name:  "config,c",
		Usage: "Function configuration",
	},
	cli.IntFlag{
		Name:  "timeout",
		Usage: "Function timeout (eg. 30)",
	},
	cli.IntFlag{
		Name:  "idle-timeout",
		Usage: "Function idle timeout (eg. 30)",
	},
	cli.StringSliceFlag{
		Name:  "annotation",
		Usage: "Function annotation (can be specified multiple times)",
	},
	cli.StringFlag{
		Name:  "image",
		Usage: "Function image",
	},
}
var updateFnFlags = FnFlags

// WithSlash appends "/" to function path
func WithSlash(p string) string {
	p = path.Clean(p)

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// WithoutSlash removes "/" from function path
func WithoutSlash(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}

func printFunctions(c *cli.Context, fns []*adapter.Fn) error {
	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		var newFns []interface{}
		for _, fn := range fns {
			newFns = append(newFns, struct {
				Name  string `json:"name"`
				Image string `json:"image"`
				ID    string `json:"id"`
			}{
				fn.Name,
				fn.Image,
				fn.ID,
			})
		}
		b, err := json.MarshalIndent(newFns, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "NAME", "\t", "IMAGE", "\t", "ID", "\n")

		for _, f := range fns {
			fmt.Fprint(w, f.Name, "\t", f.Image, "\t", f.ID, "\t", "\n")
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (f *fnsCmd) list(c *cli.Context) error {
	resFns, err := getFns(c, f.apiClientAdapter)
	if err != nil {
		return err
	}
	return printFunctions(c, resFns)
}

func getFns(c *cli.Context, apiClient adapter.APIClient) (fns []*adapter.Fn, err error) {
	appName := c.Args().Get(0)

	app, _, err := apiClient.AppClient().GetApp(appName)
	if err != nil {
		return nil, err
	}

	resFns, err := apiClient.FnClient().ListFn(app.ID, c.Int64("n"))

	if len(resFns) == 0 {
		return nil, fmt.Errorf("no functions found for app: %s", appName)
	}
	return resFns, nil
}

// BashCompleteFns can be called from a BashComplete function
// to provide function completion suggestions (Assumes the
// current context already contains an app name as an argument.
// This should be confirmed before calling this)
func BashCompleteFns(c *cli.Context) {
	providerAdapter, err := client.CurrentProviderAdapter()
	if err != nil {
		return
	}
	resp, err := getFns(c, providerAdapter.APIClient())
	if err != nil {
		return
	}
	for _, f := range resp {
		fmt.Println(f.Name)
	}
}

func getFnByAppAndFnName(appName, fnName string) (*adapter.Fn, error) {
	providerAdapter, err := client.CurrentProviderAdapter()
	if err != nil {
		return nil, errors.New("could not get context")
	}
	app, _, err := providerAdapter.APIClient().AppClient().GetApp(appName)
	if err != nil {
		return nil, fmt.Errorf("could not get app %v", appName)
	}
	fn, _, err := providerAdapter.APIClient().FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return nil, fmt.Errorf("could not get function %v", fnName)
	}
	return fn, nil
}

// WithFlags returns a function with specified flags
func WithFlags(c *cli.Context, fn *adapter.Fn) {
	if i := c.String("image"); i != "" {
		fn.Image = i
	}
	if m := c.Uint64("memory"); m > 0 {
		fn.Memory = m
	}

	fn.Config = common.ExtractConfig(c.StringSlice("config"))

	if len(c.StringSlice("annotation")) > 0 {
		fn.Annotations = common.ExtractAnnotations(c)
	}
	if t := c.Int("timeout"); t > 0 {
		to := int32(t)
		fn.Timeout = &to
	}
	if t := c.Int("idle-timeout"); t > 0 {
		to := int32(t)
		fn.IDLETimeout = &to
	}
}

// WithFuncFileV20180708 used when creating a function from a funcfile
func WithFuncFileV20180708(ff *common.FuncFileV20180708, fn *adapter.Fn) error {
	var err error
	if ff == nil {
		_, ff, err = common.LoadFuncFileV20180708(".")
		if err != nil {
			return err
		}
	}
	if ff.ImageNameV20180708() != "" { // args take precedence
		fn.Image = ff.ImageNameV20180708()
	}
	if ff.Timeout != nil {
		fn.Timeout = ff.Timeout
	}
	if ff.Memory != 0 {
		fn.Memory = ff.Memory
	}
	if ff.IDLE_timeout != nil {
		fn.IDLETimeout = ff.IDLE_timeout
	}

	if len(ff.Config) != 0 {
		fn.Config = ff.Config
	}
	if len(ff.Annotations) != 0 {
		fn.Annotations = ff.Annotations
	}
	// do something with triggers here

	return nil
}

func (f *fnsCmd) create(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	fn := &adapter.Fn{}
	fn.Name = fnName
	fn.Image = c.Args().Get(2)

	WithFlags(c, fn)

	if fn.Name == "" {
		return errors.New("fnName path is missing")
	}
	if fn.Image == "" {
		return errors.New("no image specified")
	}

	a, _, err := f.apiClientAdapter.AppClient().GetApp(appName)

	if err != nil {
		return err
	}

	_, err = CreateFn(f.apiClientAdapter.FnClient(), a.ID, fn)
	return err
}

// CreateFn request
func CreateFn(f adapter.FnClient, appID string, fn *adapter.Fn) (*adapter.Fn, error) {
	fn.AppID = appID
	err := common.ValidateTagImageName(fn.Image)
	if err != nil {
		return nil, err
	}

	resp, err := f.CreateFn(fn)

	if err != nil {
		switch e := err.(type) {
		case *apifns.CreateFnBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		case *apifns.CreateFnConflict:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}

	fmt.Println("Successfully created function:", resp.Name, "with", resp.Image)
	return resp, nil
}

// PutFn updates the fn with the given ID using the content of the provided fn
func PutFn(f adapter.FnClient, fnID string, fn *adapter.Fn, lock *string) error {
	if fn.Image != "" {
		err := common.ValidateTagImageName(fn.Image)
		if err != nil {
			return err
		}
	}

	fn.ID = fnID
	_, err := f.UpdateFn(fn, lock)

	if err != nil {
		switch e := err.(type) {
		case *apifns.UpdateFnBadRequest:
			return fmt.Errorf("%s", e.Payload.Message)

		default:
			return err
		}
	}

	return nil
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("function %s not found", n.Name)
}

func (f *fnsCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	err = f.apiClientAdapter.FnClient().HandleRetry(func() error {
		fn, etag, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
		if err != nil {
			return err
		}

		WithFlags(c, fn)

		return PutFn(f.apiClientAdapter.FnClient(), fn.ID, fn, etag)
	})

	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, "updated")
	return nil
}

func (f *fnsCmd) setConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)
	value := c.Args().Get(3)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	err = f.apiClientAdapter.FnClient().HandleRetry(func() error {
		fn, etag, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
		if err != nil {
			return err
		}

		fn.Config = make(map[string]string)
		fn.Config[key] = value

		return PutFn(f.apiClientAdapter.FnClient(), fn.ID, fn, etag)
	})

	if err != nil {
		return fmt.Errorf("Error updating function configuration: %v", err)
	}

	fmt.Println(appName, fnName, "updated", key, "with", value)
	return nil
}

func (f *fnsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	key := c.Args().Get(2)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}
	fn, _, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return err
	}

	val, ok := fn.Config[key]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (f *fnsCmd) listConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}
	fn, _, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return err
	}

	if len(fn.Config) == 0 {
		fmt.Fprintf(os.Stderr, "No config found for function: %s\n", fnName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "KEY", "\t", "VALUE", "\n")
	for key, val := range fn.Config {
		fmt.Fprint(w, key, "\t", val, "\n")
	}
	w.Flush()

	return nil
}

func (f *fnsCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	err = f.apiClientAdapter.FnClient().HandleRetry(func() error {
		fn, etag, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
		if err != nil {
			return err
		}
		_, ok := fn.Config[key]
		if !ok {
			fmt.Printf("Config key '%s' does not exist. Nothing to do.\n", key)
			return nil
		}
		fn.Config[key] = ""

		return PutFn(f.apiClientAdapter.FnClient(), fn.ID, fn, etag)
	})

	if err != nil {
		return err
	}

	fmt.Printf("Removed key '%s' from the function '%s' \n", key, fnName)
	return nil
}

func (f *fnsCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	prop := c.Args().Get(2)

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}
	fn, _, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return err
	}

	if c.Bool("endpoint") {
		endpoint, ok := fn.Annotations["fnproject.io/fn/invokeEndpoint"].(string)
		if !ok {
			return errors.New("missing or invalid endpoint on function")
		}
		fmt.Println(endpoint)
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(fn)
		return nil
	}

	data, err := json.Marshal(fn)
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %s", fnName, err)
	}
	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %s", fnName, err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return errors.New("failed to inspect that function's field")
	}
	enc.Encode(field)

	return nil
}

func (f *fnsCmd) delete(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	limit := c.Int64("n")

	app, _, err := f.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}
	fn, _, err := f.apiClientAdapter.FnClient().GetFn(app.ID, fnName)
	if err != nil {
		return err
	}

	//recursive delete of sub-objects
	if c.Bool("recursive") {
		triggers, err := f.apiClientAdapter.TriggerClient().ListTrigger(app.ID, fn.ID, limit)
		if err != nil {
			return fmt.Errorf("Failed to get associated objects: %s", err)
		}

		//Forced delete
		var shouldContinue bool
		if c.Bool("force") {
			shouldContinue = true
		} else {
			shouldContinue = common.UserConfirmedMultiResourceDeletion(nil, []*adapter.Fn{fn}, triggers)
		}

		if shouldContinue {
			err := common.DeleteTriggers(f.apiClientAdapter, triggers)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
		} else {
			return nil
		}
	}

	err = f.apiClientAdapter.FnClient().DeleteFn(fn.ID)

	if err != nil {
		return err
	}

	fmt.Println("Function", fnName, "deleted")
	return nil
}
