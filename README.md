# Fn CLI
[![CircleCI](https://circleci.com/gh/fnproject/cli.svg?style=svg)](https://circleci.com/gh/fnproject/cli)

## Install

```sh
curl -LSs https://raw.githubusercontent.com/fnproject/cli/master/install | sh
```

## Start Server

```sh
fn start
```

## Creating Functions

### init

Init will help you create a [function file](../docs/function-file.md) (func.yaml) in the current directory.

To make things simple, we try to use convention over configuration, so `init` will look for a file named `func.{language-extension}`. For example,
if you are using Node, put the code that you want to execute in the file `func.js`. If you are using Python, use `func.py`. Ruby, use `func.rb`. Go, `func.go`. Etc.

Run:

```sh
fn init [--name FUNCTION_NAME]
```

If you don't specify a name, the current directory name will be used.

If you want to override the convention with configuration, you can do that as well using:

```sh
fn init [--runtime node] [--entrypoint "node hello.js"] [--name FUNCTION_NAME]
```

Or, if you want full control, just make a Dockerfile. If `init` finds a Dockerfile, it will use that instead of runtime and entrypoint.

### Bump, Build, Run, Push

`fn` provides a few commands you'll use while creating and updating your functions: `bump`, `build`, `run` and `push`.

Bump will update the patch version number in your func.yaml file. Versions must be in [semver](http://semver.org/) format.

```sh
fn bump
```

To bump a major or minor version, pass the `--major` or `--minor` flag to `fn bump`. 

Build will build the image for your function, creating a Docker image tagged with the version number from func.yaml.

```sh
fn build
```

Run will help you test your function. Functions read input from STDIN, so you can pipe the payload into the function like this:

```sh
cat payload.json | fn run
```

Push will push the function image to Docker Hub.

```sh
fn push
```

## Using the API

You can interact with an `fn` server from the command line.

```sh
$ fn apps list                                  # list apps
myapp

$ fn apps create otherapp                       # create new app
otherapp created

$ fn apps inspect otherapp config               # show app-specific configuration
{ ... }

$ fn apps
myapp
otherapp

$ fn routes list myapp                          # list routes of an app
path	image
/hello	fnproject/hello

$ fn routes create otherapp /hello fnproject/hello   # create route
/hello created with fnproject/hello

$ fn routes delete otherapp hello              # delete route
/hello deleted

$ fn routes headers set otherapp hello header-name value         # add HTTP header to response
otherapp /hello headers updated header-name with value

$ fn calls list myapp /hello                     # lists all available calls for /hello route from myapp
ID: 45bd486b-6eec-548a-bc1a-94d59deef4ac
App: myapp
Route: /hello
Created At: 2017-06-02T15:23:53.263+03:00
Started At: 2017-06-02T15:23:53.263+03:00
Completed At: 2017-06-02T15:23:53.532+03:00
Status: success

$ fn calls get 45bd486b-6eec-548a-bc1a-94d59deef4ac   # gets specific calls by ID
ID: 45bd486b-6eec-548a-bc1a-94d59deef4ac
App: myapp
Route: /hello
Created At: 2017-06-02T15:23:53.263+03:00
Started At: 2017-06-02T15:23:53.263+03:00
Completed At: 2017-06-02T15:23:53.532+03:00
Status: success

$ fn version                                   # shows version both of client and server
Client version: 0.3.7
Server version 0.3.7
```

## Application level configuration

When creating an application, you can configure it to tweak its behavior and its
routes' with an appropriate flag, `config`.

Thus a more complete example of an application creation will look like:
```sh
fn apps create --config DB_URL=http://example.org/ otherapp
```

`--config` is a map of values passed to the route runtime in the form of
environment variables.

Repeated calls to `fn apps create` will trigger an update of the given
route, thus you will be able to change any of these attributes later in time
if necessary.

## Route level configuration

When creating a route, you can configure it to tweak its behavior, the possible
choices are: `memory`, `type` and `config`.

Thus a more complete example of route creation will look like:
```sh
fn routes create --memory 256 --type async --config DB_URL=http://example.org/ otherapp /hello fnproject/hello
```

You can also update existent routes configurations using the command `fn routes update`

For example:

```sh
fn routes update --memory 64 --type sync --image fnproject/hello
```

To know exactly what configurations you can update just use the command

```
fn routes update --help
```

To understand how each configuration affect your function checkout the [Definitions](/docs/definitions.md#Routes) document.

## Changing target host

`fn` is configured by default to talk http://localhost:8080.
You may reconfigure it to talk to a remote installation by updating a local
environment variable (`API_URL`):

```sh
export API_URL="http://myfunctions.example.org/"
fn ...
```

## Testing functions

If you added `tests` to the `func.yaml` file, you can have them tested using
`fn test`.

```sh
fn test
```

During local development cycles, you can easily force a build before test:

```sh
fn test -b
```

When preparing to deploy you application, remember adding `path` to `func.yaml`,
it will simplify both the creation of the route, and the execution of remote
tests:
```yaml
name: me/myapp
version: 1.0.0
path: /myfunc
```

## Other examples of usage

### Creating a new function from source

```
fn init --name hello --runtime ruby
fn deploy --app myapp
```

### Updating function

```
fn deploy --app myapp
```

### Testing function locally

```
fn run
```

### Testing route

```
fn call myapp /hello
```

### App management

```
fn apps create myapp
fn apps update myapp --headers "content-type=application/json"
fn apps config set log_level info
fn apps inspect myapp
fn apps delete myapp
```

### Route management

```
fn routes create myapp /hello fnproject/hello
# routes update will also update any changes in the func.yaml file too. 
fn routes update myapp /hello --timeout 30 --type async
fn routes config set myapp /hello log_level info
fn routes inspect myapp /hello
fn routes delete myapp /hello
```

## Contributing

You'll need Go installed and [Glide](https://github.com/Masterminds/glide) for dependencies.

```sh
glide install -v
go build -o fn
./fn
```
