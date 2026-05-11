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
	"time"

	client "github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	fnprovideroracle "github.com/fnproject/fn_go/provider/oracle"
	"github.com/jmoiron/jsonq"
	ocifunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

type fnsCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
}

const (
	annotationProvisionedConcurrencyStrategy = "oracle.com/oci/provisionedConcurrencyStrategy"
	annotationProvisionedConcurrencyCount    = "oracle.com/oci/provisionedConcurrencyCount"
)

const annotationDetachedTimeoutSeconds = "oracle.com/oci/detachedModeTimeoutInSeconds"

const (
	annotationSuccessDestinationKind = "oracle.com/oci/successDestinationKind"
	annotationSuccessDestinationOCID = "oracle.com/oci/successDestinationOcid"
	annotationFailureDestinationKind = "oracle.com/oci/failureDestinationKind"
	annotationFailureDestinationOCID = "oracle.com/oci/failureDestinationOcid"
)

type detachedDestinationView struct {
	Type string `json:"type,omitempty"`
	OCID string `json:"ocid,omitempty"`
}

type detachedModeView struct {
	Timeout   string                   `json:"timeout,omitempty"`
	OnSuccess *detachedDestinationView `json:"onSuccess,omitempty"`
	OnFailure *detachedDestinationView `json:"onFailure,omitempty"`
}

type provisionedConcurrencyView struct {
	Strategy string `json:"strategy"`
	Count    *int   `json:"count,omitempty"`
}

// SetProvisionedConcurrencyAnnotations adds the internal annotations used to
// carry provisioned concurrency through the create payload into the OCI shim.
func SetProvisionedConcurrencyAnnotations(fn *models.Fn, cfg *common.OCIProvisionedConcurrencyConfig) error {
	if fn == nil || cfg == nil {
		return nil
	}
	if err := common.ValidateProvisionedConcurrencyConfig(cfg); err != nil {
		return err
	}
	if fn.Annotations == nil {
		fn.Annotations = make(map[string]interface{})
	}
	strategy := strings.ToUpper(strings.TrimSpace(cfg.Strategy))
	fn.Annotations[annotationProvisionedConcurrencyStrategy] = strategy
	if strategy == common.ProvisionedConcurrencyStrategyConstant && cfg.Count != nil {
		fn.Annotations[annotationProvisionedConcurrencyCount] = *cfg.Count
	} else {
		delete(fn.Annotations, annotationProvisionedConcurrencyCount)
	}
	return nil
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
	cli.StringFlag{
		Name:  "provisioned-concurrency",
		Usage: "Set OCI provisioned concurrency using 'none' or 'constant:<count>'",
	},
	cli.StringFlag{
		Name:  "detached-timeout",
		Usage: "Set OCI detached mode timeout using a duration like 20m or 1h",
	},
	cli.StringFlag{
		Name:  "on-success",
		Usage: "Set OCI detached success destination using <stream|queue|notifications>:<ocid>",
	},
	cli.StringFlag{
		Name:  "on-failure",
		Usage: "Set OCI detached failure destination using <stream|queue|notifications>:<ocid>",
	},
}
var updateFnFlags = append(append([]cli.Flag{}, FnFlags...),
	cli.BoolFlag{
		Name:  "clear-on-success",
		Usage: "Clear OCI detached success destination",
	},
	cli.BoolFlag{
		Name:  "clear-on-failure",
		Usage: "Clear OCI detached failure destination",
	},
)

type clearDestinationRequest struct {
	Success bool
	Failure bool
}

func warnUnsupportedDetachedTimeout() {
	fmt.Fprintln(os.Stderr, "Warning: --detached-timeout is only supported with an oracle provider and will be ignored.")
}

func SetDetachedTimeoutAnnotation(fn *models.Fn, seconds int) {
	if fn == nil || seconds <= 0 {
		return
	}
	if fn.Annotations == nil {
		fn.Annotations = make(map[string]interface{})
	}
	fn.Annotations[annotationDetachedTimeoutSeconds] = seconds
}

