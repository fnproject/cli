# CLI End-To-End tests


These are end-to-end CLI tests that operate on a compiled CLI binary  - use the [Test Harness](../testharness/harness.go) for more details:


## Example:

The following test makes a function directory (`h.Mkdir`)  runs a few Fn commands `h.Fn(...)` and verifies the results on the returned commands `.AssertSuccess`

```
func TestPythonCall(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()
	funcName := h.NewFuncName(appName)

	h.MkDir(funcName)
	h.Cd(funcName)
	h.Fn("init", "--name", funcName, "--runtime", "python3.9").AssertSuccess()
	appName := h.NewAppName()
	h.Fn("deploy", "--local", appName).AssertSuccess()
	h.Fn("invoke", appName, funcName).AssertSuccess()
	h.FnWithInput(`{"name": "John"}`, "call", appName, funcName).AssertSuccess()

}
```

The test harness runs a specified CLI  binary(either "../fn" or "TEST_CLI_BINARY" from env) in a dynamically created test directory, $HOME is also mocked out to allow testing of configuration files.

## Hints for writing good tests:

* Don't write lots of tests for features: CLI end-to-end tests are primarily there to detect regressions in users' expectations about command behaviour - they are somewhat expensive (typically some seconds per test) so shouldn't be used as the only means to test changes - a good rule of thumb is to test the use cases that you would demonstrate to somebody when showing them the feature.
* Don't be spammy : You shouldn't log excessively in tests as this will impact diagnosability when a test fails.  Instead, always log the `CmdResult` you got from the last command that failed - this should include enough diagnostic history to work out what went wrong (including previous commands)
* Write parallelizable tests: Tests are slow so sequencing them will make the test package slow - the harness includes tools to help make tests isolated (e.g. any app names created with `h.NewFuncName()` will be deleted after a test is done )  - remember to defer `h.Cleanup()` to ensure test state is cleaned up
* Watch out for the Environment: The CLI package will pass on the surrounding environment to the CLI when its called - (primarily to allow easily overriding local  configuration like FN_API_URL and other env vars)
