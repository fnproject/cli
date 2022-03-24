package shim

import (
	"fmt"
	"github.com/fnproject/fn_go/clientv2/triggers"
	"github.com/go-openapi/runtime"
)

type triggersShim struct{}

var _ triggers.ClientService = &triggersShim{}

var triggersUnsupportedErr = fmt.Errorf("HTTP Triggers are not supported on Oracle Functions")

func NewTriggersShim() triggers.ClientService {
	return &triggersShim{}
}

func (*triggersShim) CreateTrigger(*triggers.CreateTriggerParams) (*triggers.CreateTriggerOK, error) {
	return nil, triggersUnsupportedErr
}

func (*triggersShim) DeleteTrigger(*triggers.DeleteTriggerParams) (*triggers.DeleteTriggerNoContent, error) {
	return nil, triggersUnsupportedErr
}

func (*triggersShim) GetTrigger(*triggers.GetTriggerParams) (*triggers.GetTriggerOK, error) {
	return nil, triggersUnsupportedErr
}

func (*triggersShim) ListTriggers(*triggers.ListTriggersParams) (*triggers.ListTriggersOK, error) {
	return nil, triggersUnsupportedErr
}

func (*triggersShim) UpdateTrigger(*triggers.UpdateTriggerParams) (*triggers.UpdateTriggerOK, error) {
	return nil, triggersUnsupportedErr
}

func (*triggersShim) SetTransport(runtime.ClientTransport) {}
