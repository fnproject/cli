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

package client

import (
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/spf13/viper"
)

func CurrentProvider() (provider.Provider, error) {
	return fn_go.DefaultProviders.ProviderFromConfig(viper.GetString(config.ContextProvider), &config.ViperConfigSource{}, &provider.TerminalPassPhraseSource{})
}
