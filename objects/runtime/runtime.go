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

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/fn_go/provider/oracle"
	ociCommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

func (c *runtimeCmd) ensureOracleProvider() (*oracle.OracleProvider, error) {
	ociProvider, ok := c.provider.(*oracle.OracleProvider)
	if !ok || ociProvider == nil {
		return nil, fmt.Errorf("runtime discovery requires an oracle provider")
	}
	return ociProvider, nil
}

func (c *runtimeCmd) listRuntimes(cliCtx *cli.Context) error {
	ociProvider, err := c.ensureOracleProvider()
	if err != nil {
		return err
	}

	items, err := listFunctionsRuntimes(ociProvider)
	if err != nil {
		return err
	}

	return printRuntimes(cliCtx, items)
}

// ResolveCodeOnlyRuntimeName resolves a code-only language alias to an active
// managed runtime name available from the configured Functions service. Explicit
// managed runtime names are also validated before they are written to func.yaml.
func ResolveCodeOnlyRuntimeName(name string) (string, error) {
	requestedName := strings.TrimSpace(name)
	alias := codeOnlyRuntimeAlias(requestedName)

	provider, err := client.CurrentProvider()
	if err != nil {
		return "", err
	}
	ociProvider, ok := provider.(*oracle.OracleProvider)
	if !ok || ociProvider == nil {
		// `fn init` is intentionally usable before a context has been configured.
		// A fully-qualified managed runtime can be written to func.yaml without
		// consulting the service; it will be validated when the function is
		// created or deployed. Language aliases, on the other hand, need runtime
		// discovery to select the currently supported runtime.
		if isExplicitCodeOnlyRuntimeName(requestedName) {
			return requestedName, nil
		}
		if isCodeOnlyRuntimeAlias(alias) {
			return "", fmt.Errorf("runtime alias %q requires an oracle provider so Fn CLI can select an active managed runtime", requestedName)
		}
		return "", fmt.Errorf("unsupported code-only runtime name %q; run `fn list runtimes` to view supported runtimes", requestedName)
	}

	items, err := listFunctionsRuntimes(ociProvider)
	if err != nil {
		return "", err
	}
	if isCodeOnlyRuntimeAlias(alias) {
		return SelectCodeOnlyRuntimeName(alias, items)
	}
	return ValidateCodeOnlyRuntimeName(requestedName, items)
}

// ValidateCodeOnlyRuntimeName ensures an explicitly supplied managed runtime
// name is active in the configured Functions service.
func ValidateCodeOnlyRuntimeName(name string, items []functions.FunctionsRuntimeSummary) (string, error) {
	requestedName := strings.TrimSpace(name)
	for _, item := range items {
		if item.LifecycleState != functions.FunctionsRuntimeLifecycleStateActive ||
			!strings.EqualFold(strings.TrimSpace(stringValue(item.Name)), requestedName) {
			continue
		}
		return stringValue(item.Name), nil
	}

	return "", fmt.Errorf("no active managed runtime named %q; run `fn list runtimes` to view supported runtimes", requestedName)
}

// SelectCodeOnlyRuntimeName selects the active managed runtime with the latest
// deprecation date for a code-only language alias.
func SelectCodeOnlyRuntimeName(alias string, items []functions.FunctionsRuntimeSummary) (string, error) {
	alias = codeOnlyRuntimeAlias(alias)
	if !isCodeOnlyRuntimeAlias(alias) {
		return "", fmt.Errorf("unsupported code-only runtime alias %q", alias)
	}

	var selected *functions.FunctionsRuntimeSummary
	for i := range items {
		item := &items[i]
		if item.LifecycleState != functions.FunctionsRuntimeLifecycleStateActive || !codeOnlyRuntimeMatchesAlias(alias, *item) {
			continue
		}
		if selected == nil || codeOnlyRuntimePreferred(*item, *selected) {
			selected = item
		}
	}
	if selected == nil || selected.Name == nil || strings.TrimSpace(*selected.Name) == "" {
		return "", fmt.Errorf("no active managed runtime found for language %q; run `fn list runtimes` to view supported runtimes", alias)
	}
	return *selected.Name, nil
}

