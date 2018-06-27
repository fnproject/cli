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
	fnclient "github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	fnmodels "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type fnsCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

// RouteFlags use to create/update routes
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
		Usage: "hot container IO format - default or http",
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

// CallFnFlags used to call a route
var CallFnFlags = append(run.RunFlags,
	cli.BoolFlag{
		Name:  "display-call-id",
		Usage: "whether display call ID or not",
	},
)

// WithSlash appends "/" to route path
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

	a, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &a.ID,
	}

	var resFns []*fnmodels.Fn
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
		fmt.Println("resFns: ", resFns)
		howManyMore := n - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	callURL := f.provider.CallURL()
	fmt.Println("Call URl:", callURL)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "name", "\t", "image", "\t", "endpoint", "\n")
	for _, route := range resFns {
		endpoint := path.Join(callURL.Host, "f", appName, route.Name)
		fmt.Fprint(w, route.Name, "\t", route.Image, "\t", endpoint, "\n")
	}
	w.Flush()
	return nil
}

// WithFlags returns a route with specified flags
func WithFlags(c *cli.Context, rt *fnmodels.Fn) {
	if i := c.String("image"); i != "" {
		rt.Image = i
	}
	if f := c.String("format"); f != "" {
		rt.Format = f
	}

	if m := c.Uint64("memory"); m > 0 {
		rt.Mem = m
	}
	//if rt.Cpus == "" {
	//	if m := c.String("cpus"); m != "" {
	//		rt.Cpus = m
	//	}
	//}
	if t := c.Int("timeout"); t > 0 {
		to := int32(t)
		rt.Timeout = &to
	}
	if t := c.Int("idle-timeout"); t > 0 {
		to := int32(t)
		rt.IDLETimeout = &to
	}

	if len(c.StringSlice("config")) > 0 {
		rt.Config = common.ExtractEnvConfig(c.StringSlice("config"))
	}
	if len(c.StringSlice("annotation")) > 0 {
		rt.Annotations = common.ExtractAnnotations(c)
	}
}

// WithFuncFile used when creating a route from a funcfile
func WithFuncFile(ff *common.FuncFile, rt *fnmodels.Fn) error {
	var err error
	if ff == nil {
		_, ff, err = common.LoadFuncfile()
		if err != nil {
			return err
		}
	}
	if ff.ImageName() != "" { // args take precedence
		rt.Image = ff.ImageName()
	}
	if ff.Format != "" {
		rt.Format = ff.Format
	}
	if ff.Timeout != nil {
		rt.Timeout = ff.Timeout
	}
	if ff.Memory != 0 {
		rt.Mem = ff.Memory
	}
	if ff.IDLETimeout != nil {
		rt.IDLETimeout = ff.IDLETimeout
	}

	if len(ff.Config) != 0 {
		rt.Config = ff.Config
	}
	if len(ff.Annotations) != 0 {
		rt.Annotations = ff.Annotations
	}

	return nil
}

func (f *fnsCmd) create(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	a, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}

	rt := &fnmodels.Fn{
		AppID: a.ID,
	}
	rt.Name = fnName
	rt.Image = c.Args().Get(2)

	WithFlags(c, rt)

	if rt.Name == "" {
		return errors.New("fnName path is missing")
	}
	if rt.Image == "" {
		return errors.New("no image specified")
	}

	return CreateFn(f.client, rt)
}

// CreateFn request
func CreateFn(r *clientv2.Fn, rt *fnmodels.Fn) error {
	err := common.ValidateImageName(rt.Image)
	if err != nil {
		return err
	}

	resp, err := r.Fns.CreateFn(&apifns.CreateFnParams{
		Context: context.Background(),
		Body:    rt,
	})

	if err != nil {
		switch e := err.(type) {
		case *apifns.CreateFnBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		case *apifns.CreateFnConflict:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	fmt.Println(resp.Payload.Name, "created with", resp.Payload.Image)
	return nil
}

// PatchRoute request
func UpdateFn(client *fnclient.Fn, fn *fnmodels.Fn) error {
	fmt.Println("Fn: ", fn)
	if fn.Image != "" {
		err := common.ValidateImageName(fn.Image)
		if err != nil {
			return err
		}
	}

	_, err := client.Fns.UpdateFn(&apifns.UpdateFnParams{
		Context: context.Background(),
		Body:    fn,
	})

	if err != nil {
		switch e := err.(type) {
		case *apifns.UpdateFnBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)

		default:
			return err
		}
	}

	return nil
}

func GetFnByName(client *clientv2.Fn, appId string, fnName string) (*fnmodels.Fn, error) {
	fnList, err := client.Fns.ListFns(&apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &appId,
		Name:    &fnName,
	})

	if err != nil {
		return nil, err
	}
	if len(fnList.Payload.Items) == 0 {
		return nil, fmt.Errorf("Function %s not found", fnName)
	}

	return fnList.Payload.Items[0], nil
}

func (f *fnsCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	WithFlags(c, fn)

	err = UpdateFn(f.client, fn)
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

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	fn.Config[key] = value

	err = UpdateFn(f.client, fn)
	if err != nil {
		return err
	}

	fmt.Println(appName, fnName, "updated", key, "with", value)
	return nil
}

func (f *fnsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	key := c.Args().Get(2)

	app, err := app.GetAppByName(f.client, appName)
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

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	for key, val := range fn.Config {
		fmt.Printf("%s=%s\n", key, val)
	}

	return nil
}

func (f *fnsCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}
	fn.Config[key] = ""

	err = UpdateFn(f.client, fn)
	if err != nil {
		return err
	}

	fmt.Printf("removed key '%s' from the function '%s.%s'", key, appName, key)
	return nil
}

func (f *fnsCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := WithoutSlash(c.Args().Get(1))
	prop := c.Args().Get(2)

	app, err := app.GetAppByName(f.client, appName)
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

	app, err := app.GetAppByName(f.client, appName)
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
