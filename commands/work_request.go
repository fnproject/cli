package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cliClient "github.com/fnproject/cli/client"
	appobj "github.com/fnproject/cli/objects/app"
	fnobj "github.com/fnproject/cli/objects/fn"
	v2Client "github.com/fnproject/fn_go/clientv2"
	fnprovider "github.com/fnproject/fn_go/provider/oracle"
	ociFunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

type workRequestCmd struct {
	provider *fnprovider.OracleProvider
	clientV2 *v2Client.Fn
}

type workRequestStatusView struct {
	WorkRequestID string
	FunctionName  string
	FunctionID    string
	Operation     string
	Status        string
	Error         string
	RecentLogs    []string
}

type workRequestErrorView struct {
	Timestamp string
	Code      string
	Message   string
}

func WorkRequestCommand() cli.Command {
	cmd := &workRequestCmd{}
	return cli.Command{
		Name:     "work-request",
		Usage:    "\tInspect Functions work requests",
		Aliases:  []string{"wr"},
		Category: "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			provider, err := cliClient.CurrentProvider()
			if err != nil {
				return err
			}
			oracleProvider, ok := provider.(*fnprovider.OracleProvider)
			if !ok || oracleProvider == nil {
				return errors.New("work-request commands require an oracle provider")
			}
			cmd.provider = oracleProvider
			cmd.clientV2 = provider.APIClientv2()
			return nil
		},
		Subcommands: []cli.Command{
			{
				Name:      "status",
				Usage:     "\tShow consolidated status, error, and recent logs for a work request or function",
				ArgsUsage: "<work-request-id|function-name>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "app",
						Usage: "App name when resolving the latest work request for a function name",
					},
					cli.IntFlag{
						Name:  "log-limit",
						Usage: "Number of recent work request logs to show",
						Value: 5,
					},
				},
				Action: cmd.status,
			},
			{
				Name:      "error",
				Usage:     "\tList errors reported by a Functions work request",
				ArgsUsage: "<work-request-id|function-name>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "app",
						Usage: "App name when resolving the latest work request for a function name",
					},
				},
				Action: cmd.listErrors,
			},
		},
	}
}

func (w *workRequestCmd) status(c *cli.Context) error {
	target := strings.TrimSpace(c.Args().First())
	if target == "" {
		return errors.New("work request id or function name is required")
	}
	logLimit := c.Int("log-limit")
	if logLimit <= 0 {
		logLimit = 5
	}
	workRequestID, functionHint, err := w.resolveWorkRequestTarget(c, target)
	if err != nil {
		return err
	}
	view, err := w.loadWorkRequestStatus(workRequestID, functionHint, logLimit)
	if err != nil {
		return err
	}
	printWorkRequestStatusView(view)
	return nil
}

func (w *workRequestCmd) listErrors(c *cli.Context) error {
	target := strings.TrimSpace(c.Args().First())
	if target == "" {
		return errors.New("work request id or function name is required")
	}
	workRequestID, _, err := w.resolveWorkRequestTarget(c, target)
	if err != nil {
		return err
	}
	items, err := w.loadWorkRequestErrors(workRequestID)
	if err != nil {
		return err
	}
	printWorkRequestErrors(workRequestID, items)
	return nil
}

func (w *workRequestCmd) resolveWorkRequestTarget(c *cli.Context, target string) (string, string, error) {
	if isWorkRequestID(target) {
		return target, "", nil
	}
	appName := strings.TrimSpace(c.String("app"))
	if appName == "" {
		return "", "", errors.New("--app is required when resolving a function name to its latest work request")
	}
	appRes, err := appobj.GetAppByName(w.clientV2, appName)
	if err != nil {
		return "", "", err
	}
	fnRes, err := fnobj.GetFnByName(w.clientV2, appRes.ID, target)
	if err != nil {
		return "", "", err
	}
	wrClient, err := buildCLIWorkRequestClient(w.provider)
	if err != nil {
		return "", "", err
	}
	limit := 1
	resp, err := wrClient.ListWorkRequests(context.Background(), ociFunctions.ListWorkRequestsRequest{
		CompartmentId: &w.provider.CompartmentID,
		ResourceId:    &fnRes.ID,
		Limit:         &limit,
		SortBy:        ociFunctions.ListWorkRequestsSortByTimeaccepted,
		SortOrder:     ociFunctions.ListWorkRequestsSortOrderDesc,
	})
	if err != nil {
		return "", "", err
	}
	if len(resp.Items) == 0 || resp.Items[0].Id == nil {
		return "", "", fmt.Errorf("no work requests found for function %s", target)
	}
	return *resp.Items[0].Id, fnRes.Name, nil
}

