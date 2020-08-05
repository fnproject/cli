package oci

import "github.com/fnproject/cli/adapter"

type TriggerClient struct {
}

func (t TriggerClient) CreateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	//TODO: call OCI client
	return nil, nil
}

func (t TriggerClient) GetTrigger(appID string, fnID string, trigName string) (*adapter.Trigger, error) {
	//TODO: call OCI client
	return nil, nil
}

func (t TriggerClient) UpdateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	//TODO: call OCI client
	return nil, nil
}

func (t TriggerClient) DeleteTrigger(trigID string) error {
	//TODO: call OCI client
	return nil
}

func (t TriggerClient) ListTrigger(appID string, fnID string, limit int64) ([]*adapter.Trigger, error) {
	//TODO: call OCI client
	return nil, nil
}