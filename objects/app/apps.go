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

package app

import (
	ocifunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"context"
	"strings"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"

	fnclient "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

const (
	SHAPE_PARAMETER  = "shape"
	annotationSubnet = "oracle.com/oci/subnetIds"
)

type appsCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
}

type appFromJSON struct {
	DisplayName             string                            `json:"displayName"`
	Config                  map[string]string                 `json:"config"`
	SyslogURL               *string                           `json:"syslogUrl"`
	Shape                   string                            `json:"shape"`
	SubnetIds               []string                          `json:"subnetIds"`
	FreeformTags            map[string]string                 `json:"freeformTags"`
	DefinedTags             common.OCIDefinedTags             `json:"definedTags"`
	TraceConfig             map[string]interface{}            `json:"traceConfig"`
	NetworkSecurityGroupIds []string                          `json:"networkSecurityGroupIds"`
	ImagePolicyConfig       map[string]interface{}            `json:"imagePolicyConfig"`
	SecurityAttributes      map[string]map[string]interface{} `json:"securityAttributes"`
	IfMatch                 string                            `json:"ifMatch"`
	WaitForState            string                            `json:"waitForState"`
	MaxWaitSeconds          int                               `json:"maxWaitSeconds"`
	WaitIntervalSeconds     int                               `json:"waitIntervalSeconds"`
}

type appDeleteFromJSON struct {
	IfMatch             string `json:"ifMatch"`
	WaitForState        string `json:"waitForState"`
	MaxWaitSeconds      int    `json:"maxWaitSeconds"`
	WaitIntervalSeconds int    `json:"waitIntervalSeconds"`
}

type appChangeCompartmentFromJSON struct {
	CompartmentID       string `json:"compartmentId"`
	IfMatch             string `json:"ifMatch"`
	WaitForState        string `json:"waitForState"`
	MaxWaitSeconds      int    `json:"maxWaitSeconds"`
	WaitIntervalSeconds int    `json:"waitIntervalSeconds"`
}

