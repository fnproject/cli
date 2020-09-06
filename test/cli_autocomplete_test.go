package test

import (
	"sort"
	"testing"

	"strings"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/testharness"
)

func TestTopLevelAutoComplete(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	allCmds := map[string]commands.Cmd{
		" ":       commands.Commands, //only has top level commands
		"create":  commands.CreateCmds,
		"config":  commands.ConfigCmds,
		"delete":  commands.DeleteCmds,
		"get":     commands.GetCmds,
		"inspect": commands.InspectCmds,
		"list":    commands.ListCmds,
		"unset":   commands.UnsetCmds,  //only has top level commands
		"update":  commands.UpdateCmds, //only has top level commands
		"use":     commands.UseCmds,
	}
	for cmdName, commands := range allCmds {
		var cmds []string
		for _, v := range commands {
			cmds = append(cmds, v.Name)
		}
		sort.Strings(cmds)
		expected := strings.Join(cmds, "\n")
		h.Fn(cmdName, "--generate-bash-completion").AssertStdoutContains(expected)
	}
}

// Needs to be seperate from top level commands as structure of config commands
// is slightly different than the standard verb-noun command structure
func TestTopLevelConfigAutoComplete(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	allCmds := map[string]commands.Cmd{
		"list":  commands.ConfigListCmds,
		"get":   commands.ConfigGetCmds,
		"":      commands.ConfigSetCmds,
		"unset": commands.ConfigUnsetCmds,
	}
	for cmdName, commands := range allCmds {
		var cmds []string
		for _, v := range commands {
			cmds = append(cmds, v.Name)
		}
		sort.Strings(cmds)
		expected := strings.Join(cmds, "\n")
		if cmdName == "" {
			//needs to be fudged as harness cannot handle "" as arguement for set config command
			h.Fn("config", "--generate-bash-completion").AssertStdoutContains(expected)
		} else {
			h.Fn(cmdName, "config", "--generate-bash-completion").AssertStdoutContains(expected)
		}
	}
}

func TestCreateAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	//tests begin
	//non auto-completed
	h.Fn("create", "app", "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("create", "context", "--generate-bash-completion").AssertStdoutEmpty()
	//create function
	h.Fn("create", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("create", "function", appName1, "--generate-bash-completion").AssertStdoutEmpty()
	//create trigger
	h.Fn("create", "trigger", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("create", "trigger", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("create", "trigger", appName1, funcName1, "--generate-bash-completion").AssertStdoutEmpty()
}

func TestConfigAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	configKeyApp := h.NewRandString(5)
	configValApp := h.NewRandString(5)
	configKeyFunc := h.NewRandString(5)
	configValFunc := h.NewRandString(5)

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	h.Fn("config", "app", appName1, configKeyApp, configValApp).AssertSuccess()
	h.Fn("config", "function", appName1, funcName1, configKeyFunc, configValFunc).AssertSuccess()
	//tests begin
	//set config app/function
	h.Fn("config", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("config", "app", appName1, "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("config", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("config", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("config", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutEmpty()
	//list config app/function
	h.Fn("list", "config", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("list", "config", "app", appName1, "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("list", "config", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("list", "config", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("list", "config", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutEmpty()
	//unset config app/function
	h.Fn("unset", "config", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("unset", "config", "app", appName1, "--generate-bash-completion").AssertStdoutContains(configKeyApp)
	h.Fn("unset", "config", "app", appName1, configKeyApp, "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("unset", "config", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("unset", "config", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("unset", "config", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutContains(configKeyFunc)
	h.Fn("unset", "config", "function", appName1, funcName1, configKeyFunc, "--generate-bash-completion").AssertStdoutEmpty()
	//get config app/function
	h.Fn("get", "config", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("get", "config", "app", appName1, "--generate-bash-completion").AssertStdoutContains(configKeyApp)
	h.Fn("get", "config", "app", appName1, configKeyApp, "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("get", "config", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("get", "config", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("get", "config", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutContains(configKeyFunc)
	h.Fn("get", "config", "function", appName1, funcName1, configKeyFunc, "--generate-bash-completion").AssertStdoutEmpty()

}

func TestDeleteAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	triggerName1 := h.NewTriggerName(appName1, funcName1)
	contextName1 := h.NewContextName()

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	if !h.IsOCITestMode() {
		h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/trig").AssertSuccess()
	}
	h.Fn("create", "context", contextName1).AssertSuccess()
	//tests begin
	//delete app
	h.Fn("delete", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("delete", "app", appName1, "--generate-bash-completion").AssertStdoutEmpty()
	//delete function
	h.Fn("delete", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("delete", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("delete", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutEmpty()
	if !h.IsOCITestMode() {
		//delete trigger
		h.Fn("delete", "trigger", "--generate-bash-completion").AssertStdoutContains(appName1)
		h.Fn("delete", "trigger", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
		h.Fn("delete", "trigger", appName1, funcName1, "--generate-bash-completion").AssertStdoutContains(triggerName1)
		h.Fn("delete", "trigger", appName1, funcName1, triggerName1, "--generate-bash-completion").AssertStdoutEmpty()
	}
	//delete context
	h.Fn("delete", "context", "--generate-bash-completion").AssertStdoutContains(contextName1)
	h.Fn("delete", "context", contextName1, "--generate-bash-completion").AssertStdoutEmpty()
}

func TestGetAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	//tests begin
	//No tests for Config because covered by config tests
}

func TestInspectAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	triggerName1 := h.NewTriggerName(appName1, funcName1)
	contextName1 := h.NewContextName()

	appProperties := []string{"id", "name", "updated_at", "created_at"}
	fnProperties := []string{"app_id", "created_at", "id", "idle_timeout", "image", "memory", "name", "timeout", "updated_at"}
	triggerProperties := []string{"app_id", "created_at", "id", "name", "source", "updated_at", "annotations", "fn_id", "type"}

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	if !h.IsOCITestMode() {
		h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/trig").AssertSuccess()
	}
	h.Fn("create", "context", contextName1).AssertSuccess()
	//tests begin
	//inspect app
	h.Fn("inspect", "app", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("inspect", "app", appName1, "--generate-bash-completion").AssertStdoutContainsEach(appProperties)
	h.Fn("inspect", "app", appName1, "id", "--generate-bash-completion").AssertStdoutEmpty()
	//inspect function
	h.Fn("inspect", "function", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("inspect", "function", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
	h.Fn("inspect", "function", appName1, funcName1, "--generate-bash-completion").AssertStdoutContainsEach(fnProperties)
	h.Fn("inspect", "function", appName1, funcName1, "id", "--generate-bash-completion").AssertStdoutEmpty()
	//inspect trigger
	if !h.IsOCITestMode() {
		h.Fn("inspect", "trigger", "--generate-bash-completion").AssertStdoutContains(appName1)
		h.Fn("inspect", "trigger", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
		h.Fn("inspect", "trigger", appName1, funcName1, "--generate-bash-completion").AssertStdoutContains(triggerName1)
		h.Fn("inspect", "trigger", appName1, funcName1, triggerName1, "--generate-bash-completion").AssertStdoutContainsEach(triggerProperties)
		h.Fn("inspect", "trigger", appName1, funcName1, triggerName1, "id", "--generate-bash-completion").AssertStdoutEmpty()
	}
	//inspect context
	h.Fn("inspect", "context", "--generate-bash-completion").AssertStdoutContains(contextName1)
	h.Fn("inspect", "context", contextName1, "--generate-bash-completion").AssertStdoutEmpty()
}

func TestListAutoComplete(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	triggerName1 := h.NewTriggerName(appName1, funcName1)
	contextName1 := h.NewContextName()

	withMinimalOCIApplication(h)
	h.Fn("create", "app", appName1, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, h.GetFnImage()).AssertSuccess()
	if !h.IsOCITestMode() {
		h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/trig").AssertSuccess()
	}
	h.Fn("create", "context", contextName1).AssertSuccess()
	//tests begin
	//non auto-completed
	h.Fn("list", "apps", "--generate-bash-completion").AssertStdoutEmpty()
	h.Fn("list", "contexts", "--generate-bash-completion").AssertStdoutEmpty()
	//list functions
	h.Fn("list", "functions", "--generate-bash-completion").AssertStdoutContains(appName1)
	h.Fn("list", "functions", appName1, "--generate-bash-completion").AssertStdoutEmpty()
	//list triggers
	if !h.IsOCITestMode() {
		h.Fn("list", "triggers", "--generate-bash-completion").AssertStdoutContains(appName1)
		h.Fn("list", "triggers", appName1, "--generate-bash-completion").AssertStdoutContains(funcName1)
		h.Fn("list", "triggers", appName1, funcName1, "--generate-bash-completion").AssertStdoutEmpty()
	}
	//No tests for listing Config because covered by config tests
}

func TestUseAutoComplete(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	contextName1 := h.NewContextName()
	h.Fn("create", "context", contextName1).AssertSuccess()
	//tests begin
	//use context
	h.Fn("use", "--generate-bash-completion").AssertStdoutContains("context")
	h.Fn("use", "context", "--generate-bash-completion").AssertStdoutContains(contextName1)
	h.Fn("use", "context", contextName1, "--generate-bash-completion").AssertStdoutEmpty()
}