func isWorkRequestID(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "ocid1.") && strings.Contains(lower, "workrequest.")
}

func (w *workRequestCmd) loadWorkRequestStatus(workRequestID, functionHint string, logLimit int) (*workRequestStatusView, error) {
	wrClient, err := buildCLIWorkRequestClient(w.provider)
	if err != nil {
		return nil, err
	}
	resp, err := wrClient.GetWorkRequest(context.Background(), ociFunctions.GetWorkRequestRequest{WorkRequestId: &workRequestID})
	if err != nil {
		return nil, err
	}
	view := &workRequestStatusView{
		WorkRequestID: workRequestID,
		Operation:     simplifyWorkRequestOperation(resp.OperationType),
		Status:        string(resp.Status),
		FunctionName:  functionHint,
	}
	functionID := extractFunctionResourceID(resp.WorkRequest)
	view.FunctionID = functionID
	if view.FunctionName == "" && functionID != "" {
		if name, err := lookupFunctionName(w.provider, functionID); err == nil {
			view.FunctionName = name
		}
	}
	if view.FunctionName == "" && functionID != "" {
		view.FunctionName = functionID
	}

	if errResp, err := wrClient.ListWorkRequestErrors(context.Background(), ociFunctions.ListWorkRequestErrorsRequest{
		WorkRequestId: &workRequestID,
		Limit:         intPointer(1),
		SortBy:        ociFunctions.ListWorkRequestErrorsSortByTimestamp,
		SortOrder:     ociFunctions.ListWorkRequestErrorsSortOrderDesc,
	}); err == nil && len(errResp.Items) > 0 && errResp.Items[0].Message != nil {
		view.Error = *errResp.Items[0].Message
	}

	logsResp, err := wrClient.ListWorkRequestLogs(context.Background(), ociFunctions.ListWorkRequestLogsRequest{
		WorkRequestId: &workRequestID,
		Limit:         intPointer(logLimit),
		SortBy:        ociFunctions.ListWorkRequestLogsSortByTimestamp,
		SortOrder:     ociFunctions.ListWorkRequestLogsSortOrderDesc,
	})
	if err == nil && len(logsResp.Items) > 0 {
		for i := len(logsResp.Items) - 1; i >= 0; i-- {
			if logsResp.Items[i].Message != nil {
				view.RecentLogs = append(view.RecentLogs, *logsResp.Items[i].Message)
			}
		}
	}

	return view, nil
}

func (w *workRequestCmd) loadWorkRequestErrors(workRequestID string) ([]workRequestErrorView, error) {
	wrClient, err := buildCLIWorkRequestClient(w.provider)
	if err != nil {
		return nil, err
	}

	const pageSize = 50
	request := ociFunctions.ListWorkRequestErrorsRequest{
		WorkRequestId: &workRequestID,
		Limit:         intPointer(pageSize),
		SortBy:        ociFunctions.ListWorkRequestErrorsSortByTimestamp,
		SortOrder:     ociFunctions.ListWorkRequestErrorsSortOrderDesc,
	}
	var result []workRequestErrorView
	for {
		response, err := wrClient.ListWorkRequestErrors(context.Background(), request)
		if err != nil {
			return nil, err
		}
		for _, item := range response.Items {
			view := workRequestErrorView{
				Code:    optionalString(item.Code),
				Message: optionalString(item.Message),
			}
			if item.Timestamp != nil {
				view.Timestamp = item.Timestamp.Time.Format(time.RFC3339)
			}
			result = append(result, view)
		}
		if response.OpcNextPage == nil || strings.TrimSpace(*response.OpcNextPage) == "" {
			break
		}
		request.Page = response.OpcNextPage
	}
	return result, nil
}

