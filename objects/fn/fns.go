/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	client "github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/jmoiron/jsonq"
	ociCommon "github.com/oracle/oci-go-sdk/v65/common"
	ociFunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

type fnsCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
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
	cli.BoolFlag{
		Name:  "code-only",
		Usage: "Create a code-only function using archive source details and runtime configuration",
	},
	cli.StringFlag{
		Name:  "source-type",
		Usage: "Code-only source type: direct or object-storage",
	},
	cli.StringFlag{
		Name:  "source-file",
		Usage: "Path to a zip archive for direct code-only source upload",
	},
	cli.StringFlag{
		Name:  "bucket-name",
		Usage: "Object Storage bucket name for code-only source",
	},
	cli.StringFlag{
		Name:  "namespace",
		Usage: "Object Storage namespace for code-only source",
	},
	cli.StringFlag{
		Name:  "object-name",
		Usage: "Object Storage object name for code-only source",
	},
	cli.StringFlag{
		Name:  "object-version-id",
		Usage: "Object Storage object version id for code-only source",
	},
	cli.StringFlag{
		Name:  "runtime-config-type",
		Usage: "Runtime configuration type for code-only creation: function-update or manual",
	},
	cli.StringFlag{
		Name:  "runtime-name",
		Usage: "Runtime name for code-only creation",
	},
	cli.StringFlag{
		Name:  "runtime-version-id",
		Usage: "Runtime version OCID for manual runtime configuration",
	},
	cli.StringFlag{
		Name:  "handler",
		Usage: "Handler for code-only archive functions",
	},
}
var updateFnFlags = FnFlags

type codeOnlyUpdateOptions struct {
	codeOnly          bool
	sourceType        string
	sourceFile        string
	bucketName        string
	namespace         string
	objectName        string
	objectVersionID   string
	runtimeConfigType string
	runtimeName       string
	runtimeVersionID  string
	handler           string
}

type codeOnlyCreateOptions struct {
	codeOnly          bool
	sourceType        string
	sourceFile        string
	bucketName        string
	namespace         string
	objectName        string
	objectVersionID   string
	runtimeConfigType string
	runtimeName       string
	runtimeVersionID  string
	handler           string
}

func readCodeOnlyCreateOptions(c *cli.Context) codeOnlyCreateOptions {
	return codeOnlyCreateOptions{
		codeOnly:          c.Bool("code-only"),
		sourceType:        strings.TrimSpace(c.String("source-type")),
		sourceFile:        strings.TrimSpace(c.String("source-file")),
		bucketName:        strings.TrimSpace(c.String("bucket-name")),
		namespace:         strings.TrimSpace(c.String("namespace")),
		objectName:        strings.TrimSpace(c.String("object-name")),
		objectVersionID:   strings.TrimSpace(c.String("object-version-id")),
		runtimeConfigType: strings.TrimSpace(c.String("runtime-config-type")),
		runtimeName:       strings.TrimSpace(c.String("runtime-name")),
		runtimeVersionID:  strings.TrimSpace(c.String("runtime-version-id")),
		handler:           strings.TrimSpace(c.String("handler")),
	}
}

func (o codeOnlyCreateOptions) enabled() bool {
	return o.codeOnly || o.sourceType != "" || o.sourceFile != "" || o.bucketName != "" || o.namespace != "" || o.objectName != "" || o.objectVersionID != "" || o.runtimeConfigType != "" || o.runtimeName != "" || o.runtimeVersionID != "" || o.handler != ""
}

func readCodeOnlyUpdateOptions(c *cli.Context) codeOnlyUpdateOptions {
	return codeOnlyUpdateOptions{
		codeOnly:          c.Bool("code-only"),
		sourceType:        strings.TrimSpace(c.String("source-type")),
		sourceFile:        strings.TrimSpace(c.String("source-file")),
		bucketName:        strings.TrimSpace(c.String("bucket-name")),
		namespace:         strings.TrimSpace(c.String("namespace")),
		objectName:        strings.TrimSpace(c.String("object-name")),
		objectVersionID:   strings.TrimSpace(c.String("object-version-id")),
		runtimeConfigType: strings.TrimSpace(c.String("runtime-config-type")),
		runtimeName:       strings.TrimSpace(c.String("runtime-name")),
		runtimeVersionID:  strings.TrimSpace(c.String("runtime-version-id")),
		handler:           strings.TrimSpace(c.String("handler")),
	}
}

