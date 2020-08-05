package oss

import (
	"context"
	"fmt"
	"github.com/fnproject/cli/adapter"
	oss "github.com/fnproject/fn_go/clientv2"
	apiTriggers "github.com/fnproject/fn_go/clientv2/triggers"
	"github.com/fnproject/fn_go/modelsv2"
)

type TriggerClient struct {
	client *oss.Fn
}

func (t TriggerClient) CreateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	resp, err := t.client.Triggers.CreateTrigger(&apiTriggers.CreateTriggerParams{
		Context: context.Background(),
		Body:    convertAdapterTrToV2Tr(trig),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiTriggers.CreateTriggerBadRequest:
			fmt.Println(e)
			return nil, fmt.Errorf("%s", e.Payload.Message)
		case *apiTriggers.CreateTriggerConflict:
			return nil, fmt.Errorf("%s", e.Payload.Message)
		default:
			return nil, err
		}
	}

	fmt.Println("Successfully created trigger:", resp.Payload.Name)
	endpoint := resp.Payload.Annotations["fnproject.io/trigger/httpEndpoint"]
	fmt.Println("Trigger Endpoint:", endpoint)

	return convertV2TrToAdapterTr(resp.Payload), nil
}

func (t TriggerClient) GetTrigger(appID string, fnID string, trigName string) (*adapter.Trigger, error) {
	triggerList, err := t.client.Triggers.ListTriggers(&apiTriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &appID,
		FnID:    &fnID,
		Name:    &trigName,
	})

	if err != nil {
		return nil, err
	}
	if len(triggerList.Payload.Items) == 0 {
		return nil, adapter.TriggerNameNotFoundError{Name: trigName}
	}

	return convertV2TrToAdapterTr(triggerList.Payload.Items[0]), nil
}

func (t TriggerClient) UpdateTrigger(trig *adapter.Trigger) (*adapter.Trigger, error) {
	res, err := t.client.Triggers.UpdateTrigger(&apiTriggers.UpdateTriggerParams{
		Context:   context.Background(),
		TriggerID: trig.ID,
		Body:      convertAdapterTrToV2Tr(trig),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiTriggers.UpdateTriggerBadRequest:
			return nil, fmt.Errorf("%s", e.Payload.Message)
		default:
			return nil, err
		}
	}
	return convertV2TrToAdapterTr(res.Payload), nil
}

func (t TriggerClient) DeleteTrigger(trigID string) error {
	params := apiTriggers.NewDeleteTriggerParams()
	params.TriggerID = trigID
	_, err := t.client.Triggers.DeleteTrigger(params)
	return err
}

func (t TriggerClient) ListTrigger(appID string, fnID string, limit int64) ([]*adapter.Trigger, error) {
	var params *apiTriggers.ListTriggersParams

	if len(fnID) == 0 {
		params = &apiTriggers.ListTriggersParams{
			Context: context.Background(),
			AppID:   &appID,
		}

	} else {
		params = &apiTriggers.ListTriggersParams{
			Context: context.Background(),
			AppID:   &appID,
			FnID:    &fnID,
		}
	}

	var resTriggers []*adapter.Trigger
	for {

		resp, err := t.client.Triggers.ListTriggers(params)
		if err != nil {
			return nil, err
		}

		adapterTriggers := convertV2TrsToAdapterTrs(resp.Payload.Items)
		resTriggers = append(resTriggers, adapterTriggers...)
		howManyMore := limit - int64(len(resTriggers)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	return resTriggers, nil
}

func convertV2TrsToAdapterTrs(v2Trs []*modelsv2.Trigger) []*adapter.Trigger {
	var resTrs []*adapter.Trigger
	for _, v2Tr := range v2Trs {
		resTrs = append(resTrs, convertV2TrToAdapterTr(v2Tr))
	}
	return resTrs
}

func convertV2TrToAdapterTr(v2Tr *modelsv2.Trigger) *adapter.Trigger {
	resTr := adapter.Trigger{
		AppID:			v2Tr.AppID,
		ID: 			v2Tr.ID,
		FnID: 			v2Tr.FnID,
		CreatedAt: 		v2Tr.CreatedAt,
		UpdatedAt: 		v2Tr.UpdatedAt,
		Annotations: 	v2Tr.Annotations,
		Name: 			v2Tr.Name,
		Source: 		v2Tr.Source,
		Type: 			v2Tr.Type,
	}
	return &resTr
}

func convertAdapterTrToV2Tr(tr *adapter.Trigger) *modelsv2.Trigger {
	resTr := modelsv2.Trigger{
		AppID:			tr.AppID,
		ID: 			tr.ID,
		FnID: 			tr.FnID,
		CreatedAt: 		tr.CreatedAt,
		UpdatedAt: 		tr.UpdatedAt,
		Annotations: 	tr.Annotations,
		Name: 			tr.Name,
		Source: 		tr.Source,
		Type: 			tr.Type,
	}
	return &resTr
}