func applyAppFromJSON(app *modelsv2.App, control *common.OCIRequestControl, input *appFromJSON) {
	if app == nil || input == nil {
		return
	}
	if strings.TrimSpace(app.Name) == "" && strings.TrimSpace(input.DisplayName) != "" {
		app.Name = strings.TrimSpace(input.DisplayName)
	}
	if len(input.Config) > 0 {
		app.Config = input.Config
	}
	if input.SyslogURL != nil {
		app.SyslogURL = input.SyslogURL
	}
	if input.Shape != "" {
		app.Shape = input.Shape
	}
	app.Annotations = common.ApplyOCIResourceTagsToAnnotations(app.Annotations, input.FreeformTags, input.DefinedTags)
	if app.Annotations == nil {
		app.Annotations = map[string]interface{}{}
	}
	if len(input.SubnetIds) > 0 {
		values := make([]interface{}, 0, len(input.SubnetIds))
		for _, id := range input.SubnetIds {
			values = append(values, strings.TrimSpace(id))
		}
		app.Annotations[annotationSubnet] = values
	}
	if input.TraceConfig != nil {
		app.Annotations[annotationOCIParityTraceConfig] = input.TraceConfig
	}
	if len(input.NetworkSecurityGroupIds) > 0 {
		values := make([]interface{}, 0, len(input.NetworkSecurityGroupIds))
		for _, id := range input.NetworkSecurityGroupIds {
			values = append(values, strings.TrimSpace(id))
		}
		app.Annotations[annotationOCIParityNetworkSecurityGroupIds] = values
	}
	if input.ImagePolicyConfig != nil {
		app.Annotations[annotationOCIParityImagePolicyConfig] = input.ImagePolicyConfig
	}
	if input.SecurityAttributes != nil {
		app.Annotations[annotationOCIParitySecurityAttributes] = input.SecurityAttributes
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

func printApps(c *cli.Context, apps []*modelsv2.App) error {
	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		var allApps []interface{}
		for _, app := range apps {
			a := struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}{app.Name,
				app.ID,
			}
			allApps = append(allApps, a)
		}
		b, err := json.MarshalIndent(allApps, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "NAME", "\t", "ID", "\t", "\n")
		for _, app := range apps {
			fmt.Fprint(w, app.Name, "\t", app.ID, "\t", "\n")

		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (a *appsCmd) list(c *cli.Context) error {
	resApps, err := getApps(c, a.client)
	if err != nil {
		return err
	}
	return printApps(c, resApps)
}

// getApps returns an array of all apps in the given context and client
func getApps(c *cli.Context, client *fnclient.Fn) ([]*modelsv2.App, error) {
	params := &apiapps.ListAppsParams{Context: context.Background()}
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
	ApplyGeneratedOCIParityAppListParams(c, params)
	var resApps []*modelsv2.App
	for {
		resp, err := client.Apps.ListApps(params)
		if err != nil {
			return nil, err
		}

		resApps = append(resApps, resp.Payload.Items...)

		n := c.Int64("n")

		howManyMore := n - int64(len(resApps)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	if len(resApps) == 0 {
		fmt.Fprint(os.Stderr, "No apps found\n")
		return nil, nil
	}
	return resApps, nil
}

// BashCompleteApps can be called from a BashComplete function
// to provide app completion suggestions (Does not check if the
// current context already contains an app name as an argument.
// This should be checked before calling this)
func BashCompleteApps(c *cli.Context) {
	provider, err := client.CurrentProvider()
	if err != nil {
		return
	}
	resp, err := getApps(c, provider.APIClientv2())
	if err != nil {
		return
	}
	for _, r := range resp {
		fmt.Println(r.Name)
	}
}

func appWithFlags(c *cli.Context, app *modelsv2.App) error {
	if c.IsSet("syslog-url") {
		str := c.String("syslog-url")
		app.SyslogURL = &str
	}
	if len(c.StringSlice("config")) > 0 {
		app.Config = common.ExtractConfig(c.StringSlice("config"))
	}
	if len(c.StringSlice("annotation")) > 0 {
		app.Annotations = common.ExtractAnnotations(c)
	}
	annotations, err := common.ApplyOCIResourceTagFlagsToAnnotations(
		app.Annotations,
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
	app.Annotations = annotations
	if err := ApplyGeneratedOCIParityAppFlags(c, app); err != nil {
		return err
	}
	if err := setSubnetIDAnnotations(app, c.StringSlice("subnet-id")); err != nil {
		return err
	}
	return nil
}

func normalizeSubnetIDs(subnetIDs []string) ([]string, error) {
	if len(subnetIDs) == 0 {
		return nil, nil
	}
	normalized := make([]string, 0, len(subnetIDs))
	for _, subnetID := range subnetIDs {
		trimmed := strings.TrimSpace(subnetID)
		if trimmed == "" {
			return nil, fmt.Errorf("subnet IDs must not be empty")
		}
		normalized = append(normalized, trimmed)
	}
	return normalized, nil
}

func subnetIDsToAnnotationValue(subnetIDs []string) []interface{} {
	values := make([]interface{}, len(subnetIDs))
	for i, subnetID := range subnetIDs {
		values[i] = subnetID
	}
	return values
}

func setSubnetIDAnnotations(app *modelsv2.App, subnetIDs []string) error {
	normalized, err := normalizeSubnetIDs(subnetIDs)
	if err != nil {
		return err
	}
	if len(normalized) == 0 {
		return nil
	}
	if app.Annotations == nil {
		app.Annotations = make(map[string]interface{})
	}
	if _, exists := app.Annotations[annotationSubnet]; exists {
		return fmt.Errorf("--subnet-id cannot be used together with --annotation %s", annotationSubnet)
	}
	app.Annotations[annotationSubnet] = subnetIDsToAnnotationValue(normalized)
	return nil
}

func validateSubnetIDUpdateSupported(p provider.Provider, subnetIDs []string) error {
	if len(subnetIDs) == 0 {
		return nil
	}
	if common.IsOracleProvider(p) {
		return fmt.Errorf("--subnet-id is not currently supported with `fn update app` for Oracle-backed apps; use the OCI CLI or recreate the app with the desired subnet IDs")
	}
	return nil
}

func validateSubnetIDCreateRequired(p provider.Provider, app *modelsv2.App) error {
	if !common.IsOracleProvider(p) {
		return nil
	}
	if app != nil && app.Annotations != nil {
		if _, ok := app.Annotations[annotationSubnet]; ok {
			return nil
		}
	}
	return fmt.Errorf("Oracle-backed app creation requires at least one subnet; use --subnet-id <ocid> (or --annotation %s='[\"<subnet-ocid>\"]')", annotationSubnet)
}

func (a *appsCmd) create(c *cli.Context) error {
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(a.provider, control)
	app := &modelsv2.App{
		Name: c.Args().Get(0),
	}
	var fromJSON appFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	applyAppFromJSON(app, &control, &fromJSON)

	if err := appWithFlags(c, app); err != nil {
		return err
	}
	// If architectures flag is not set then default it to nil
	if c.IsSet(SHAPE_PARAMETER) {
		shapeParam := c.String(SHAPE_PARAMETER)

		// Check for architectures parameter passed or set to default
		if len(shapeParam) == 0 {
			return errors.New("no shape specified for the application")
		}

		if _, ok := common.ShapeMap[shapeParam]; !ok {
			return errors.New("invalid shape specified for the application")
		}
		app.Shape = shapeParam
	}
	if err := validateSubnetIDCreateRequired(a.provider, app); err != nil {
		return err
	}
	createdApp, err := CreateAppWithControl(a.client, app, control)
	if err != nil {
		return err
	}
	if err := common.WaitForAppState(a.provider, createdApp.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}
	return err
}

// CreateApp creates a new app using the given client
func CreateApp(a *fnclient.Fn, app *modelsv2.App) (*modelsv2.App, error) {
	return CreateAppWithControl(a, app, common.OCIRequestControl{})
}

func CreateAppWithControl(a *fnclient.Fn, app *modelsv2.App, control common.OCIRequestControl) (*modelsv2.App, error) {
	resp, err := a.Apps.CreateApp(&apiapps.CreateAppParams{
		Context:             context.Background(),
		Body:                app,
		WaitForState:        control.WaitForState,
		MaxWaitSeconds:      int64(control.MaxWaitSeconds),
		WaitIntervalSeconds: int64(control.WaitIntervalSeconds),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.CreateAppBadRequest:
			err = fmt.Errorf("%v", e.Payload.Message)
		case *apiapps.CreateAppConflict:
			err = fmt.Errorf("%v", e.Payload.Message)
		}
		return nil, err
	}
	fmt.Println("Successfully created app: ", resp.Payload.Name)
	return resp.Payload, nil
}

func (a *appsCmd) update(c *cli.Context) error {
	appName := c.Args().First()
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(a.provider, control)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}
	var fromJSON appFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	applyAppFromJSON(app, &control, &fromJSON)

	if err := validateSubnetIDUpdateSupported(a.provider, c.StringSlice("subnet-id")); err != nil {
		return err
	}

	if err := appWithFlags(c, app); err != nil {
		return err
	}

	updatedApp, err := PutAppWithControl(a.client, app.ID, app, control)
	if err != nil {
		return err
	}
	if err := common.WaitForAppState(a.provider, updatedApp.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func (a *appsCmd) setConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)
	value := c.Args().Get(2)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	app.Config = make(map[string]string)
	app.Config[key] = value

	if _, err = PutApp(a.client, app.ID, app); err != nil {
		return fmt.Errorf("Error updating app configuration: %v", err)
	}

	fmt.Println(appName, "updated", key, "with", value)
	return nil
}

func (a *appsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	val, ok := app.Config[key]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (a *appsCmd) listConfig(c *cli.Context) error {
	appName := c.Args().Get(0)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	if len(app.Config) == 0 {
		fmt.Fprintf(os.Stderr, "No config found for app: %s\n", appName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "KEY", "\t", "VALUE", "\n")
	for key, val := range app.Config {
		fmt.Fprint(w, key, "\t", val, "\n")
	}
	w.Flush()

	return nil
}

func (a *appsCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}
	_, ok := app.Config[key]
	if !ok {
		fmt.Printf("Config key '%s' does not exist. Nothing to do.\n", key)
		return nil
	}
	app.Config[key] = ""

	_, err = PutApp(a.client, app.ID, app)
	if err != nil {
		return err
	}

	fmt.Printf("Removed key '%s' from app '%s' \n", key, appName)
	return nil
}

func (a *appsCmd) inspect(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("Missing app name after the inspect command")
	}

	appName := c.Args().First()
	prop := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	inspectData, err := buildAppInspectData(app)
	if err != nil {
		return fmt.Errorf("Could not build app inspect data: %v", err)
	}

	if prop == "" {
		enc.Encode(inspectData)
		return nil
	}
	jq := jsonq.NewQuery(inspectData)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return fmt.Errorf("Failed to inspect field %v", prop)
	}
	enc.Encode(field)

	return nil
}

func buildAppInspectData(app *modelsv2.App) (map[string]interface{}, error) {
	data, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	inspect := map[string]interface{}{}
	if err := json.Unmarshal(data, &inspect); err != nil {
		return nil, err
	}

	if app == nil || app.Annotations == nil {
		return inspect, nil
	}

	if v, ok := app.Annotations[annotationOCIParityTraceConfig]; ok {
		inspect["traceConfig"] = v
	}
	if v, ok := app.Annotations[annotationOCIParityNetworkSecurityGroupIds]; ok {
		inspect["networkSecurityGroupIds"] = v
	}
	if v, ok := app.Annotations[annotationOCIParityImagePolicyConfig]; ok {
		inspect["imagePolicyConfig"] = v
	}
	if v, ok := app.Annotations[annotationOCIParitySecurityAttributes]; ok {
		inspect["securityAttributes"] = v
	}

	return inspect, nil
}

func (a *appsCmd) delete(c *cli.Context) error {
	appName := c.Args().First()
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(a.provider, control)
	var fromJSON appDeleteFromJSON
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
	if appName == "" {
		//return errors.New("App name required to delete")
	}

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	//recursive delete of sub-objects
	if c.Bool("recursive") {
		fns, triggers, err := common.ListFnsAndTriggersInApp(c, a.client, app)
		if err != nil {
			return fmt.Errorf("Failed to get associated objects: %s", err)
		}

		//Forced deletion
		var shouldContinue bool
		if c.Bool("force") {
			shouldContinue = true
		} else {
			shouldContinue = common.UserConfirmedMultiResourceDeletion([]*modelsv2.App{app}, fns, triggers)
		}

		if shouldContinue {
			err := common.DeleteTriggers(c, a.client, triggers)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
			err = common.DeleteFunctions(c, a.client, fns)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
		} else {
			return nil
		}
	}

	_, err = a.client.Apps.DeleteApp(&apiapps.DeleteAppParams{
		Context:             context.Background(),
		AppID:               app.ID,
		IfMatch:             control.IfMatch,
		WaitForState:        control.WaitForState,
		MaxWaitSeconds:      int64(control.MaxWaitSeconds),
		WaitIntervalSeconds: int64(control.WaitIntervalSeconds),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.DeleteAppNotFound:
			return errors.New(e.Payload.Message)
		}
		return err
	}
	if err := common.WaitForAppState(a.provider, app.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}

	fmt.Println("App", appName, "deleted")
	return nil
}

func (a *appsCmd) changeCompartment(c *cli.Context) error {
	control := common.ExtractOCIRequestControl(c)
	common.WarnUnsupportedOCIRequestControl(a.provider, control)
	var fromJSON appChangeCompartmentFromJSON
	if err := common.LoadCLIJSONInput(c.String("from-json"), &fromJSON); err != nil {
		return err
	}
	compartmentID := strings.TrimSpace(c.String("compartment-id"))
	if compartmentID == "" {
		compartmentID = strings.TrimSpace(fromJSON.CompartmentID)
	}
	if compartmentID == "" {
		return fmt.Errorf("compartment id is required")
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
	if !common.IsOracleProvider(a.provider) {
		return fmt.Errorf("change-compartment is only supported with an oracle provider")
	}
	appName := c.Args().First()
	appObj, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}
	mgmtClient, err := common.BuildOCIManagementClient(a.provider)
	if err != nil {
		return err
	}
	if mgmtClient == nil {
		return fmt.Errorf("unable to build OCI Functions management client")
	}
	req := ocifunctions.ChangeApplicationCompartmentRequest{
		ApplicationId: &appObj.ID,
		ChangeApplicationCompartmentDetails: ocifunctions.ChangeApplicationCompartmentDetails{
			CompartmentId: &compartmentID,
		},
		IfMatch: stringPtr(control.IfMatch),
	}
	_, err = mgmtClient.ChangeApplicationCompartment(context.Background(), req)
	if err != nil {
		return err
	}
	if err := common.WaitForAppState(a.provider, appObj.ID, control.WaitForState, control.MaxWaitSeconds, control.WaitIntervalSeconds); err != nil {
		return err
	}
	fmt.Printf("App %s moved to compartment %s\n", appName, compartmentID)
	return nil
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

// PutApp updates the app with the given ID using the content of the provided app
func PutApp(a *fnclient.Fn, appID string, app *modelsv2.App) (*modelsv2.App, error) {
	return PutAppWithControl(a, appID, app, common.OCIRequestControl{})
}

func PutAppWithControl(a *fnclient.Fn, appID string, app *modelsv2.App, control common.OCIRequestControl) (*modelsv2.App, error) {
	resp, err := a.Apps.UpdateApp(&apiapps.UpdateAppParams{
		Context:             context.Background(),
		AppID:               appID,
		Body:                app,
		IfMatch:             control.IfMatch,
		WaitForState:        control.WaitForState,
		MaxWaitSeconds:      int64(control.MaxWaitSeconds),
		WaitIntervalSeconds: int64(control.WaitIntervalSeconds),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.UpdateAppBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}

	return resp.Payload, nil
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}

// GetAppByName looks up an app by name using the given client
func GetAppByName(client *fnclient.Fn, appName string) (*modelsv2.App, error) {
	appsResp, err := client.Apps.ListApps(&apiapps.ListAppsParams{
		Context: context.Background(),
		Name:    &appName,
	})
	if err != nil {
		return nil, err
	}

	var app *modelsv2.App
	if len(appsResp.Payload.Items) > 0 {
		app = appsResp.Payload.Items[0]
	} else {
		return nil, NameNotFoundError{appName}
	}

	return app, nil
}