func (o codeOnlyUpdateOptions) enabled() bool {
	return o.codeOnly || o.sourceType != "" || o.sourceFile != "" || o.bucketName != "" || o.namespace != "" || o.objectName != "" || o.objectVersionID != "" || o.runtimeConfigType != "" || o.runtimeName != "" || o.runtimeVersionID != "" || o.handler != ""
}

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

func printFunctions(c *cli.Context, fns []*models.Fn) error {
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
	resFns, err := getFns(c, f.client)
	if err != nil {
		return err
	}
	return printFunctions(c, resFns)
}

func getFns(c *cli.Context, client *fnclient.Fn) ([]*modelsv2.Fn, error) {
	appName := c.Args().Get(0)

	a, err := app.GetAppByName(client, appName)
	if err != nil {
		return nil, err
	}
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &a.ID,
	}

	var resFns []*models.Fn
	for {
		resp, err := client.Fns.ListFns(params)

		if err != nil {
			return nil, err
		}
		n := c.Int64("n")

		resFns = append(resFns, resp.Payload.Items...)
		howManyMore := n - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

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
	provider, err := client.CurrentProvider()
	if err != nil {
		return
	}
	resp, err := getFns(c, provider.APIClientv2())
	if err != nil {
		return
	}
	for _, f := range resp {
		fmt.Println(f.Name)
	}
}

func getFnByAppAndFnName(appName, fnName string) (*models.Fn, error) {
	provider, err := client.CurrentProvider()
	if err != nil {
		return nil, errors.New("could not get context")
	}
	app, err := app.GetAppByName(provider.APIClientv2(), appName)
	if err != nil {
		return nil, fmt.Errorf("could not get app %v", appName)
	}
	fn, err := GetFnByName(provider.APIClientv2(), app.ID, fnName)
	if err != nil {
		return nil, fmt.Errorf("could not get function %v", fnName)
	}
	return fn, nil
}

// WithFlags returns a function with specified flags
func WithFlags(c *cli.Context, fn *models.Fn) {
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
		fn.IdleTimeout = &to
	}
}

// WithFuncFileV20180708 used when creating a function from a funcfile
func WithFuncFileV20180708(ff *common.FuncFileV20180708, fn *models.Fn) error {
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
		fn.IdleTimeout = ff.IDLE_timeout
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
	codeOnly := readCodeOnlyCreateOptions(c)

	fn := &models.Fn{}
	fn.Name = fnName
	fn.Image = c.Args().Get(2)

	WithFlags(c, fn)

	if fn.Name == "" {
		return errors.New("fnName path is missing")
	}
	if codeOnly.enabled() {
		if err := applyCodeOnlyCreateOptions(f.provider, fn, codeOnly); err != nil {
			return err
		}
	} else if fn.Image == "" {
		return errors.New("no image specified")
	}

	a, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}

	_, err = CreateFn(f.client, a.ID, fn)
	return err
}

// CreateFn request
func CreateFn(r *fnclient.Fn, appID string, fn *models.Fn) (*models.Fn, error) {
	fn.AppID = appID
	if fn.Image != "" {
		err := common.ValidateTagImageName(fn.Image)
		if err != nil {
			return nil, err
		}
	}

	resp, err := r.Fns.CreateFn(&apifns.CreateFnParams{
		Context: context.Background(),
		Body:    fn,
	})

	if err != nil {
		switch e := err.(type) {
		case *apifns.CreateFnBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		case *apifns.CreateFnConflict:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}

	if fn.CodeOnly || resp.Payload.Image == "" {
		fmt.Println("Successfully created code-only function:", resp.Payload.Name)
	} else {
		fmt.Println("Successfully created function:", resp.Payload.Name, "with", resp.Payload.Image)
	}
	return resp.Payload, nil
}

// PutFn updates the fn with the given ID using the content of the provided fn
func PutFn(f *fnclient.Fn, fnID string, fn *models.Fn) error {
	if fn.Image != "" {
		err := common.ValidateTagImageName(fn.Image)
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

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("function %s not found", n.Name)
}

// GetFnByName looks up a fn by name using the given client
func GetFnByName(client *fnclient.Fn, appID, fnName string) (*models.Fn, error) {
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
		return nil, NameNotFoundError{fnName}
	}

	return fn, nil
}

func (f *fnsCmd) update(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	codeOnly := readCodeOnlyUpdateOptions(c)

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	WithFlags(c, fn)
	if codeOnly.enabled() {
		if err := applyCodeOnlyUpdateOptions(f.provider, fn, codeOnly); err != nil {
			return err
		}
	}

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

	app, err := app.GetAppByName(f.client, appName)
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

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}
	_, ok := fn.Config[key]
	if !ok {
		fmt.Printf("Config key '%s' does not exist. Nothing to do.\n", key)
		return nil
	}
	fn.Config[key] = ""

	err = PutFn(f.client, fn.ID, fn)
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

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
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

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	//recursive delete of sub-objects
	if c.Bool("recursive") {
		triggers, err := common.ListTriggersInFunc(c, f.client, fn)
		if err != nil {
			return fmt.Errorf("Failed to get associated objects: %s", err)
		}

		//Forced delete
		var shouldContinue bool
		if c.Bool("force") {
			shouldContinue = true
		} else {
			shouldContinue = common.UserConfirmedMultiResourceDeletion(nil, []*modelsv2.Fn{fn}, triggers)
		}

		if shouldContinue {
			err := common.DeleteTriggers(c, f.client, triggers)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
		} else {
			return nil
		}
	}

	params := apifns.NewDeleteFnParams()
	params.FnID = fn.ID
	_, err = f.client.Fns.DeleteFn(params)

	if err != nil {
		return err
	}

	fmt.Println("Function", fnName, "deleted")
	return nil
}

