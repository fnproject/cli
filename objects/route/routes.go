package route

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	common "github.com/fnproject/cli/common"
	run "github.com/fnproject/cli/run"
	fnclient "github.com/fnproject/fn_go/client"
	apiroutes "github.com/fnproject/fn_go/client/routes"
	fnmodels "github.com/fnproject/fn_go/models"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type routesCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
}

// RouteFlags use to create/update routes
var RouteFlags = []cli.Flag{
	cli.Uint64Flag{
		Name:  "memory,m",
		Usage: "Memory in MiB",
	},
	cli.StringFlag{
		Name:  "type,t",
		Usage: "Route type - sync or async",
	},
	cli.StringSliceFlag{
		Name:  "config,c",
		Usage: "Route configuration",
	},
	cli.StringSliceFlag{
		Name:  "headers",
		Usage: "Route response headers",
	},
	cli.StringFlag{
		Name:  "format,f",
		Usage: "Hot container IO format - default or http",
	},
	cli.IntFlag{
		Name:  "timeout",
		Usage: "Route timeout (eg. 30)",
	},
	cli.IntFlag{
		Name:  "idle-timeout",
		Usage: "Route idle timeout (eg. 30)",
	},
	cli.StringSliceFlag{
		Name:  "annotation",
		Usage: "Route annotation (can be specified multiple times)",
	},
}
var updateRouteFlags = RouteFlags

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

func (r *routesCmd) list(c *cli.Context) error {
	appName := c.Args().Get(0)

	params := &apiroutes.GetAppsAppRoutesParams{
		Context: context.Background(),
		App:     appName,
	}

	var resRoutes []*fnmodels.Route
	for {
		resp, err := r.client.Routes.GetAppsAppRoutes(params)

		if err != nil {
			switch e := err.(type) {
			case *apiroutes.GetAppsAppRoutesNotFound:
				return fmt.Errorf("%s", e.Payload.Error.Message)
			default:
				return err
			}
		}
		n := c.Int64("n")
		if n < 0 {
			return errors.New("Number of calls: negative value not allowed")
		}

		resRoutes = append(resRoutes, resp.Payload.Routes...)
		howManyMore := n - int64(len(resRoutes)+len(resp.Payload.Routes))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	callURL, err := r.provider.CallURL(appName)
	if err != nil {
		return err
	}

	if len(resRoutes) == 0 {
		fmt.Fprint(os.Stderr, "No routes found for app: %s\n", appName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "PATH", "\t", "IMAGE", "\t", "ENDPOINT", "\n")
	for _, route := range resRoutes {
		endpoint := path.Join(callURL.Host, "r", appName, route.Path)
		fmt.Fprint(w, route.Path, "\t", route.Image, "\t", endpoint, "\n")
	}
	w.Flush()
	return nil
}

// WithFlags returns a route with specified flags
func WithFlags(c *cli.Context, rt *fnmodels.Route) {
	if rt.Image == "" {
		if i := c.String("image"); i != "" {
			rt.Image = i
		}
	}
	if rt.Format == "" {
		if f := c.String("format"); f != "" {
			rt.Format = f
		}
	}
	if rt.Type == "" {
		if t := c.String("type"); t != "" {
			rt.Type = t
		}
	}
	if rt.Memory == 0 {
		if m := c.Uint64("memory"); m > 0 {
			rt.Memory = m
		}
	}
	if rt.Cpus == "" {
		if m := c.String("cpus"); m != "" {
			rt.Cpus = m
		}
	}
	if rt.Timeout == nil {
		if t := c.Int("timeout"); t > 0 {
			to := int32(t)
			rt.Timeout = &to
		}
	}
	if rt.IDLETimeout == nil {
		if t := c.Int("idle-timeout"); t > 0 {
			to := int32(t)
			rt.IDLETimeout = &to
		}
	}
	if len(rt.Headers) == 0 {
		if len(c.StringSlice("headers")) > 0 {
			headers := map[string][]string{}
			for _, header := range c.StringSlice("headers") {
				parts := strings.Split(header, "=")
				headers[parts[0]] = strings.Split(parts[1], ";")
			}
			rt.Headers = headers
		}
	}
	if len(rt.Config) == 0 {
		if len(c.StringSlice("config")) > 0 {
			rt.Config = common.ExtractEnvConfig(c.StringSlice("config"))
		}
	}
	if len(rt.Annotations) == 0 {
		if len(c.StringSlice("annotation")) > 0 {
			rt.Annotations = common.ExtractAnnotations(c)
		}
	}
}

// WithFuncFile used when creating a route from a funcfile
func WithFuncFile(ff *common.FuncFile, rt *fnmodels.Route) error {
	var err error
	if ff == nil {
		_, ff, err = common.LoadFuncfile(".")
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
	if rt.Path == "" && ff.Path != "" {
		rt.Path = ff.Path
	}
	if rt.Type == "" && ff.Type != "" {
		rt.Type = ff.Type
	}
	if ff.Memory != 0 {
		rt.Memory = ff.Memory
	}
	if ff.Cpus != "" {
		rt.Cpus = ff.Cpus
	}
	if ff.IDLETimeout != nil {
		rt.IDLETimeout = ff.IDLETimeout
	}
	if len(ff.Headers) != 0 {
		rt.Headers = ff.Headers
	}
	if len(ff.Config) != 0 {
		rt.Config = ff.Config
	}
	if len(ff.Annotations) != 0 {
		rt.Annotations = ff.Annotations
	}

	return nil
}

func (r *routesCmd) create(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithSlash(c.Args().Get(1))

	rt := &fnmodels.Route{}
	rt.Path = route
	rt.Image = c.Args().Get(2)

	WithFlags(c, rt)

	if rt.Path == "" {
		return errors.New("Route path is missing")
	}
	if rt.Image == "" {
		return errors.New("No image specified")
	}

	return PostRoute(r.client, appName, rt)
}

// PostRoute request
func PostRoute(r *fnclient.Fn, appName string, rt *fnmodels.Route) error {
	image, err := common.ValidateImageName(rt.Image)
	if err != nil {
		return err
	}
	rt.Image = image

	body := &fnmodels.RouteWrapper{
		Route: rt,
	}

	resp, err := r.Routes.PostAppsAppRoutes(&apiroutes.PostAppsAppRoutesParams{
		Context: context.Background(),
		App:     appName,
		Body:    body,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiroutes.PostAppsAppRoutesBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		case *apiroutes.PostAppsAppRoutesConflict:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	fmt.Println(resp.Payload.Route.Path, "created with", resp.Payload.Route.Image)
	return nil
}

// PatchRoute request
func PatchRoute(r *fnclient.Fn, appName, routePath string, rt *fnmodels.Route) error {
	if rt.Image != "" {
		_, err := common.ValidateImageName(rt.Image)
		if err != nil {
			return err
		}
	}

	_, err := r.Routes.PatchAppsAppRoutesRoute(&apiroutes.PatchAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   WithoutSlash(routePath),
		Body:    &fnmodels.RouteWrapper{Route: rt},
	})

	if err != nil {
		switch e := err.(type) {
		case *apiroutes.PatchAppsAppRoutesRouteBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		case *apiroutes.PatchAppsAppRoutesRouteNotFound:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	return nil
}

// PutRoute request
func PutRoute(r *fnclient.Fn, appName, routePath string, rt *fnmodels.Route) error {
	_, err := r.Routes.PutAppsAppRoutesRoute(&apiroutes.PutAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   WithoutSlash(routePath),
		Body:    &fnmodels.RouteWrapper{Route: rt},
	})
	if err != nil {
		switch e := err.(type) {
		case *apiroutes.PutAppsAppRoutesRouteBadRequest:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}
	return nil
}

func (r *routesCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))

	rt := &fnmodels.Route{}

	WithFlags(c, rt)

	err := PatchRoute(r.client, appName, route, rt)
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "updated")
	return nil
}

func (r *routesCmd) setConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)
	value := c.Args().Get(3)

	rt := fnmodels.Route{
		Config: make(map[string]string),
	}

	rt.Config[key] = value

	err := PatchRoute(r.client, appName, route, &rt)
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "updated", key, "with", value)
	return nil
}

