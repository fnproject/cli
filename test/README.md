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
	h.Fn("init", "--name", funcName, "--runtime", "python3.6").AssertSuccess()
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

## Testing against OCI APIs ##

### Pre-req: #### 
1. You need an OCI account.
2. You need OCI CLI profile setup and working with OCI Fn API. 
3. You need a VCN subnet and it subnet IDs.

### WARNING: ####
1. The integration test against OCI API may take 40-60 mins.
2. The integration tests create new container repositories that need to be manually cleaned up after each run. Approx 10 are created per run.
3. If the test is interrupted using `Ctrl+C`, you may find Fn Apps left behind that need to be manually cleaned up in addition to the container repositories.

Follow the steps below to run CLI integration tests against OCI APIs:

1. Update the `oci-test.sh` file with:
    ```
    FN_API_URL
    FN_API_URL
    FN_IMAGE
    FN_IMAGE_2
    FN_REGISTRY
    ```
2. Update the following files with appropriate values: 
    ```
    oci-auth/fn/config.yaml
    oci-auth/fn/contexts/functions-test.yaml
    oci-auth/oci/config
    simpleapp/app.json
    docker/config.json
    ```

3. Run `make oci-test`