func applyCodeOnlyCreateOptions(p provider.Provider, fn *models.Fn, opts codeOnlyCreateOptions) error {
	if fn.Image != "" {
		return fmt.Errorf("Specify either an image or --code-only options, not both")
	}
	if !opts.codeOnly {
		return fmt.Errorf("--code-only is required when specifying code-only source or runtime flags")
	}
	sourceType, err := normalizeSourceType(opts.sourceType)
	if err != nil {
		return err
	}
	mode, err := normalizeRuntimeConfigType(opts.runtimeConfigType)
	if err != nil {
		return err
	}
	if sourceType == "" {
		return fmt.Errorf("--source-type is required for code-only create")
	}
	if mode == "" {
		return fmt.Errorf("--runtime-config-type is required for code-only create")
	}
	if opts.runtimeName == "" {
		return fmt.Errorf("--runtime-name is required for code-only create")
	}
	if requiresHandlerForRuntime(opts.runtimeName) && opts.handler == "" {
		return fmt.Errorf("--handler is required for runtime %s", opts.runtimeName)
	}
	if err := validateHandlerForRuntime(opts.runtimeName, opts.handler); err != nil {
		return err
	}
	if err := validateCodeOnlySourceOptions(sourceType, opts); err != nil {
		return err
	}
	if err := validateRuntimeConfig(p, mode, opts.runtimeName, opts.runtimeVersionID); err != nil {
		return err
	}

	fn.CodeOnly = true
	fn.Image = ""
	fn.SourceType = sourceType
	fn.SourceFile = opts.sourceFile
	fn.SourceBucketName = opts.bucketName
	fn.SourceNamespace = opts.namespace
	fn.SourceObjectName = opts.objectName
	fn.SourceObjectVersion = opts.objectVersionID
	fn.RuntimeConfigType = mode
	fn.RuntimeName = opts.runtimeName
	fn.RuntimeVersionID = opts.runtimeVersionID
	fn.Handler = opts.handler

	if sourceType == "direct" {
		archive, err := os.ReadFile(opts.sourceFile)
		if err != nil {
			return fmt.Errorf("failed to read --source-file %s: %w", opts.sourceFile, err)
		}
		fn.SourceArchive = archive
	}

	return nil
}

func normalizeSourceType(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "":
		return "", nil
	case "direct":
		return "direct", nil
	case "object-storage", "object_storage", "objectstorage":
		return "object-storage", nil
	default:
		return "", fmt.Errorf("unsupported --source-type %q. Supported values are direct and object-storage", value)
	}
}

func normalizeRuntimeConfigType(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "":
		return "", nil
	case "function-update", "function_update":
		return "FUNCTION_UPDATE", nil
	case "manual":
		return "MANUAL", nil
	default:
		return "", fmt.Errorf("unsupported --runtime-config-type %q. Supported values are function-update and manual", value)
	}
}