func isCodeOnlyRuntimeAlias(name string) bool {
	for _, language := range []string{"java", "go", "node", "python"} {
		if name == language {
			return true
		}
		if strings.HasPrefix(name, language) && allDigits(strings.TrimPrefix(name, language)) {
			return true
		}
	}
	return false
}

// isExplicitCodeOnlyRuntimeName reports whether name has the managed-runtime
// form used by the code-only service. These names can be used by `fn init`
// without an OCI context because no selection or service lookup is needed.
func isExplicitCodeOnlyRuntimeName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "ol8" || name == "ol9" {
		return true
	}

	for _, language := range []string{"java", "node", "python"} {
		if !strings.HasPrefix(name, language) {
			continue
		}
		versionAndOS := strings.TrimPrefix(name, language)
		version, osName, hasOS := strings.Cut(versionAndOS, ".")
		if !allDigits(version) {
			return false
		}
		return !hasOS || osName == "ol8" || osName == "ol9"
	}
	return false
}

func codeOnlyRuntimeAlias(name string) string {
	var alias strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			alias.WriteRune(r)
		}
	}
	return alias.String()
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func codeOnlyRuntimeMatchesAlias(alias string, item functions.FunctionsRuntimeSummary) bool {
	alias = codeOnlyRuntimeAlias(alias)
	name := strings.ToLower(strings.TrimSpace(stringValue(item.Name)))
	if strings.HasPrefix(alias, "go") {
		return name == "ol9"
	}
	language := codeOnlyRuntimeAlias(stringValue(item.Language))
	if alias == "java" || alias == "node" || alias == "python" {
		return strings.HasPrefix(language, alias)
	}
	return language == alias
}

func codeOnlyRuntimePreferred(candidate, current functions.FunctionsRuntimeSummary) bool {
	if candidate.TimeDeprecated != nil && current.TimeDeprecated != nil {
		if !candidate.TimeDeprecated.Equal(current.TimeDeprecated.Time) {
			return candidate.TimeDeprecated.After(current.TimeDeprecated.Time)
		}
	} else if candidate.TimeDeprecated != nil {
		return true
	} else if current.TimeDeprecated != nil {
		return false
	}
	return stringValue(candidate.Name) > stringValue(current.Name)
}

func listFunctionsRuntimes(ociProvider *oracle.OracleProvider) ([]functions.FunctionsRuntimeSummary, error) {
	client, err := newFunctionsClient(ociProvider)
	if err != nil {
		return nil, err
	}

	request := functions.ListFunctionsRuntimesRequest{}
	var items []functions.FunctionsRuntimeSummary
	for {
		response, err := client.ListFunctionsRuntimes(context.Background(), request)
		if err != nil {
			return nil, err
		}
		items = append(items, response.Items...)
		if response.OpcNextPage == nil {
			break
		}
		request.Page = response.OpcNextPage
	}
	return items, nil
}

func (c *runtimeCmd) listRuntimeVersions(cliCtx *cli.Context) error {
	runtimeName := strings.TrimSpace(cliCtx.String("runtime-name"))
	if runtimeName == "" {
		return fmt.Errorf("--runtime-name is required")
	}

	ociProvider, err := c.ensureOracleProvider()
	if err != nil {
		return err
	}

	client, err := newFunctionsClient(ociProvider)
	if err != nil {
		return err
	}

	request := functions.ListFunctionsRuntimeVersionsRequest{
		FunctionsRuntimeName: &runtimeName,
	}
	var items []functions.FunctionsRuntimeVersionSummary
	for {
		response, err := client.ListFunctionsRuntimeVersions(context.Background(), request)
		if err != nil {
			return err
		}
		items = append(items, response.Items...)
		if response.OpcNextPage == nil {
			break
		}
		request.Page = response.OpcNextPage
	}

	return printRuntimeVersions(cliCtx, items)
}

func (c *runtimeCmd) getLatestRuntimeVersion(cliCtx *cli.Context) error {
	runtimeName := strings.TrimSpace(cliCtx.String("runtime-name"))
	if runtimeName == "" {
		return fmt.Errorf("--runtime-name is required")
	}

	ociProvider, err := c.ensureOracleProvider()
	if err != nil {
		return err
	}

	client, err := newFunctionsClient(ociProvider)
	if err != nil {
		return err
	}

	request := functions.ListFunctionsRuntimeVersionsRequest{
		FunctionsRuntimeName: &runtimeName,
		IsCurrentVersion:     ociCommon.Bool(true),
		Limit:                ociCommon.Int(1),
	}
	response, err := client.ListFunctionsRuntimeVersions(context.Background(), request)
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		return fmt.Errorf("no runtime versions found for runtime %s", runtimeName)
	}
	return printLatestRuntimeVersion(cliCtx, response.Items[0])
}