func (r *routesCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)

	resp, err := r.client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   route,
	})

	if err != nil {
		return err
	}

	val, ok := resp.Payload.Route.Config[key]
	if !ok {
		return fmt.Errorf("Config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (r *routesCmd) listConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))

	resp, err := r.client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   route,
	})

	if err != nil {
		return err
	}

	for key, val := range resp.Payload.Route.Config {
		fmt.Printf("%s=%s\n", key, val)
	}

	return nil
}

func (r *routesCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))
	key := c.Args().Get(2)

	rt := fnmodels.Route{
		Config: make(map[string]string),
	}

	rt.Config[key] = ""

	err := PatchRoute(r.client, appName, route, &rt)
	if err != nil {
		return err
	}

	fmt.Printf("Removed key '%s' from the route '%s%s'", key, appName, key)
	return nil
}

func (r *routesCmd) inspect(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))
	prop := c.Args().Get(2)

	resp, err := r.client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   route,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiroutes.GetAppsAppRoutesRouteNotFound:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(resp.Payload.Route)
		return nil
	}

	data, err := json.Marshal(resp.Payload.Route)
	if err != nil {
		return fmt.Errorf("Failed to inspect route: %s", err)
	}
	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("Failed to inspect route: %s", err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return errors.New("Failed to inspect that route's field")
	}
	enc.Encode(field)

	return nil
}

func (r *routesCmd) delete(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := WithoutSlash(c.Args().Get(1))

	_, err := r.client.Routes.DeleteAppsAppRoutesRoute(&apiroutes.DeleteAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   route,
	})
	if err != nil {
		switch e := err.(type) {
		case *apiroutes.DeleteAppsAppRoutesRouteNotFound:
			return fmt.Errorf("%s", e.Payload.Error.Message)
		default:
			return err
		}
	}

	fmt.Println(appName, route, "deleted")
	return nil
}