func validateCodeOnlySourceOptions(sourceType string, opts codeOnlyCreateOptions) error {
	switch sourceType {
	case "direct":
		if opts.sourceFile == "" {
			return fmt.Errorf("--source-file is required when --source-type=direct")
		}
		if opts.bucketName != "" || opts.namespace != "" || opts.objectName != "" || opts.objectVersionID != "" {
			return fmt.Errorf("Object Storage flags cannot be used when --source-type=direct")
		}
	case "object-storage":
		if opts.bucketName == "" || opts.namespace == "" || opts.objectName == "" {
			return fmt.Errorf("--bucket-name, --namespace, and --object-name are required when --source-type=object-storage")
		}
		if opts.sourceFile != "" {
			return fmt.Errorf("--source-file cannot be used when --source-type=object-storage")
		}
	}
	return nil
}

func validateRuntimeConfig(p provider.Provider, mode, runtimeName, runtimeVersionID string) error {
	if strings.TrimSpace(runtimeName) != "" {
		if err := validateRuntimeName(p, runtimeName); err != nil {
			return err
		}
	}
	switch mode {
	case "FUNCTION_UPDATE":
		if runtimeVersionID != "" {
			return fmt.Errorf("--runtime-version-id is only valid for manual runtime configuration")
		}
	case "MANUAL":
		if runtimeVersionID == "" {
			return fmt.Errorf("--runtime-version-id is required when --runtime-config-type=manual")
		}
		if err := validateRuntimeVersionMatchesRuntime(p, runtimeName, runtimeVersionID); err != nil {
			return err
		}
	}
	return nil
}

func validateRuntimeName(p provider.Provider, runtimeName string) error {
	ociProvider, ok := p.(*oracle.OracleProvider)
	if !ok || ociProvider == nil {
		return fmt.Errorf("runtime validation requires an oracle provider")
	}
	client, err := ociFunctions.NewFunctionsManagementClientWithConfigurationProvider(ociProvider.ConfigurationProvider)
	if err != nil {
		return err
	}
	if ociProvider.FnApiUrl != nil {
		client.Host = ociProvider.FnApiUrl.String()
	} else {
		region, err := ociProvider.ConfigurationProvider.Region()
		if err != nil {
			return err
		}
		client.SetRegion(region)
	}
	request := ociFunctions.ListFunctionsRuntimesRequest{}
	for {
		response, err := client.ListFunctionsRuntimes(context.Background(), request)
		if err != nil {
			return err
		}
		for _, item := range response.Items {
			if item.Name == nil || strings.TrimSpace(*item.Name) != runtimeName {
				continue
			}
			if item.LifecycleState != ociFunctions.FunctionsRuntimeLifecycleStateActive {
				return fmt.Errorf("runtime %s is not active", runtimeName)
			}
			return nil
		}
		if response.OpcNextPage == nil {
			break
		}
		request.Page = response.OpcNextPage
	}
	return fmt.Errorf("runtime %s does not exist", runtimeName)
}

func validateRuntimeVersionMatchesRuntime(p provider.Provider, runtimeName, runtimeVersionID string) error {
	ociProvider, ok := p.(*oracle.OracleProvider)
	if !ok || ociProvider == nil {
		return fmt.Errorf("runtime version validation requires an oracle provider")
	}
	client, err := ociFunctions.NewFunctionsManagementClientWithConfigurationProvider(ociProvider.ConfigurationProvider)
	if err != nil {
		return err
	}
	if ociProvider.FnApiUrl != nil {
		client.Host = ociProvider.FnApiUrl.String()
	} else {
		region, err := ociProvider.ConfigurationProvider.Region()
		if err != nil {
			return err
		}
		client.SetRegion(region)
	}
	request := ociFunctions.ListFunctionsRuntimeVersionsRequest{
		FunctionsRuntimeName:      &runtimeName,
		FunctionsRuntimeVersionId: &runtimeVersionID,
		Limit:                     ociCommon.Int(1),
	}
	response, err := client.ListFunctionsRuntimeVersions(context.Background(), request)
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		return fmt.Errorf("runtime version %s does not belong to runtime %s", runtimeVersionID, runtimeName)
	}
	if response.Items[0].LifecycleState != ociFunctions.FunctionsRuntimeVersionLifecycleStateActive {
		return fmt.Errorf("runtime version %s is not active for runtime %s", runtimeVersionID, runtimeName)
	}
	return nil
}

func requiresHandlerForRuntime(runtimeName string) bool {
	baseRuntime := strings.ToLower(strings.TrimSpace(runtimeName))
	for _, sep := range []string{".", "-"} {
		if idx := strings.Index(baseRuntime, sep); idx != -1 {
			baseRuntime = baseRuntime[:idx]
			break
		}
	}
	return strings.HasPrefix(baseRuntime, "java") || strings.HasPrefix(baseRuntime, "python") || strings.HasPrefix(baseRuntime, "node") || strings.HasPrefix(baseRuntime, "javascript")
}

