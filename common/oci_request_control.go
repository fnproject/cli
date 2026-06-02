package common

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fnproject/fn_go/provider"
	fnprovideroracle "github.com/fnproject/fn_go/provider/oracle"
	ociCommon "github.com/oracle/oci-go-sdk/v65/common"
	ocifunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

type OCIRequestControl struct {
	IfMatch             string
	WaitForState        string
	MaxWaitSeconds      int
	WaitIntervalSeconds int
}

func ExtractOCIRequestControl(c *cli.Context) OCIRequestControl {
	return OCIRequestControl{
		IfMatch:             strings.TrimSpace(c.String("if-match")),
		WaitForState:        strings.ToUpper(strings.TrimSpace(c.String("wait-for-state"))),
		MaxWaitSeconds:      c.Int("max-wait-seconds"),
		WaitIntervalSeconds: c.Int("wait-interval-seconds"),
	}
}

func (o OCIRequestControl) HasIfMatch() bool { return o.IfMatch != "" }
func (o OCIRequestControl) HasWait() bool    { return o.WaitForState != "" }

func NormalizeWaitSettings(maxWait, interval int) (int, int) {
	if maxWait <= 0 {
		maxWait = 1200
	}
	if interval <= 0 {
		interval = 5
	}
	return maxWait, interval
}

func WarnUnsupportedOCIRequestControl(p provider.Provider, control OCIRequestControl) {
	if IsOracleProvider(p) {
		return
	}
	if control.HasIfMatch() {
		fmt.Fprintln(os.Stderr, "Warning: --if-match is only supported with an oracle provider and will be ignored.")
	}
	if control.HasWait() {
		fmt.Fprintln(os.Stderr, "Warning: wait control flags are only supported with an oracle provider and will be ignored.")
	}
}

func BuildOCIManagementClient(p provider.Provider) (*ocifunctions.FunctionsManagementClient, error) {
	oracleProvider, ok := p.(*fnprovideroracle.OracleProvider)
	if !ok || oracleProvider == nil {
		return nil, nil
	}
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

func waitUntil(deadline time.Time, interval time.Duration, check func() (bool, error)) error {
	for {
		done, err := check()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for requested state")
		}
		time.Sleep(interval)
	}
}

func WaitForAppState(p provider.Provider, appID, targetState string, maxWaitSeconds, waitIntervalSeconds int) error {
	if strings.TrimSpace(targetState) == "" || !IsOracleProvider(p) {
		return nil
	}
	client, err := BuildOCIManagementClient(p)
	if err != nil || client == nil {
		return err
	}
	maxWaitSeconds, waitIntervalSeconds = NormalizeWaitSettings(maxWaitSeconds, waitIntervalSeconds)
	deadline := time.Now().Add(time.Duration(maxWaitSeconds) * time.Second)
	interval := time.Duration(waitIntervalSeconds) * time.Second
	targetState = strings.ToUpper(strings.TrimSpace(targetState))
	return waitUntil(deadline, interval, func() (bool, error) {
		res, err := client.GetApplication(context.Background(), ocifunctions.GetApplicationRequest{ApplicationId: &appID})
		if err != nil {
			if targetState == "DELETED" {
				if serr, ok := err.(ociCommon.ServiceError); ok && serr.GetHTTPStatusCode() == 404 {
					return true, nil
				}
			}
			return false, err
		}
		return strings.EqualFold(string(res.Application.LifecycleState), targetState), nil
	})
}

func WaitForFunctionState(p provider.Provider, fnID, targetState string, maxWaitSeconds, waitIntervalSeconds int) error {
	if strings.TrimSpace(targetState) == "" || !IsOracleProvider(p) {
		return nil
	}
	client, err := BuildOCIManagementClient(p)
	if err != nil || client == nil {
		return err
	}
	maxWaitSeconds, waitIntervalSeconds = NormalizeWaitSettings(maxWaitSeconds, waitIntervalSeconds)
	deadline := time.Now().Add(time.Duration(maxWaitSeconds) * time.Second)
	interval := time.Duration(waitIntervalSeconds) * time.Second
	targetState = strings.ToUpper(strings.TrimSpace(targetState))
	return waitUntil(deadline, interval, func() (bool, error) {
		res, err := client.GetFunction(context.Background(), ocifunctions.GetFunctionRequest{FunctionId: &fnID})
		if err != nil {
			if targetState == "DELETED" {
				if serr, ok := err.(ociCommon.ServiceError); ok && serr.GetHTTPStatusCode() == 404 {
					return true, nil
				}
			}
			return false, err
		}
		return strings.EqualFold(string(res.Function.LifecycleState), targetState), nil
	})
}
