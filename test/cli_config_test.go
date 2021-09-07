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

package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestContextCrud(t *testing.T) {
	t.Parallel()
	h := testharness.Create(t)
	defer h.Cleanup()

	h.Fn("list", "context").AssertSuccess().AssertStdoutContains("default")
	h.Fn("create", "context", "--api-url", "alskjalskdjasd/v2", "mycontext").AssertFailed().AssertStderrContains("Invalid Fn API URL: does not contain ://")
	h.Fn("create", "context", "--api-url", "http://alskjalskdjasd", "mycontext").AssertSuccess()
	h.Fn("use", "context", "mycontext").AssertSuccess()
	h.Fn("update", "context", "api-url", "alskjalskdjaff/v2").AssertFailed()
	h.Fn("update", "context", "api-url").AssertFailed().AssertStderrContains("Please specify a value")
	h.Fn("update", "context", "api-url", "http://alskjalskdjaf").AssertSuccess()
	h.Fn("update", "context", "--delete").AssertFailed().AssertStderrContains("Update context files using fn update context requires the missing argument '<key>'")
	h.Fn("update", "context", "api-u", "--delete").AssertFailed().AssertStderrContains("Context file does not contain key: api-u")
	h.Fn("update", "context", "api-url", "--delete").AssertSuccess().AssertStdoutContains("Current context deleted api-url")
	h.Fn("delete", "context", "mycontext").AssertFailed().AssertStderrContains("Cannot delete the current context: mycontext")
	h.Fn("unset", "context", "mycontext").AssertSuccess()
	h.Fn("delete", "context", "mycontext").AssertSuccess()
}
