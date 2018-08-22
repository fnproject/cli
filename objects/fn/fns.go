package fn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/run"
	"github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type fnsCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

/// FnFlags use to create/update functions
var FnFlags = []cli.Flag{
	cli.Uint64Flag{
		Name:  "memory,m",
		Usage: "memory in MiB",
	},
	cli.StringSliceFlag{
		Name:  "config,c",
		Usage: "fn configuration",
	},
	cli.StringFlag{
		Name:  "format,f",
		Usage: "hot container IO format - can be one of: default, http, json or cloudevent (check FDK docs to see which are supported for the FDK in use.)",
	},
	cli.IntFlag{
		Name:  "timeout",
		Usage: "fn timeout (eg. 30)",
	},
	cli.IntFlag{
		Name:  "idle-timeout",
		Usage: "fn idle timeout (eg. 30)",
	},
	cli.StringSliceFlag{
		Name:  "annotation",
		Usage: "fn annotation (can be specified multiple times)",
	},
}
var updateFnFlags = FnFlags

// CallFnFlags used to call a function
var CallFnFlags = append(run.RunFlags,
	cli.BoolFlag{
		Name:  "display-call-id",
		Usage: "whether display call ID or not",
	},
)

// WithSlash appends "/" to function path
func WithSlash(p string) string {
	p = path.Clean(p)

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// WithoutSlash removes "/" from route path
func WithoutSlash(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}

func (f *fnsCmd) list(c *cli.Context) error {
	appName := c.Args().Get(0)

	a, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &a.ID,
	}

	var resFns []*models.Fn
	for {
		resp, err := f.client.Fns.ListFns(params)

		if err != nil {
			return err
		}
		n := c.Int64("n")
		if n < 0 {
			return errors.New("number of calls: negative value not allowed")
		}

		resFns = append(resFns, resp.Payload.Items...)
		howManyMore := n - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	if len(resFns) == 0 {
		fmt.Fprintf(os.Stderr, "No functions found for app: %s\n", appName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "NAME", "\t", "IMAGE", "\n")
	for _, fn := range resFns {
		fmt.Fprint(w, fn.Name, "\t", fn.Image, "\t", "\n")
	}
	w.Flush()

	return nil
}

// WithFlags returns a function with specified flags
func FnWithFlags(c *cli.Context, fn *models.Fn) {
	if i := c.String("image"); i != "" {
		fn.Image = i
	}
	if f := c.String("format"); f != "" {
		fn.Format = f
	}

	if m := c.Uint64("memory"); m > 0 {
		fn.Memory = m
	}
	if len(fn.Config) == 0 {
		fn.Config = common.ExtractEnvConfig(c.StringSlice("config"))
	}
	if len(fn.Annotations) == 0 {
		if len(c.StringSlice("annotation")) > 0 {
			fn.Annotations = common.ExtractAnnotations(c)
		}
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

// WithFuncFile used when creating a function from a funcfile
func WithFuncFileV20180707(ff *common.FuncFileV20180707, fn *models.Fn) error {
	var err error
	if ff == nil {
		_, ff, err = common.LoadFuncFileV20180707(".")
		if err != nil {
			return err
		}
	}
	if ff.ImageNameV20180707() != "" { // args take precedence
		fn.Image = ff.ImageNameV20180707()
	}

	if ff.Format != "" {
		fn.Format = ff.Format
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

	fn := &models.Fn{}
	fn.Name = fnName
	fn.Image = c.Args().Get(2)

	FnWithFlags(c, fn)

	if fn.Name == "" {
		return errors.New("fnName path is missing")
	}
	if fn.Image == "" {
		return errors.New("no image specified")
	}

	return CreateFn(f.client, appName, fn)
}

// CreateFn request
func CreateFn(r *clientv2.Fn, appName string, fn *models.Fn) error {
	a, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}

	fn.AppID = a.ID
	image, err := common.ValidateImageName(fn.Image)
	if err != nil {
		return err
	}

	fn.Image = image
	resp, err := r.Fns.CreateFn(&apifns.CreateFnParams{
		Context: context.Background(),
		Body:    fn,
	})

	if err != nil {
		switch e := err.(type) {
		case *apifns.CreateFnBadRequest:
			return fmt.Errorf("%s", e.Payload.Message)
		case *apifns.CreateFnConflict:
			return fmt.Errorf("%s", e.Payload.Message)
		default:
			return err
		}
	}

	fmt.Println("Successfully created function:", resp.Payload.Name, "with", resp.Payload.Image)
	return nil
}

func PutFn(f *clientv2.Fn, fnID string, fn *models.Fn) error {
	if fn.Image != "" {
		_, err := common.ValidateImageName(fn.Image)
		if err != nil {
			return err
		}
	}

	_, err := f.Fns.UpdateFn(&apifns.UpdateFnParams{
		Context: context.Background(),
		FnID:    fnID,
		Body:    fn,
	})

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

func GetFnByName(client *clientv2.Fn, appID, fnName string) (*models.Fn, error) {
	resp, err := client.Fns.ListFns(&apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &appID,
		Name:    &fnName,
	})
	if err != nil {
		return nil, err
	}

	var fn *models.Fn
	for i := 0; i < len(resp.Payload.Items); i++ {
		if resp.Payload.Items[i].Name == fnName {
			fn = resp.Payload.Items[i]
		}
	}
	if fn == nil {
		return nil, fmt.Errorf("function %s not found", fnName)
	}

	return fn, nil
}

func (f *fnsCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	FnWithFlags(c, fn)

	err = PutFn(f.client, fn.ID, fn)
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

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	fn.Config = make(map[string]string)
	fn.Config[key] = value

	if err = PutFn(f.client, fn.ID, fn); err != nil {
		return fmt.Errorf("Error updating function configuration: %v", err)
	}

	fmt.Println(appName, fnName, "updated", key, "with", value)
	return nil
}

func (f *fnsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	key := c.Args().Get(2)

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
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

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
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

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}
	fn.Config[key] = ""

	err = PutFn(f.client, fn.ID, fn)
	if err != nil {
		return err
	}

	fmt.Printf("removed key '%s' from the function '%s' \n", key, fnName)
	return nil
}

func (f *fnsCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	prop := c.Args().Get(2)

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
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
		return errors.New("failed to inspect that fnName's field")
	}
	enc.Encode(field)

	return nil
}

func (f *fnsCmd) delete(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, err := app.GetAppByName(appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	params := apifns.NewDeleteFnParams()
	params.FnID = fn.ID
	_, err = f.client.Fns.DeleteFn(params)

	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, "deleted")
	return nil
}
