# Fn CLI
[![CircleCI](https://circleci.com/gh/fnproject/cli.svg?style=svg)](https://circleci.com/gh/fnproject/cli)

## Install
MacOS installation:
```sh
brew update && brew install fn
```

or

Alternatively for Linux/Unix/MacOS:

```sh
curl -LSs https://raw.githubusercontent.com/fnproject/cli/master/install | sh
```

## General Information
* See the Fn [Quickstart](https://github.com/fnproject/fn/blob/master/README.md) for sample commands.
* [Detailed installation instructions](http://fnproject.io/tutorials/install/).
* [Configure your CLI Context](http://fnproject.io/tutorials/install/#ConfigureyourContext).
* For a list of commands see [Fn CLI Command Guide and Reference](https://github.com/fnproject/docs/blob/master/cli/README.md).
* For general information see Fn [docs](https://github.com/fnproject/docs) and [tutorials](https://fnproject.io/tutorials/).

## OCI Functions configuration
Fn CLI now supports additional OCI Functions settings for:

- **Provisioned concurrency** via `--provisioned-concurrency`
- **Detached / long-running functions** via:
  - `--detached-timeout`
  - `--on-success`
  - `--on-failure`
  - `--clear-on-success`
  - `--clear-on-failure`

These OCI Functions settings can be used through multiple Fn CLI workflows:

- `fn init` to scaffold and persist them in `func.yaml`
- `fn create function` / `fn update function` to apply them directly from the CLI
- `fn deploy` to apply values persisted in `func.yaml`

### Application networking
Create an OCI Functions application with a subnet OCID using a dedicated flag:

```sh
fn create app <app-name> --subnet-id <subnet-ocid>
```

For Oracle-backed apps, `fn create app` requires at least one subnet. `fn update app --subnet-id` is currently not supported through the current OCI update API model and returns a clear error instead.

### Resource tags
Add freeform tags using `--tag key=value`:

```sh
fn create app <app-name> --subnet-id <subnet-ocid> --tag Department=Finance
fn create function <app-name> <function-name> <image> --tag Team=Payments
```

Add defined tags using `--defined-tag namespace.key=value`:

```sh
fn create app <app-name> --subnet-id <subnet-ocid> --defined-tag dry_run_tag.example-tag=10
fn create function <app-name> <function-name> <image> --defined-tag Operations.CostCenter=42
```

By default, plain scalar defined-tag values are treated as strings. Use explicit JSON only when needed:

```sh
fn create app <app-name> --subnet-id <subnet-ocid> \
  --defined-tag 'custom.meta={"level":2}'
```

Update flows also support tag removal and clear semantics:

```sh
fn update app <app-name> --remove-tag Department --clear-defined-tags
fn update function <app-name> <function-name> --remove-defined-tag Operations.CostCenter
```

### Examples
Initialize a function with provisioned concurrency:

```sh
fn init --runtime go --provisioned-concurrency constant:40 hello
```

Create a function with provisioned concurrency directly:

```sh
fn create function <app-name> <function-name> <image> \
  --provisioned-concurrency constant:40
```

Update a function with provisioned concurrency directly:

```sh
fn update function <app-name> <function-name> \
  --provisioned-concurrency constant:40
```

Initialize a function for detached mode with OCI destinations:

```sh
fn init --runtime go \
  --detached-timeout 20m \
  --on-success stream:<stream-ocid> \
  --on-failure notifications:<topic-ocid> \
  hello
```

Create a function with detached-mode settings directly:

```sh
fn create function <app-name> <function-name> <image> \
  --detached-timeout 20m \
  --on-success stream:<stream-ocid> \
  --on-failure notifications:<topic-ocid>
```

Update a function with detached-mode settings directly:

```sh
fn update function <app-name> <function-name> \
  --detached-timeout 20m \
  --on-success stream:<stream-ocid> \
  --on-failure notifications:<topic-ocid>
```

Example `func.yaml` output:

```yaml
deploy:
  oci:
    freeform_tags:
      Department: Finance
    defined_tags:
      Operations:
        CostCenter: "42"
    provisionedConcurrency:
      strategy: CONSTANT
      count: 40
    detachedMode:
      timeout: 20m
      onSuccess:
        type: stream
        ocid: <stream-ocid>
      onFailure:
        type: notifications
        ocid: <topic-ocid>
```

When these OCI-specific flags are used with a non-Oracle provider or local Fn server workflows, Fn CLI accepts them and emits user-friendly warnings where the settings are not applicable.

### Pre-Built Function (PBF) support
List available Pre-Built Functions:

```sh
fn list pbfs
fn list pbfs --search Document
fn list pbfs --trigger http
fn list pbfs --output json
```

Get a specific PBF listing:

```sh
fn get pbfs <pbf-name-or-ocid>
```

List versions for a PBF:

```sh
fn list pbfs versions <pbf-name-or-ocid>
fn list pbfs versions <pbf-name-or-ocid> --current
```

Get a specific PBF version:

```sh
fn get pbfs version <pbf-listing-version-ocid>
```

List supported PBF trigger names:

```sh
fn list pbfs triggers
fn list pbfs triggers http
```

Create a function from a Pre-Built Function (PBF) listing OCID:

```sh
fn create function <app-name> <function-name> --pbf <pbf-listing-ocid>
```

For PBF create flows:
- `--image` and `--pbf` are mutually exclusive
- the image positional argument is optional when `--pbf` is used
- Fn CLI automatically resolves the minimum required memory from the current PBF version when possible
- if you specify `--memory`, it must be greater than or equal to the PBF minimum requirement

Persist a PBF-backed function definition using `fn init`:

```sh
fn init --name hello-pbf --pbf <pbf-listing-ocid>
```

Example `func.yaml` for a PBF-backed function:

```yaml
deploy:
  oci:
    pbf:
      listing_id: <pbf-listing-ocid>
```

Deploy a persisted PBF-backed function:

```sh
fn deploy --app <app-name>
```

For PBF-backed deploys, Fn CLI skips image build/push/sign flows and creates or updates the function using PBF source details instead.

Inspect/list output also surfaces PBF-backed functions:

```sh
fn inspect function <app-name> <function-name>
fn list functions <app-name>
```

`fn list functions` includes a `SOURCE` column, for example:

```text
NAME          IMAGE   SOURCE                                      ID
hello-pbf             pbf:<pbf-listing-ocid>                     <function-ocid>
```

### Detached invoke examples
Invoke a function in detached mode:

```sh
fn invoke detached <app-name> <function-name> --display-call-id
```

Equivalent OCI-style invoke flag:

```sh
fn invoke <app-name> <function-name> --fn-invoke-type detached --display-call-id
```

Invoke with an intent header:

```sh
fn invoke detached <app-name> <function-name> --fn-intent cloudevent --display-call-id
```

Invoke as a dry run:

```sh
fn invoke detached <app-name> <function-name> --is-dry-run --output json
```

For OCI Functions, detached invocation typically returns immediately and, when requested, prints a call ID that can be used for correlation with downstream success/failure destinations.

## CLI Development
* Refer to the [Fn CLI Wiki](https://github.com/fnproject/cli/wiki) for development details.

## Watch (local auto-deploy)
To watch a directory and automatically redeploy to a local Fn server when files change:

```sh
fn watch --app <app>
```

This watches the current directory recursively and triggers:

```sh
fn deploy --app <app> --local
```

### Ignoring paths
`fn watch` ignores these directories by default:

- `.git`, `.fn`, `node_modules`, `target`, `dist`, `vendor`, `Dockerfile-fn-tmp*`

You can add more ignore rules by creating a `.fnignore` file in the watched directory (one pattern per line; `#` comments supported), and/or by passing `--ignore` flags.

### Build from source
See [CONTRIBUTING](https://github.com/fnproject/cli/blob/master/CONTRIBUTING.md) for instructions to build the CLI from source.