func validateHandlerForRuntime(runtimeName, handler string) error {
	baseRuntime := strings.ToLower(strings.TrimSpace(runtimeName))
	for _, sep := range []string{".", "-"} {
		if idx := strings.Index(baseRuntime, sep); idx != -1 {
			baseRuntime = baseRuntime[:idx]
			break
		}
	}
	h := strings.TrimSpace(handler)
	if h == "" {
		return nil
	}
	switch {
	case strings.HasPrefix(baseRuntime, "python"):
		if strings.Contains(h, ":") || strings.Count(h, ".") != 1 {
			return fmt.Errorf("handler for runtime %s must be in the format <fileName>.<function>", runtimeName)
		}
	case strings.HasPrefix(baseRuntime, "java"):
		if !strings.Contains(h, "::") {
			return fmt.Errorf("handler for runtime %s must be in the format <class>::<method>", runtimeName)
		}
	case strings.HasPrefix(baseRuntime, "node"), strings.HasPrefix(baseRuntime, "javascript"):
		if strings.Contains(h, ":") || strings.Count(h, ".") != 1 {
			return fmt.Errorf("handler for runtime %s must be in the format <fileName>.<function>", runtimeName)
		}
	}
	return nil
}

func applyCodeOnlyUpdateOptions(p provider.Provider, fn *models.Fn, opts codeOnlyUpdateOptions) error {
	if fn.Image != "" {
		return fmt.Errorf("Specify either an image update or code-only update flags, not both")
	}
	if !opts.codeOnly {
		return fmt.Errorf("--code-only is required when specifying code-only update flags")
	}

	sourceType, err := normalizeSourceType(opts.sourceType)
	if err != nil {
		return err
	}
	mode, err := normalizeRuntimeConfigType(opts.runtimeConfigType)
	if err != nil {
		return err
	}

	if sourceType != "" {
		if err := validateCodeOnlySourceOptions(sourceType, codeOnlyCreateOptions{
			codeOnly:        true,
			sourceType:      sourceType,
			sourceFile:      opts.sourceFile,
			bucketName:      opts.bucketName,
			namespace:       opts.namespace,
			objectName:      opts.objectName,
			objectVersionID: opts.objectVersionID,
		}); err != nil {
			return err
		}
	}

	if mode != "" {
		if opts.runtimeName == "" {
			return fmt.Errorf("--runtime-name is required when changing --runtime-config-type")
		}
		if err := validateRuntimeConfig(p, mode, opts.runtimeName, opts.runtimeVersionID); err != nil {
			return err
		}
	}

	effectiveRuntimeName := opts.runtimeName
	if effectiveRuntimeName == "" {
		effectiveRuntimeName = fn.RuntimeName
	}
	if effectiveRuntimeName != "" && requiresHandlerForRuntime(effectiveRuntimeName) {
		effectiveHandler := strings.TrimSpace(opts.handler)
		if effectiveHandler == "" {
			effectiveHandler = strings.TrimSpace(fn.Handler)
		}
		if effectiveHandler == "" && (sourceType != "" || mode != "") {
			return fmt.Errorf("--handler is required for runtime %s", effectiveRuntimeName)
		}
		if err := validateHandlerForRuntime(effectiveRuntimeName, effectiveHandler); err != nil {
			return err
		}
	}

	fn.CodeOnly = true
	fn.Image = ""
	if sourceType != "" {
		fn.SourceType = sourceType
		fn.SourceFile = opts.sourceFile
		fn.SourceBucketName = opts.bucketName
		fn.SourceNamespace = opts.namespace
		fn.SourceObjectName = opts.objectName
		fn.SourceObjectVersion = opts.objectVersionID
		if sourceType == "direct" {
			archive, err := os.ReadFile(opts.sourceFile)
			if err != nil {
				return fmt.Errorf("failed to read --source-file %s: %w", opts.sourceFile, err)
			}
			fn.SourceArchive = archive
		} else {
			fn.SourceArchive = nil
		}
	}
	if mode != "" {
		fn.RuntimeConfigType = mode
		fn.RuntimeName = opts.runtimeName
		fn.RuntimeVersionID = opts.runtimeVersionID
	}
	if opts.handler != "" {
		fn.Handler = opts.handler
	}

	if sourceType == "" && mode == "" && opts.handler == "" {
		return fmt.Errorf("no code-only update fields were provided")
	}

	return nil
}