func warnUnsupportedDestination(flagName string) {
	fmt.Fprintf(os.Stderr, "Warning: %s is only supported with an oracle provider and will be ignored.\n", flagName)
}

func validateDestinationFlagCombination(c *cli.Context) (*clearDestinationRequest, error) {
	clearReq := &clearDestinationRequest{
		Success: c.Bool("clear-on-success"),
		Failure: c.Bool("clear-on-failure"),
	}
	if clearReq.Success && strings.TrimSpace(c.String("on-success")) != "" {
		return nil, fmt.Errorf("--on-success and --clear-on-success cannot be used together")
	}
	if clearReq.Failure && strings.TrimSpace(c.String("on-failure")) != "" {
		return nil, fmt.Errorf("--on-failure and --clear-on-failure cannot be used together")
	}
	return clearReq, nil
}

func SetDestinationAnnotations(fn *models.Fn, onSuccess, onFailure *common.OCIDestination) {
	if fn == nil {
		return
	}
	if fn.Annotations == nil {
		fn.Annotations = make(map[string]interface{})
	}
	if onSuccess != nil {
		fn.Annotations[annotationSuccessDestinationKind] = strings.ToUpper(onSuccess.Type)
		fn.Annotations[annotationSuccessDestinationOCID] = onSuccess.OCID
	}
	if onFailure != nil {
		fn.Annotations[annotationFailureDestinationKind] = strings.ToUpper(onFailure.Type)
		fn.Annotations[annotationFailureDestinationOCID] = onFailure.OCID
	}
}

func SetClearDestinationAnnotations(fn *models.Fn, clearSuccess, clearFailure bool) {
	if fn == nil || (!clearSuccess && !clearFailure) {
		return
	}
	if fn.Annotations == nil {
		fn.Annotations = make(map[string]interface{})
	}
	if clearSuccess {
		fn.Annotations[annotationSuccessDestinationKind] = "NONE"
		fn.Annotations[annotationSuccessDestinationOCID] = ""
	}
	if clearFailure {
		fn.Annotations[annotationFailureDestinationKind] = "NONE"
		fn.Annotations[annotationFailureDestinationOCID] = ""
	}
}

func formatDetachedTimeout(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	d := time.Duration(seconds) * time.Second
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return fmt.Sprintf("%ds", seconds)
}

func parseDetachedTimeoutFromAnnotations(fn *models.Fn) string {
	if fn == nil || fn.Annotations == nil {
		return ""
	}
	raw, ok := fn.Annotations[annotationDetachedTimeoutSeconds]
	if !ok {
		return ""
	}
	switch typed := raw.(type) {
	case int:
		return formatDetachedTimeout(typed)
	case int32:
		return formatDetachedTimeout(int(typed))
	case int64:
		return formatDetachedTimeout(int(typed))
	case float64:
		return formatDetachedTimeout(int(typed))
	default:
		return ""
	}
}

func parseDetachedDestinationFromAnnotations(fn *models.Fn, kindKey, ocidKey string) *detachedDestinationView {
	if fn == nil || fn.Annotations == nil {
		return nil
	}
	kindRaw, ok := fn.Annotations[kindKey]
	if !ok {
		return nil
	}
	kind, ok := kindRaw.(string)
	if !ok || strings.TrimSpace(kind) == "" {
		return nil
	}
	ocidRaw, ok := fn.Annotations[ocidKey]
	if !ok {
		return &detachedDestinationView{Type: strings.ToLower(kind)}
	}
	ocid, ok := ocidRaw.(string)
	if !ok {
		return &detachedDestinationView{Type: strings.ToLower(kind)}
	}
	return &detachedDestinationView{Type: strings.ToLower(kind), OCID: ocid}
}

