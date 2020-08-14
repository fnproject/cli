package oss

import (
	"context"
	"fmt"
	"github.com/fnproject/cli/adapter"
	oss "github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
)

type FnClient struct {
	client *oss.Fn
}

func (f FnClient) CreateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	resp, err := f.client.Fns.CreateFn(&apifns.CreateFnParams{
		Context: context.Background(),
		Body:    convertAdapterFnToV2Fn(fn),
	})

	if err != nil {
		return nil, err
	}
	return convertV2FnToAdapterFn(resp.Payload), err
}

func (f FnClient) GetFn(appID string, fnName string) (*adapter.Fn, error) {
	resp, err := f.client.Fns.ListFns(&apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &appID,
		Name:    &fnName,
	})
	if err != nil {
		return nil, err
	}

	var fn *modelsv2.Fn
	for i := 0; i < len(resp.Payload.Items); i++ {
		if resp.Payload.Items[i].Name == fnName {
			fn = resp.Payload.Items[i]
		}
	}
	if fn == nil {
		return nil, adapter.FunctionNameNotFoundError{ Name: fnName}
	}
	return convertV2FnToAdapterFn(fn), nil
}

func (f FnClient) GetFnByFnID(fnID string) (*adapter.Fn, error) {
	resp, err := f.client.Fns.GetFn(&apifns.GetFnParams{
		FnID:		fnID,
		Context: 	context.Background(),
	})

	if err != nil {
		return nil, err
	}

	return convertV2FnToAdapterFn(resp.Payload), nil
}

func (f FnClient) UpdateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	resp, err := f.client.Fns.UpdateFn(&apifns.UpdateFnParams{
		Context: context.Background(),
		FnID:   fn.ID,
		Body:   convertAdapterFnToV2Fn(fn),
	})
	if err != nil {
		switch e := err.(type) {
		case *apifns.UpdateFnBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}
	return convertV2FnToAdapterFn(resp.Payload), nil
}

func (f FnClient) DeleteFn(fnID string) error {
	params := apifns.NewDeleteFnParams()
	params.FnID = fnID
	_, err := f.client.Fns.DeleteFn(params)
	return err
}

func (f FnClient) ListFn(appID string, limit int64) ([]*adapter.Fn, error) {
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &appID,
	}

	var resFns []*adapter.Fn
	for {
		resp, err := f.client.Fns.ListFns(params)

		if err != nil {
			return nil, err
		}

		adapterFns := convertV2FnsToAdapterFns(resp.Payload.Items)
		resFns = append(resFns, adapterFns...)
		howManyMore := limit - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	return resFns, nil
}

func convertV2FnsToAdapterFns(v2Fns []*modelsv2.Fn) []*adapter.Fn {
	var resFns []*adapter.Fn
	for _, v2Fn := range v2Fns {
		resFns = append(resFns, convertV2FnToAdapterFn(v2Fn))
	}
	return resFns
}

func convertV2FnToAdapterFn(v2Fn *modelsv2.Fn) *adapter.Fn {
	resFn := adapter.Fn{
		Name: 			v2Fn.Name,
		ID: 			v2Fn.ID,
		Annotations: 	v2Fn.Annotations,
		Config: 		v2Fn.Config,
		UpdatedAt: 		v2Fn.UpdatedAt,
		CreatedAt: 		v2Fn.CreatedAt,
		AppID: 			v2Fn.AppID,
		Memory: 		v2Fn.Memory,
		IDLETimeout: 	v2Fn.IDLETimeout,
		Image: 			v2Fn.Image,
		Timeout: 		v2Fn.Timeout,
		}
	return &resFn
}

func convertAdapterFnToV2Fn(fn *adapter.Fn) *modelsv2.Fn {
	resFn := modelsv2.Fn{
		Name: 			fn.Name,
		ID: 			fn.ID,
		Annotations: 	fn.Annotations,
		Config: 		fn.Config,
		UpdatedAt: 		fn.UpdatedAt,
		CreatedAt: 		fn.CreatedAt,
		AppID: 			fn.AppID,
		Memory: 		fn.Memory,
		IDLETimeout: 	fn.IDLETimeout,
		Image: 			fn.Image,
		Timeout: 		fn.Timeout,
	}
	return &resFn
}