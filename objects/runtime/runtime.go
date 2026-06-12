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

	client, err := newFunctionsClient(ociProvider)
	if err != nil {
		return err
	}

	request := functions.ListFunctionsRuntimesRequest{}
	var items []functions.FunctionsRuntimeSummary
	for {
		response, err := client.ListFunctionsRuntimes(context.Background(), request)
		if err != nil {
			return err
		}
		items = append(items, response.Items...)
		if response.OpcNextPage == nil {
			break
		}
		request.Page = response.OpcNextPage
	}

	return printRuntimes(cliCtx, items)
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