func getDetachedModeView(fn *models.Fn) *detachedModeView {
	timeout := parseDetachedTimeoutFromAnnotations(fn)
	onSuccess := parseDetachedDestinationFromAnnotations(fn, annotationSuccessDestinationKind, annotationSuccessDestinationOCID)
	onFailure := parseDetachedDestinationFromAnnotations(fn, annotationFailureDestinationKind, annotationFailureDestinationOCID)
	if timeout == "" && onSuccess == nil && onFailure == nil {
		return nil
	}
	return &detachedModeView{Timeout: timeout, OnSuccess: onSuccess, OnFailure: onFailure}
}

func formatDetachedDestination(view *detachedDestinationView) string {
	if view == nil {
		return ""
	}
	if view.OCID == "" {
		return view.Type
	}
	return fmt.Sprintf("%s:%s", view.Type, view.OCID)
}

func formatDetachedDestinations(fn *models.Fn) string {
	view := getDetachedModeView(fn)
	if view == nil {
		return ""
	}
	parts := []string{}
	if view.OnSuccess != nil {
		parts = append(parts, "success="+formatDetachedDestination(view.OnSuccess))
	}
	if view.OnFailure != nil {
		parts = append(parts, "failure="+formatDetachedDestination(view.OnFailure))
	}
	return strings.Join(parts, ",")
}

