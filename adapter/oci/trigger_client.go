package oci

import (
	"fmt"
	"github.com/fnproject/cli/adapter"
)

type TriggerClient struct {
}

// TriggerNotSupportedError error for unsupported trigger operations
type TriggerNotSupportedError struct {
}

func (n TriggerNotSupportedError) Error() string {
	return fmt.Sprintf("HTTP Triggers are not supported on Oracle Functions.")
}

func (t TriggerClient) CreateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	return nil, TriggerNotSupportedError{}
}

func (t TriggerClient) GetTrigger(appID string, fnID string, trigName string) (*adapter.Trigger, error) {
	return nil, TriggerNotSupportedError{}
}

func (t TriggerClient) UpdateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	return nil, TriggerNotSupportedError{}
}

func (t TriggerClient) DeleteTrigger(trigID string) error {
	return TriggerNotSupportedError{}
}

func (t TriggerClient) ListTrigger(appID string, fnID string, limit int64) ([]*adapter.Trigger, error) {
	return nil, TriggerNotSupportedError{}
}
