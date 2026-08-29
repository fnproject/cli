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
	annotationSourceType                     = "oracle.com/oci/sourceType"
	annotationPbfListingID                   = "oracle.com/oci/pbfListingId"
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

type sourceDetailsView struct {
	SourceType   string `json:"sourceType,omitempty"`
	PbfListingID string `json:"pbfListingId,omitempty"`
}

type fnFromJSON struct {
	DisplayName         string                 `json:"displayName"`
	Image               string                 `json:"image"`
	MemoryInMBs         uint64                 `json:"memoryInMBs"`
	Config              map[string]string      `json:"config"`
	TimeoutInSeconds    *int32                 `json:"timeoutInSeconds"`
	TraceConfig         map[string]interface{} `json:"traceConfig"`
	FreeformTags        map[string]string      `json:"freeformTags"`
	DefinedTags         common.OCIDefinedTags  `json:"definedTags"`
	IfMatch             string                 `json:"ifMatch"`
	WaitForState        string                 `json:"waitForState"`
	MaxWaitSeconds      int                    `json:"maxWaitSeconds"`
	WaitIntervalSeconds int                    `json:"waitIntervalSeconds"`
}

type fnDeleteFromJSON struct {
	IfMatch             string `json:"ifMatch"`
	WaitForState        string `json:"waitForState"`
	MaxWaitSeconds      int    `json:"maxWaitSeconds"`
	WaitIntervalSeconds int    `json:"waitIntervalSeconds"`
}

func applyFnFromJSON(fn *models.Fn, control *common.OCIRequestControl, input *fnFromJSON) {
	if fn == nil || input == nil {
		return
	}
	if strings.TrimSpace(fn.Name) == "" && strings.TrimSpace(input.DisplayName) != "" {
		fn.Name = strings.TrimSpace(input.DisplayName)
	}
	if strings.TrimSpace(input.Image) != "" {
		fn.Image = strings.TrimSpace(input.Image)
	}
	if input.MemoryInMBs > 0 {
		fn.Memory = input.MemoryInMBs
	}
	if len(input.Config) > 0 {
		fn.Config = input.Config
	}
	if input.TimeoutInSeconds != nil {
		fn.Timeout = input.TimeoutInSeconds
	}
	fn.Annotations = common.ApplyOCIResourceTagsToAnnotations(fn.Annotations, input.FreeformTags, input.DefinedTags)
	if fn.Annotations == nil {
		fn.Annotations = map[string]interface{}{}
	}
	if input.TraceConfig != nil {
		fn.Annotations[annotationOCIParityFnTraceConfig] = input.TraceConfig
	}
	if control != nil {
		if control.IfMatch == "" {
			control.IfMatch = strings.TrimSpace(input.IfMatch)
		}
		if control.WaitForState == "" {
			control.WaitForState = strings.ToUpper(strings.TrimSpace(input.WaitForState))
		}
		if control.MaxWaitSeconds == 0 {
			control.MaxWaitSeconds = input.MaxWaitSeconds
		}
		if control.WaitIntervalSeconds == 0 {
			control.WaitIntervalSeconds = input.WaitIntervalSeconds
		}
	}
}