func buildInspectFnMap(fn *models.Fn) (map[string]interface{}, error) {
	data, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}
	inspect := map[string]interface{}{}
	if err := json.Unmarshal(data, &inspect); err != nil {
		return nil, err
	}
	if detached := getDetachedModeView(fn); detached != nil {
		detachedData, err := json.Marshal(detached)
		if err != nil {
			return nil, err
		}
		var detachedValue map[string]interface{}
		if err := json.Unmarshal(detachedData, &detachedValue); err != nil {
			return nil, err
		}
		inspect["detachedMode"] = detachedValue
	}
	if pc := getProvisionedConcurrencyView(fn); pc != nil {
		pcData, err := json.Marshal(pc)
		if err != nil {
			return nil, err
		}
		var pcValue map[string]interface{}
		if err := json.Unmarshal(pcData, &pcValue); err != nil {
			return nil, err
		}
		inspect["provisionedConcurrency"] = pcValue
	}
	return inspect, nil
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
				Name                   string                      `json:"name"`
				Image                  string                      `json:"image"`
				ID                     string                      `json:"id"`
				ProvisionedConcurrency *provisionedConcurrencyView `json:"provisionedConcurrency,omitempty"`
				DetachedMode           *detachedModeView           `json:"detachedMode,omitempty"`
			}{
				Name:                   fn.Name,
				Image:                  fn.Image,
				ID:                     fn.ID,
				ProvisionedConcurrency: getProvisionedConcurrencyView(fn),
				DetachedMode:           getDetachedModeView(fn),
			})
		}
		b, err := json.MarshalIndent(newFns, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "NAME", "\t", "IMAGE", "\t", "PC", "\t", "DETACHED_TIMEOUT", "\t", "DESTINATIONS", "\t", "ID", "\n")

		for _, f := range fns {
			view := getDetachedModeView(f)
			timeout := ""
			if view != nil {
				timeout = view.Timeout
			}
			fmt.Fprint(w, f.Name, "\t", f.Image, "\t", formatProvisionedConcurrencyDisplay(f), "\t", timeout, "\t", formatDetachedDestinations(f), "\t", f.ID, "\t", "\n")
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func getProvisionedConcurrencyView(fn *models.Fn) *provisionedConcurrencyView {
	if fn == nil || fn.Annotations == nil {
		return nil
	}
	strategyRaw, ok := fn.Annotations[annotationProvisionedConcurrencyStrategy]
	if !ok {
		return nil
	}
	strategy, ok := strategyRaw.(string)
	if !ok || strings.TrimSpace(strategy) == "" {
		return nil
	}
	view := &provisionedConcurrencyView{Strategy: strings.ToUpper(strategy)}
	if countRaw, ok := fn.Annotations[annotationProvisionedConcurrencyCount]; ok {
		switch typed := countRaw.(type) {
		case int:
			count := typed
			view.Count = &count
		case int32:
			count := int(typed)
			view.Count = &count
		case int64:
			count := int(typed)
			view.Count = &count
		case float64:
			count := int(typed)
			view.Count = &count
		}
	}
	return view
}

func formatProvisionedConcurrencyDisplay(fn *models.Fn) string {
	view := getProvisionedConcurrencyView(fn)
	if view == nil {
		return ""
	}
	switch strings.ToUpper(view.Strategy) {
	case "NONE":
		return "none"
	case "CONSTANT":
		if view.Count == nil {
			return "constant"
		}
		return fmt.Sprintf("constant:%d", *view.Count)
	default:
		return strings.ToLower(view.Strategy)
	}
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

func warnUnsupportedProvisionedConcurrency() {
	fmt.Fprintln(os.Stderr, "Warning: --provisioned-concurrency is only supported with an oracle provider and will be ignored.")
}

func buildFunctionsManagementClient(oracleProvider *fnprovideroracle.OracleProvider) (*ocifunctions.FunctionsManagementClient, error) {
	client, err := ocifunctions.NewFunctionsManagementClientWithConfigurationProvider(oracleProvider.ConfigurationProvider)
	if err != nil {
		return nil, err
	}
	if oracleProvider.FnApiUrl != nil {
		client.Host = oracleProvider.FnApiUrl.String()
	} else {
		region, _ := oracleProvider.ConfigurationProvider.Region()
		if region != "" {
			client.SetRegion(region)
		}
	}
	return &client, nil
}

// ApplyProvisionedConcurrency applies OCI provisioned concurrency to a function when the active provider is Oracle.
func ApplyProvisionedConcurrency(p provider.Provider, fnID string, cfg *common.OCIProvisionedConcurrencyConfig) error {
	if p == nil || cfg == nil || fnID == "" {
		return nil
	}
	if err := common.ValidateProvisionedConcurrencyConfig(cfg); err != nil {
		return err
	}
	oracleProvider, ok := p.(*fnprovideroracle.OracleProvider)
	if !ok || oracleProvider == nil {
		return nil
	}
	mgmtClient, err := buildFunctionsManagementClient(oracleProvider)
	if err != nil {
		return err
	}

	var pcConfig ocifunctions.FunctionProvisionedConcurrencyConfig
	switch strings.ToUpper(cfg.Strategy) {
	case common.ProvisionedConcurrencyStrategyNone:
		pcConfig = ocifunctions.NoneProvisionedConcurrencyConfig{}
	case common.ProvisionedConcurrencyStrategyConstant:
		if cfg.Count == nil {
			return fmt.Errorf("provisioned concurrency count is required for CONSTANT strategy")
		}
		pcConfig = ocifunctions.ConstantProvisionedConcurrencyConfig{Count: cfg.Count}
	default:
		return fmt.Errorf("unsupported provisioned concurrency strategy %q", cfg.Strategy)
	}

	_, err = mgmtClient.UpdateFunction(context.Background(), ocifunctions.UpdateFunctionRequest{
		FunctionId: &fnID,
		UpdateFunctionDetails: ocifunctions.UpdateFunctionDetails{
			ProvisionedConcurrencyConfig: pcConfig,
		},
	})
	return err
}

func (f *fnsCmd) create(c *cli.Context) error {
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	pcConfig, err := common.ParseProvisionedConcurrencySpec(c.String("provisioned-concurrency"))
	if err != nil {
		return err
	}
	_, detachedSeconds, err := common.ParseDetachedTimeoutSpec(c.String("detached-timeout"))
	if err != nil {
		return err
	}
	onSuccess, err := common.ParseOCIDestinationSpec("--on-success", c.String("on-success"))
	if err != nil {
		return err
	}
	onFailure, err := common.ParseOCIDestinationSpec("--on-failure", c.String("on-failure"))
	if err != nil {
		return err
	}
	clearReq, err := validateDestinationFlagCombination(c)
	if err != nil {
		return err
	}

	fn := &models.Fn{}
	fn.Name = fnName
	fn.Image = c.Args().Get(2)

	WithFlags(c, fn)

	if fn.Name == "" {
		return errors.New("fnName path is missing")
	}
	if fn.Image == "" {
		return errors.New("no image specified")
	}
	if pcConfig != nil {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedProvisionedConcurrency()
		} else if err := SetProvisionedConcurrencyAnnotations(fn, pcConfig); err != nil {
			return err
		}
	}
	if detachedSeconds > 0 {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedDetachedTimeout()
		} else {
			SetDetachedTimeoutAnnotation(fn, detachedSeconds)
		}
	}
	if onSuccess != nil || onFailure != nil {
		if !common.IsOracleProvider(f.provider) {
			if onSuccess != nil {
				warnUnsupportedDestination("--on-success")
			}
			if onFailure != nil {
				warnUnsupportedDestination("--on-failure")
			}
		} else {
			SetDestinationAnnotations(fn, onSuccess, onFailure)
		}
	}
	if clearReq.Success || clearReq.Failure {
		if !common.IsOracleProvider(f.provider) {
			if clearReq.Success {
				warnUnsupportedDestination("--clear-on-success")
			}
			if clearReq.Failure {
				warnUnsupportedDestination("--clear-on-failure")
			}
		} else {
			SetClearDestinationAnnotations(fn, clearReq.Success, clearReq.Failure)
		}
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
	err := common.ValidateTagImageName(fn.Image)
	if err != nil {
		return nil, err
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

	fmt.Println("Successfully created function:", resp.Payload.Name, "with", resp.Payload.Image)
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
	pcConfig, err := common.ParseProvisionedConcurrencySpec(c.String("provisioned-concurrency"))
	if err != nil {
		return err
	}
	if err := common.ValidateProvisionedConcurrencyConfig(pcConfig); err != nil {
		return err
	}
	_, detachedSeconds, err := common.ParseDetachedTimeoutSpec(c.String("detached-timeout"))
	if err != nil {
		return err
	}
	onSuccess, err := common.ParseOCIDestinationSpec("--on-success", c.String("on-success"))
	if err != nil {
		return err
	}
	onFailure, err := common.ParseOCIDestinationSpec("--on-failure", c.String("on-failure"))
	if err != nil {
		return err
	}
	clearReq, err := validateDestinationFlagCombination(c)
	if err != nil {
		return err
	}

	app, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, app.ID, fnName)
	if err != nil {
		return err
	}

	WithFlags(c, fn)
	if detachedSeconds > 0 {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedDetachedTimeout()
		} else {
			SetDetachedTimeoutAnnotation(fn, detachedSeconds)
		}
	}
	if onSuccess != nil || onFailure != nil {
		if !common.IsOracleProvider(f.provider) {
			if onSuccess != nil {
				warnUnsupportedDestination("--on-success")
			}
			if onFailure != nil {
				warnUnsupportedDestination("--on-failure")
			}
		} else {
			SetDestinationAnnotations(fn, onSuccess, onFailure)
		}
	}
	if clearReq.Success || clearReq.Failure {
		if !common.IsOracleProvider(f.provider) {
			if clearReq.Success {
				warnUnsupportedDestination("--clear-on-success")
			}
			if clearReq.Failure {
				warnUnsupportedDestination("--clear-on-failure")
			}
		} else {
			SetClearDestinationAnnotations(fn, clearReq.Success, clearReq.Failure)
		}
	}

	err = PutFn(f.client, fn.ID, fn)
	if err != nil {
		return err
	}
	if pcConfig != nil {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedProvisionedConcurrency()
		} else if err := ApplyProvisionedConcurrency(f.provider, fn.ID, pcConfig); err != nil {
			return err
		}
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
	inspect, err := buildInspectFnMap(fn)
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %s", fnName, err)
	}

	if prop == "" {
		enc.Encode(inspect)
		return nil
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
