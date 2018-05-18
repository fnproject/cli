package route

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	client "github.com/fnproject/cli/client"
	cmd "github.com/fnproject/cli/commands"
	common "github.com/fnproject/cli/common"
	apiroutes "github.com/fnproject/fn_go/client/routes"
	fnmodels "github.com/fnproject/fn_go/models"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

var routeFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "image,i",
		Usage: "image name",
	},
	cli.Uint64Flag{
		Name:  "memory,m",
		Usage: "memory in MiB",
	},
	cli.StringFlag{
		Name:  "type,t",
		Usage: "route type - sync or async",
	},
	cli.StringSliceFlag{
		Name:  "config,c",
		Usage: "route configuration",
	},
	cli.StringSliceFlag{
		Name:  "headers",
		Usage: "route response headers",
	},
	cli.StringFlag{
		Name:  "format,f",
		Usage: "hot container IO format - default or http",
	},
	cli.IntFlag{
		Name:  "timeout",
		Usage: "route timeout (eg. 30)",
	},
	cli.IntFlag{
		Name:  "idle-timeout",
		Usage: "route idle timeout (eg. 30)",
	},
}

var updateRouteFlags = routeFlags

var callFnFlags = append(runflags(),
	cli.BoolFlag{
		Name:  "display-call-id",
		Usage: "whether display call ID or not",
	},
)

type routeCmd common.FnClient

func routes(command string, apiClient *routeCmd) cli.Command {
	var rCmd cli.Command

	switch command {
	case cmd.CreateCmd:
		rCmd = apiClient.getCreateRouteCommand()
	case cmd.CallCmd:
		rCmd = getCallRoutesCommand()
	case cmd.ListCmd:
		rCmd = apiClient.getListRoutesCommand()
	case cmd.DeleteCmd:
		rCmd = apiClient.getDeleteRouteCommand()
	case cmd.UpdateCmd:
		rCmd = apiClient.getUpdateRouteCommand()
	case cmd.ConfigCmd:
		rCmd = apiClient.getConfigRoutesCommand()
	case cmd.InspectCmd:
		rCmd = apiClient.getInspectRoutesCommand()
	}

	return rCmd
}

func getCallRoutesCommand() cli.Command {
	return cli.Command{
		Name:      "routes",
		Usage:     "call a route",
		ArgsUsage: "<app> </path> [image]",
		Action:    call,
		Flags:     callFnFlags,
	}
}

func (apiClient *routeCmd) getCreateRouteCommand() cli.Command {
	return cli.Command{
		Name:      "route",
		Usage:     "Create a route in an application",
		ArgsUsage: "<app> </path>",
		Action:    apiClient.createRoute,
		Flags:     routeFlags,
	}
}

func (client *routeCmd) getListRoutesCommand() cli.Command {
	return cli.Command{
		Name:      "routes",
		Usage:     "list routes for `app`",
		ArgsUsage: "<app>",
		Action:    client.listRoutes,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of routes to return",
				Value: int64(100),
			},
		},
	}
}

func (apiClient *routeCmd) getDeleteRouteCommand() cli.Command {
	return cli.Command{
		Name:      "route",
		Usage:     "Delete a route from an application `app`",
		ArgsUsage: "<app> </path>",
		Action:    apiClient.deleteRoutes,
	}
}

func (apiClient *routeCmd) getInspectRoutesCommand() cli.Command {
	return cli.Command{
		Name:      "routes",
		Usage:     "retrieve one or all routes properties",
		ArgsUsage: "<app> </path> [property.[key]]",
		Action:    apiClient.inspectRoutes,
	}
}

func (apiClient *routeCmd) getConfigRoutesCommand() cli.Command {
	return cli.Command{
		Name:  "routes",
		Usage: "operate a route configuration set",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "store a configuration key for this route",
				ArgsUsage: "<app> </path> <key> <value>",
				Action:    apiClient.configSetRoutes,
			},
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "inspect configuration key for this route",
				ArgsUsage: "<app> </path> <key>",
				Action:    apiClient.configGetRoutes,
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list configuration key/value pairs for this route",
				ArgsUsage: "<app> </path>",
				Action:    apiClient.configListRoutes,
			},
			{
				Name:      "unset",
				Aliases:   []string{"u"},
				Usage:     "remove a configuration key for this route",
				ArgsUsage: "<app> </path> <key>",
				Action:    apiClient.configUnsetRoutes,
			},
		},
	}
}

func (apiClient *routeCmd) getUpdateRouteCommand() cli.Command {
	return cli.Command{
		Name:      "route",
		Aliases:   []string{"u"},
		Usage:     "Update a Route in an `app`",
		ArgsUsage: "<app> </path>",
		Action:    apiClient.updateRoutes,
		Flags:     updateRouteFlags,
	}
}

func call() cli.Command {
	apiClient := routeCmd{}

	return cli.Command{
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Name:      "call",
		Usage:     "call a remote function",
		ArgsUsage: "<app> </path>",
		Flags:     callFnFlags,
		Action:    apiClient.call,
	}
}

func cleanRoutePath(p string) string {
	p = path.Clean(p)
	if !path.IsAbs(p) {
		p = "/" + p
	}
	return p
}