func formatSourceDisplay(fn *models.Fn) string {
	view := getSourceDetailsView(fn)
	if view == nil {
		return ""
	}
	if view.PbfListingID != "" {
		return fmt.Sprintf("pbf:%s", view.PbfListingID)
	}
	return strings.ToLower(view.SourceType)
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
	cli.StringSliceFlag{
		Name:  "tag",
		Usage: "Freeform tag in key=value form (can be specified multiple times)",
	},
	cli.StringSliceFlag{
		Name:  "defined-tag",
		Usage: "Defined tag in namespace.key=value form (can be specified multiple times)",
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
	cli.StringFlag{
		Name:  "pbf",
		Usage: "Create the function from a Pre-Built Function listing OCID",
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
	cli.StringSliceFlag{
		Name:  "remove-tag",
		Usage: "Remove a freeform tag by key (can be specified multiple times)",
	},
	cli.StringSliceFlag{
		Name:  "remove-defined-tag",
		Usage: "Remove a defined tag by namespace.key (can be specified multiple times)",
	},
	cli.BoolFlag{
		Name:  "clear-tags",
		Usage: "Clear all freeform and defined tags",
	},
	cli.BoolFlag{
		Name:  "clear-freeform-tags",
		Usage: "Clear all freeform tags",
	},
	cli.BoolFlag{
		Name:  "clear-defined-tags",
		Usage: "Clear all defined tags",
	},
	cli.BoolFlag{
		Name:  "clear-on-success",
		Usage: "Clear OCI detached success destination",
	},
	cli.BoolFlag{
		Name:  "clear-on-failure",
		Usage: "Clear OCI detached failure destination",
	},
)

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
	if source := getSourceDetailsView(fn); source != nil {
		sourceData, err := json.Marshal(source)
		if err != nil {
			return nil, err
		}
		var sourceValue map[string]interface{}
		if err := json.Unmarshal(sourceData, &sourceValue); err != nil {
			return nil, err
		}
		inspect["sourceDetails"] = sourceValue
	}
	if fn != nil && fn.Annotations != nil {
		if trace, ok := fn.Annotations[annotationOCIParityFnTraceConfig]; ok {
			inspect["traceConfig"] = trace
		}
	}
	return inspect, nil
}

func getSourceDetailsView(fn *models.Fn) *sourceDetailsView {
	if fn == nil || fn.Annotations == nil {
		return nil
	}
	sourceTypeRaw, ok := fn.Annotations[annotationSourceType]
	if !ok {
		return nil
	}
	sourceType, ok := sourceTypeRaw.(string)
	if !ok || strings.TrimSpace(sourceType) == "" {
		return nil
	}
	view := &sourceDetailsView{SourceType: sourceType}
	if listingRaw, ok := fn.Annotations[annotationPbfListingID]; ok {
		if listingID, ok := listingRaw.(string); ok {
			view.PbfListingID = listingID
		}
	}
	return view
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
				SourceDetails          *sourceDetailsView          `json:"sourceDetails,omitempty"`
			}{
				Name:                   fn.Name,
				Image:                  fn.Image,
				ID:                     fn.ID,
				ProvisionedConcurrency: getProvisionedConcurrencyView(fn),
				DetachedMode:           getDetachedModeView(fn),
				SourceDetails:          getSourceDetailsView(fn),
			})
		}
		b, err := json.MarshalIndent(newFns, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "NAME", "\t", "IMAGE", "\t", "SOURCE", "\t", "PC", "\t", "DETACHED_TIMEOUT", "\t", "DESTINATIONS", "\t", "ID", "\n")

		for _, f := range fns {
			view := getDetachedModeView(f)
			timeout := ""
			if view != nil {
				timeout = view.Timeout
			}
			fmt.Fprint(w, f.Name, "\t", f.Image, "\t", formatSourceDisplay(f), "\t", formatProvisionedConcurrencyDisplay(f), "\t", timeout, "\t", formatDetachedDestinations(f), "\t", f.ID, "\t", "\n")
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

func buildCreateFnSuccessMessage(fn *models.Fn) string {
	if fn == nil {
		return "Successfully created function"
	}
	if source := getSourceDetailsView(fn); source != nil && strings.EqualFold(source.SourceType, "PRE_BUILT_FUNCTIONS") {
		if source.PbfListingID != "" {
			return fmt.Sprintf("Successfully created function: %s from PBF %s", fn.Name, source.PbfListingID)
		}
		return fmt.Sprintf("Successfully created function: %s from PBF", fn.Name)
	}
	if strings.TrimSpace(fn.Image) != "" {
		return fmt.Sprintf("Successfully created function: %s with %s", fn.Name, fn.Image)
	}
	return fmt.Sprintf("Successfully created function: %s", fn.Name)
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
	var fromJSON struct {
		DisplayName    string `json:"displayName"`
		ID             string `json:"id"`
		LifecycleState string `json:"lifecycleState"`
		SortBy         string `json:"sortBy"`
		SortOrder      string `json:"sortOrder"`
	}
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return nil, err
	}
	if strings.TrimSpace(fromJSON.DisplayName) != "" {
		params.DisplayName = &fromJSON.DisplayName
	}
	if strings.TrimSpace(fromJSON.ID) != "" {
		params.ID = &fromJSON.ID
	}
	if strings.TrimSpace(fromJSON.LifecycleState) != "" {
		lifecycle := strings.TrimSpace(fromJSON.LifecycleState)
		params.LifecycleState = &lifecycle
	}
	if strings.TrimSpace(fromJSON.SortBy) != "" {
		sortBy := strings.TrimSpace(fromJSON.SortBy)
		params.SortBy = &sortBy
	}
	if strings.TrimSpace(fromJSON.SortOrder) != "" {
		sortOrder := strings.TrimSpace(fromJSON.SortOrder)
		params.SortOrder = &sortOrder
	}
	ApplyGeneratedOCIParityFnListParams(c, params)

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
	isPBFSource := ff.Deploy != nil && ff.Deploy.OCI != nil && ff.Deploy.OCI.PBF != nil && strings.TrimSpace(ff.Deploy.OCI.PBF.ListingID) != ""
	if !isPBFSource && ff.ImageNameV20180708() != "" { // args take precedence
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
	if ff.Deploy != nil && ff.Deploy.OCI != nil {
		fn.Annotations = common.ApplyOCIResourceTagsToAnnotations(fn.Annotations, ff.Deploy.OCI.FreeformTags, ff.Deploy.OCI.DefinedTags)
		if ff.Deploy.OCI.PBF != nil {
			fn.Image = ""
			if err := setPBFSourceAnnotations(fn, ff.Deploy.OCI.PBF.ListingID); err != nil {
				return err
			}
		}
	}
	// do something with triggers here

	return nil
}

func warnUnsupportedProvisionedConcurrency() {
	fmt.Fprintln(os.Stderr, "Warning: --provisioned-concurrency is only supported with an oracle provider and will be ignored.")
}

func setPBFSourceAnnotations(fn *models.Fn, listingID string) error {
	listingID = strings.TrimSpace(listingID)
	if listingID == "" {
		return nil
	}
	if fn.Annotations == nil {
		fn.Annotations = make(map[string]interface{})
	}
	fn.Annotations[annotationSourceType] = "PRE_BUILT_FUNCTIONS"
	fn.Annotations[annotationPbfListingID] = listingID
	return nil
}

func fetchCurrentPBFMemoryRequirement(p provider.Provider, listingID string) (*int64, error) {
	oracleProvider, ok := p.(*fnprovideroracle.OracleProvider)
	if !ok || oracleProvider == nil {
		return nil, nil
	}
	mgmtClient, err := buildFunctionsManagementClient(oracleProvider)
	if err != nil {
		return nil, err
	}
	isCurrent := true
	limit := 1
	res, err := mgmtClient.ListPbfListingVersions(context.Background(), ocifunctions.ListPbfListingVersionsRequest{
		PbfListingId:     &listingID,
		IsCurrentVersion: &isCurrent,
		Limit:            &limit,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 || res.Items[0].Requirements == nil {
		return nil, nil
	}
	return res.Items[0].Requirements.MinMemoryRequiredInMBs, nil
}

func resolvePBFMemory(memory uint64, minRequired *int64) (uint64, error) {
	if minRequired == nil || *minRequired <= 0 {
		if memory > 0 {
			return memory, nil
		}
		return 0, fmt.Errorf("unable to determine the minimum memory required for this PBF; please supply --memory explicitly")
	}
	minimum := uint64(*minRequired)
	if memory == 0 {
		return minimum, nil
	}
	if memory < minimum {
		return 0, fmt.Errorf("--memory %d is below the minimum required for this PBF (%d MiB)", memory, minimum)
	}
	return memory, nil
}

// ResolvePBFMemoryForListing auto-selects or validates memory for a PBF-backed function.
func ResolvePBFMemoryForListing(p provider.Provider, fn *models.Fn, listingID string) error {
	if fn == nil || strings.TrimSpace(listingID) == "" {
		return nil
	}
	minMemory, err := fetchCurrentPBFMemoryRequirement(p, listingID)
	if err != nil {
		return fmt.Errorf("unable to determine PBF memory requirements: %w", err)
	}
	resolvedMemory, err := resolvePBFMemory(fn.Memory, minMemory)
	if err != nil {
		return err
	}
	fn.Memory = resolvedMemory
	return nil
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
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(f.provider, control)
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	codeOnly := readCodeOnlyCreateOptions(c)
	pbfListingID := strings.TrimSpace(c.String("pbf"))
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
	var fromJSON fnFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	applyFnFromJSON(fn, &control, &fromJSON)

	WithFlags(c, fn)
	if err := ApplyGeneratedOCIParityFnFlags(c, fn); err != nil {
		return err
	}
	annotations, err := common.ApplyOCIResourceTagFlagsToAnnotations(
		fn.Annotations,
		c.StringSlice("tag"),
		c.StringSlice("defined-tag"),
		nil,
		nil,
		false,
		false,
	)
	if err != nil {
		return err
	}
	fn.Annotations = annotations

	if fn.Name == "" {
		return errors.New("fnName path is missing")
	}
	if codeOnly.enabled() {
		if pbfListingID != "" || pcConfig != nil || detachedSeconds > 0 || onSuccess != nil || onFailure != nil || clearReq.Success || clearReq.Failure {
			return errors.New("code-only options cannot be combined with --pbf or OCI managed-function flags")
		}
		if err := applyCodeOnlyCreateOptions(f.provider, fn, codeOnly); err != nil {
			return err
		}
	} else if fn.Image != "" && pbfListingID != "" {
		return errors.New("--image and --pbf cannot be used together")
	} else if fn.Image == "" && pbfListingID == "" {
		return errors.New("no image specified")
	} else if pbfListingID != "" {
		if !common.IsOracleProvider(f.provider) {
			return errors.New("--pbf is only supported with an oracle provider")
		}
		if err := setPBFSourceAnnotations(fn, pbfListingID); err != nil {
			return err
		}
		if err := ResolvePBFMemoryForListing(f.provider, fn, pbfListingID); err != nil {
			return err
		}
	}
	if !codeOnly.enabled() && pcConfig != nil {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedProvisionedConcurrency()
		} else if err := SetProvisionedConcurrencyAnnotations(fn, pcConfig); err != nil {
			return err
		}
	}
	if !codeOnly.enabled() && detachedSeconds > 0 {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedDetachedTimeout()
		} else {
			SetDetachedTimeoutAnnotation(fn, detachedSeconds)
		}
	}
	if !codeOnly.enabled() && (onSuccess != nil || onFailure != nil) {
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
	if !codeOnly.enabled() && (clearReq.Success || clearReq.Failure) {
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

	createdFn, err := CreateFnWithControl(f.client, a.ID, fn, control)
	if err != nil {
		return err
	}
	if err := common.WaitForFunctionState(f.provider, createdFn.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}
	return err
}

// CreateFn request
func CreateFn(r *fnclient.Fn, appID string, fn *models.Fn) (*models.Fn, error) {
	return CreateFnWithControl(r, appID, fn, common.OCIRequestControl{})
}

func CreateFnWithControl(r *fnclient.Fn, appID string, fn *models.Fn, control common.OCIRequestControl) (*models.Fn, error) {
	fn.AppID = appID
	if fn.Image != "" {
		err := common.ValidateTagImageName(fn.Image)
		if err != nil {
			return nil, err
		}
	}

	resp, err := r.Fns.CreateFn(&apifns.CreateFnParams{
		Context:             context.Background(),
		Body:                fn,
		WaitForState:        control.WaitForState,
		MaxWaitSeconds:      int64(control.MaxWaitSeconds),
		WaitIntervalSeconds: int64(control.WaitIntervalSeconds),
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

	fmt.Println(buildCreateFnSuccessMessage(resp.Payload))
	return resp.Payload, nil
}

// PutFn updates the fn with the given ID using the content of the provided fn
func PutFn(f *fnclient.Fn, fnID string, fn *models.Fn) error {
	return PutFnWithControl(f, fnID, fn, common.OCIRequestControl{})
}

func PutFnWithControl(f *fnclient.Fn, fnID string, fn *models.Fn, control common.OCIRequestControl) error {
	if fn.Image != "" {
		err := common.ValidateTagImageName(fn.Image)
		if err != nil {
			return err
		}
	}

	_, err := f.Fns.UpdateFn(&apifns.UpdateFnParams{
		Context:             context.Background(),
		FnID:                fnID,
		Body:                fn,
		IfMatch:             control.IfMatch,
		WaitForState:        control.WaitForState,
		MaxWaitSeconds:      int64(control.MaxWaitSeconds),
		WaitIntervalSeconds: int64(control.WaitIntervalSeconds),
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
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(f.provider, control)
	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)
	codeOnly := readCodeOnlyUpdateOptions(c)
	if strings.TrimSpace(c.String("pbf")) != "" {
		return errors.New("--pbf is only supported when creating a function")
	}

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

	appObj, err := app.GetAppByName(f.client, appName)
	if err != nil {
		return err
	}
	fn, err := GetFnByName(f.client, appObj.ID, fnName)
	if err != nil {
		return err
	}
	var fromJSON fnFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	applyFnFromJSON(fn, &control, &fromJSON)

	WithFlags(c, fn)
	if err := ApplyGeneratedOCIParityFnFlags(c, fn); err != nil {
		return err
	}
	if codeOnly.enabled() {
		if pcConfig != nil || detachedSeconds > 0 || onSuccess != nil || onFailure != nil || clearReq.Success || clearReq.Failure {
			return errors.New("code-only update options cannot be combined with OCI managed-function flags")
		}
		if err := applyCodeOnlyUpdateOptions(f.provider, fn, codeOnly); err != nil {
			return err
		}
	}
	annotations, err := common.ApplyOCIResourceTagFlagsToAnnotations(
		fn.Annotations,
		c.StringSlice("tag"),
		c.StringSlice("defined-tag"),
		c.StringSlice("remove-tag"),
		c.StringSlice("remove-defined-tag"),
		c.Bool("clear-freeform-tags") || c.Bool("clear-tags"),
		c.Bool("clear-defined-tags") || c.Bool("clear-tags"),
	)
	if err != nil {
		return err
	}
	fn.Annotations = annotations

	if !codeOnly.enabled() && detachedSeconds > 0 {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedDetachedTimeout()
		} else {
			SetDetachedTimeoutAnnotation(fn, detachedSeconds)
		}
	}
	if !codeOnly.enabled() && (onSuccess != nil || onFailure != nil) {
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
	if !codeOnly.enabled() && (clearReq.Success || clearReq.Failure) {
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

	err = PutFnWithControl(f.client, fn.ID, fn, control)
	if err != nil {
		return err
	}
	if !codeOnly.enabled() && pcConfig != nil {
		if !common.IsOracleProvider(f.provider) {
			warnUnsupportedProvisionedConcurrency()
		} else if err := ApplyProvisionedConcurrency(f.provider, fn.ID, pcConfig); err != nil {
			return err
		}
	}
	if err := common.WaitForFunctionState(f.provider, fn.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}
	if err := common.InvalidateInvokeEndpointCacheForFunction(f.provider, appName, fnName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: unable to invalidate invoke endpoint cache: %v\n", err)
	}

	fmt.Println(appName, fnName, "updated")
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

func validateRuntimeVersionMatchesRuntime(p provider.Provider, runtimeName, runtimeVersionID string) error {
	ociProvider, ok := p.(*fnprovideroracle.OracleProvider)
	if !ok || ociProvider == nil {
		return fmt.Errorf("runtime version validation requires an oracle provider")
	}
	client, err := ocifunctions.NewFunctionsManagementClientWithConfigurationProvider(ociProvider.ConfigurationProvider)
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
	limit := 1
	request := ocifunctions.ListFunctionsRuntimeVersionsRequest{
		FunctionsRuntimeName:      &runtimeName,
		FunctionsRuntimeVersionId: &runtimeVersionID,
		Limit:                     &limit,
	}
	response, err := client.ListFunctionsRuntimeVersions(context.Background(), request)
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		return fmt.Errorf("runtime version %s does not belong to runtime %s", runtimeVersionID, runtimeName)
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
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(f.provider, control)
	var fromJSON fnDeleteFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	if control.IfMatch == "" {
		control.IfMatch = strings.TrimSpace(fromJSON.IfMatch)
	}
	if control.WaitForState == "" {
		control.WaitForState = strings.ToUpper(strings.TrimSpace(fromJSON.WaitForState))
	}
	if control.MaxWaitSeconds == 0 {
		control.MaxWaitSeconds = fromJSON.MaxWaitSeconds
	}
	if control.WaitIntervalSeconds == 0 {
		control.WaitIntervalSeconds = fromJSON.WaitIntervalSeconds
	}
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
	params.IfMatch = control.IfMatch
	params.WaitForState = control.WaitForState
	params.MaxWaitSeconds = int64(control.MaxWaitSeconds)
	params.WaitIntervalSeconds = int64(control.WaitIntervalSeconds)
	_, err = f.client.Fns.DeleteFn(params)

	if err != nil {
		return err
	}
	if err := common.WaitForFunctionState(f.provider, fn.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}
	if err := common.InvalidateInvokeEndpointCacheForFunction(f.provider, appName, fnName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: unable to invalidate invoke endpoint cache: %v\n", err)
	}

	fmt.Println("Function", fnName, "deleted")
	return nil
}