func buildCLIWorkRequestClient(provider *fnprovider.OracleProvider) (*ociFunctions.WorkRequestManagementClient, error) {
	client, err := ociFunctions.NewWorkRequestManagementClientWithConfigurationProvider(provider.ConfigurationProvider)
	if err != nil {
		return nil, err
	}
	if provider.FnApiUrl != nil {
		client.Host = provider.FnApiUrl.String()
	} else if region := getRegion(provider); region != "" {
		client.SetRegion(region)
	}
	return &client, nil
}

func buildCLIFunctionsManagementClient(provider *fnprovider.OracleProvider) (*ociFunctions.FunctionsManagementClient, error) {
	client, err := ociFunctions.NewFunctionsManagementClientWithConfigurationProvider(provider.ConfigurationProvider)
	if err != nil {
		return nil, err
	}
	if provider.FnApiUrl != nil {
		client.Host = provider.FnApiUrl.String()
	} else if region := getRegion(provider); region != "" {
		client.SetRegion(region)
	}
	return &client, nil
}

func lookupFunctionName(provider *fnprovider.OracleProvider, functionID string) (string, error) {
	client, err := buildCLIFunctionsManagementClient(provider)
	if err != nil {
		return "", err
	}
	resp, err := client.GetFunction(context.Background(), ociFunctions.GetFunctionRequest{FunctionId: &functionID})
	if err != nil {
		return "", err
	}
	if resp.DisplayName == nil {
		return "", errors.New("function display name not found")
	}
	return *resp.DisplayName, nil
}

func extractFunctionResourceID(workRequest ociFunctions.WorkRequest) string {
	for _, resource := range workRequest.Resources {
		entityType := ""
		if resource.EntityType != nil {
			entityType = strings.ToLower(strings.TrimSpace(*resource.EntityType))
		}
		if strings.Contains(entityType, "function") && resource.Identifier != nil {
			return *resource.Identifier
		}
	}
	for _, resource := range workRequest.Resources {
		if resource.Identifier != nil {
			return *resource.Identifier
		}
	}
	return ""
}

func simplifyWorkRequestOperation(operation ociFunctions.OperationTypeEnum) string {
	value := strings.TrimSpace(string(operation))
	if value == "" {
		return "UNKNOWN"
	}
	if idx := strings.Index(value, "_"); idx != -1 {
		return value[:idx]
	}
	return value
}

func printWorkRequestStatusView(view *workRequestStatusView) {
	fmt.Printf("Work Request: %s\n", view.WorkRequestID)
	if view.FunctionName != "" {
		fmt.Printf("Function: %s\n", view.FunctionName)
	}
	if view.Operation != "" {
		fmt.Printf("Operation: %s\n", view.Operation)
	}
	fmt.Printf("Status: %s\n", view.Status)
	if view.Error != "" {
		fmt.Printf("Error: %s\n", view.Error)
	}
	if len(view.RecentLogs) > 0 {
		fmt.Println("Recent Logs:")
		for _, entry := range view.RecentLogs {
			fmt.Printf("- %s\n", entry)
		}
	}
}

func printWorkRequestErrors(workRequestID string, items []workRequestErrorView) {
	fmt.Printf("Work Request: %s\n", workRequestID)
	if len(items) == 0 {
		fmt.Println("No errors found.")
		return
	}
	fmt.Println("Errors:")
	for _, item := range items {
		prefix := ""
		if item.Timestamp != "" {
			prefix = item.Timestamp + " "
		}
		if item.Code != "" {
			prefix += item.Code + ": "
		}
		fmt.Printf("- %s%s\n", prefix, item.Message)
	}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intPointer(v int) *int {
	return &v
}