func printRuntimes(cliCtx *cli.Context, items []functions.FunctionsRuntimeSummary) error {
	outputFormat := strings.ToLower(cliCtx.String("output"))
	if outputFormat == "json" {
		b, err := json.MarshalIndent(items, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "NAME", "\t", "LANGUAGE", "\t", "OS", "\t", "STATE", "\t", "CURRENT_VERSION_ID", "\n")
	for _, item := range items {
		fmt.Fprint(w,
			stringValue(item.Name), "\t",
			stringValue(item.Language), "\t",
			stringValue(item.Os), "\t",
			item.LifecycleState, "\t",
			stringValue(item.CurrentFunctionsRuntimeVersionId), "\t",
			"\n",
		)
	}
	return w.Flush()
}

func printRuntimeVersions(cliCtx *cli.Context, items []functions.FunctionsRuntimeVersionSummary) error {
	outputFormat := strings.ToLower(cliCtx.String("output"))
	if outputFormat == "json" {
		b, err := json.MarshalIndent(items, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "DISPLAY_NAME", "\t", "LANGUAGE_VERSION", "\t", "OS_VERSION", "\t", "STATE", "\t", "ID", "\n")
	for _, item := range items {
		fmt.Fprint(w,
			stringValue(item.DisplayName), "\t",
			stringValue(item.LanguageVersion), "\t",
			stringValue(item.OsVersion), "\t",
			item.LifecycleState, "\t",
			stringValue(item.Id), "\t",
			"\n",
		)
	}
	return w.Flush()
}

func printLatestRuntimeVersion(cliCtx *cli.Context, item functions.FunctionsRuntimeVersionSummary) error {
	outputFormat := strings.ToLower(cliCtx.String("output"))
	if outputFormat == "json" {
		b, err := json.MarshalIndent(item, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "DISPLAY_NAME", "\t", "LANGUAGE_VERSION", "\t", "OS_VERSION", "\t", "STATE", "\t", "ID", "\n")
	fmt.Fprint(w,
		stringValue(item.DisplayName), "\t",
		stringValue(item.LanguageVersion), "\t",
		stringValue(item.OsVersion), "\t",
		item.LifecycleState, "\t",
		stringValue(item.Id), "\t",
		"\n",
	)
	return w.Flush()
}

func suggestRuntimeNames(cliCtx *cli.Context) {
	provider, err := client.CurrentProvider()
	if err != nil {
		return
	}
	ociProvider, ok := provider.(*oracle.OracleProvider)
	if !ok || ociProvider == nil {
		return
	}
	client, err := newFunctionsClient(ociProvider)
	if err != nil {
		return
	}

	request := functions.ListFunctionsRuntimesRequest{}
	for {
		response, err := client.ListFunctionsRuntimes(context.Background(), request)
		if err != nil {
			return
		}
		for _, item := range response.Items {
			name := stringValue(item.Name)
			if name != "" {
				fmt.Println(name)
			}
		}
		if response.OpcNextPage == nil {
			break
		}
		request.Page = response.OpcNextPage
	}
}

func getRegion(oracleProvider *oracle.OracleProvider) string {
	if oracleProvider.FnApiUrl != nil {
		parts := strings.Split(oracleProvider.FnApiUrl.Host, ".")
		if len(parts) >= 4 {
			return parts[1]
		}
	}
	region, _ := oracleProvider.ConfigurationProvider.Region()
	return region
}

func newFunctionsClient(oracleProvider *oracle.OracleProvider) (functions.FunctionsManagementClient, error) {
	client, err := functions.NewFunctionsManagementClientWithConfigurationProvider(oracleProvider.ConfigurationProvider)
	if err != nil {
		return client, err
	}
	if oracleProvider.FnApiUrl != nil {
		client.Host = oracleProvider.FnApiUrl.String()
		return client, nil
	}
	client.SetRegion(getRegion(oracleProvider))
	return client, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
