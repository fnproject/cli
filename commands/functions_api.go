package commands

import (
	"fmt"

	"github.com/fnproject/cli/common"
	apps "github.com/fnproject/cli/objects/app"
	function "github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/cli/objects/route"
	"github.com/fnproject/cli/objects/trigger"
	fnclient "github.com/fnproject/fn_go/client"
	v2Client "github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/models"
	modelsV2 "github.com/fnproject/fn_go/modelsv2"
)

func updateFunction(clientV1 *fnclient.Fn, clientV2 *v2Client.Fn, appName string, ff *common.FuncFileV20180708) error {
	fn := &modelsV2.Fn{}
	if err := function.WithFuncFileV20180708(ff, fn); err != nil {
		return fmt.Errorf("Error getting route with funcfile: %s", err)
	}

	app, err := apps.GetAppByName(appName)
	if err != nil {
		app = &models.App{
			Name: appName,
		}

		err = apps.CreateApp(clientV1, app)
		if err != nil {
			return err
		}
		app, err = apps.GetAppByName(appName)
		if err != nil {
			return err
		}
	}

	fnRes, err := function.GetFnByName(clientV2, app.ID, ff.Name)
	if err != nil {
		fn.Name = ff.Name
		err := function.CreateFn(clientV2, appName, fn)
		if err != nil {
			return err
		}
	} else {
		fn.ID = fnRes.ID
		err = function.PutFn(clientV2, fn.ID, fn)
		if err != nil {
			return err
		}
	}

	if fnRes == nil {
		fn, err = function.GetFnByName(clientV2, app.ID, ff.Name)
		if err != nil {
			return err
		}
	}

	if len(ff.Triggers) != 0 {
		for _, t := range ff.Triggers {
			trig := &modelsV2.Trigger{
				AppID:  app.ID,
				FnID:   fn.ID,
				Name:   t.Name,
				Source: t.Source,
				Type:   t.Type,
			}

			trigs, err := trigger.GetTriggerByName(clientV2, app.ID, fn.ID, t.Name)
			if err != nil {
				err = trigger.CreateTrigger(clientV2, trig)
				if err != nil {
					return err
				}
			} else {
				trig.ID = trigs.ID
				err = trigger.PutTrigger(clientV2, trig)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func updateRoute(client *fnclient.Fn, appName string, ff *common.FuncFile) error {
	rt := &models.Route{}
	if err := route.WithFuncFile(ff, rt); err != nil {
		return fmt.Errorf("Error getting route with funcfile: %s", err)
	}
	return route.PutRoute(client, appName, ff.Path, rt)
}