func printRoutes(appName string, routes []*fnmodels.Route) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "path", "\t", "image", "\t", "endpoint", "\n")
	for _, route := range routes {
		endpoint := path.Join(client.Host(), "r", appName, route.Path)
		fmt.Fprint(w, route.Path, "\t", route.Image, "\t", endpoint, "\n")
	}
	w.Flush()
}

func (apiClient *routeCmd) listRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)

	params := &apiroutes.GetAppsAppRoutesParams{
		Context: context.Background(),
		App:     appName,
	}

	var resRoutes []*fnmodels.Route
	for {
		resp, err := apiClient.Client.Routes.GetAppsAppRoutes(params)

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
			return errors.New("number of calls: negative value not allowed")
		}

		resRoutes = append(resRoutes, resp.Payload.Routes...)
		howManyMore := n - int64(len(resRoutes)+len(resp.Payload.Routes))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	printRoutes(appName, resRoutes)
	return nil
}

func (apiClient *routeCmd) call(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))

	u := url.URL{
		Scheme: "http",
		Host:   client.Host(),
	}
	u.Path = path.Join(u.Path, "r", appName, route)
	content := stdin()

	return client.CallFN(u.String(), content, os.Stdout, c.String("method"), c.StringSlice("e"), c.String("content-type"), c.Bool("display-call-id"))
}

func routeWithFlags(c *cli.Context, rt *fnmodels.Route) {
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
			rt.Config = extractEnvConfig(c.StringSlice("config"))
		}
	}
}

func routeWithFuncFile(ff *common.FuncFile, rt *fnmodels.Route) error {
	var err error
	if ff == nil {
		_, ff, err = loadFuncfile()
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

	return nil
}

func (apiClient *routeCmd) createRoute(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))

	rt := &fnmodels.Route{}
	rt.Path = route
	rt.Image = c.Args().Get(2)

	routeWithFlags(c, rt)

	if rt.Path == "" {
		return errors.New("route path is missing")
	}
	if rt.Image == "" {
		return errors.New("no image specified")
	}

	return apiClient.postRoute(c, appName, rt)
}

func (apiClient *routeCmd) postRoute(c *cli.Context, appName string, rt *fnmodels.Route) error {

	err := validateImageName(rt.Image)
	if err != nil {
		return err
	}

	body := &fnmodels.RouteWrapper{
		Route: rt,
	}

	resp, err := apiClient.Client.Routes.PostAppsAppRoutes(&apiroutes.PostAppsAppRoutesParams{
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

func (apiClient *routeCmd) patchRoute(c *cli.Context, appName, routePath string, r *fnmodels.Route) error {
	if r.Image != "" {
		err := validateImageName(r.Image)
		if err != nil {
			return err
		}
	}

	_, err := apiClient.Client.Routes.PatchAppsAppRoutesRoute(&apiroutes.PatchAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   routePath,
		Body:    &fnmodels.RouteWrapper{Route: r},
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

func (apiClient *routeCmd) putRoute(c *cli.Context, appName, routePath string, r *fnmodels.Route) error {
	_, err := apiClient.Client.Routes.PutAppsAppRoutesRoute(&apiroutes.PutAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   routePath,
		Body:    &fnmodels.RouteWrapper{Route: r},
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

func (apiClient *routeCmd) updateRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))

	rt := &fnmodels.Route{}

	routeWithFlags(c, rt)

	err := apiClient.patchRoute(c, appName, route, rt)
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "updated")
	return nil
}

func (apiClient *routeCmd) configSetRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))
	key := c.Args().Get(2)
	value := c.Args().Get(3)

	patchRoute := fnmodels.Route{
		Config: make(map[string]string),
	}

	patchRoute.Config[key] = value

	err := apiClient.patchRoute(c, appName, route, &patchRoute)
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "updated", key, "with", value)
	return nil
}

func (apiClient *routeCmd) configGetRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))
	key := c.Args().Get(2)

	resp, err := apiClient.Client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
		Context: context.Background(),
		App:     appName,
		Route:   route,
	})

	if err != nil {
		return err
	}

	val, ok := resp.Payload.Route.Config[key]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (apiClient *routeCmd) configListRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))

	resp, err := apiClient.Client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
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

func (apiClient *routeCmd) configUnsetRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))
	key := c.Args().Get(2)

	patchRoute := fnmodels.Route{
		Config: make(map[string]string),
	}

	patchRoute.Config[key] = ""

	err := apiClient.patchRoute(c, appName, route, &patchRoute)
	if err != nil {
		return err
	}

	fmt.Printf("removed key '%s' from the route '%s%s'", key, appName, key)
	return nil
}

func (apiClient *routeCmd) inspectRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))
	prop := c.Args().Get(2)

	resp, err := apiClient.Client.Routes.GetAppsAppRoutesRoute(&apiroutes.GetAppsAppRoutesRouteParams{
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
		return fmt.Errorf("failed to inspect route: %s", err)
	}
	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("failed to inspect route: %s", err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return errors.New("failed to inspect that route's field")
	}
	enc.Encode(field)

	return nil
}

func (apiClient *routeCmd) deleteRoutes(c *cli.Context) error {
	appName := c.Args().Get(0)
	route := cleanRoutePath(c.Args().Get(1))

	_, err := apiClient.Client.Routes.DeleteAppsAppRoutesRoute(&apiroutes.DeleteAppsAppRoutesRouteParams{
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
