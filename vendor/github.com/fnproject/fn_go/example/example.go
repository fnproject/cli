package main

import (
	"context"
	"fmt"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/provider"
)

func main() {

	config := provider.NewConfigSourceFromMap(map[string]string{
		"api-url": "http://localhost:8080",
	})

	currentProvider, err := fn_go.DefaultProviders.ProviderFromConfig("default", config, &provider.NopPassPhraseSource{})

	if err != nil {
		panic(err.Error())
	}

	appClient := currentProvider.APIClientv2().Apps

	ctx := context.Background()

	var cursor string
	for {
		params := &apps.ListAppsParams{
			Context: ctx,
		}
		if cursor != "" {
			params.Cursor = &cursor
		}

		gotApps, err := appClient.ListApps(params)
		if err != nil {
			panic(err.Error())
		}

		for _, app := range gotApps.Payload.Items {
			fmt.Printf("App %s\n", app.Name)
		}

		if gotApps.Payload.NextCursor == "" {
			break
		}
		cursor = gotApps.Payload.NextCursor
	}
}
