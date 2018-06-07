# Fn CLI
[![CircleCI](https://circleci.com/gh/fnproject/cli.svg?style=svg)](https://circleci.com/gh/fnproject/cli)

## Install

```sh
brew install fn
```
or

```sh
curl -LSs https://raw.githubusercontent.com/fnproject/cli/master/install | sh
```

## Build from source

See [CONTRIBUTING](https://github.com/fnproject/cli/blob/master/CONTRIBUTING.md) for instructions to build the CLI from source.

## Quickstart

See the Fn [Quickstart](https://github.com/fnproject/fn/blob/master/README.md) for sample commands.


## Announcement

The Fn CLI command structure has changed as of version 0.4.109

Please refer to the [Fn CLI Wiki](https://github.com/fnproject/cli/wiki) page for information on why we chose this structure and for more details.

### Commands that have not changed:
```
build
bump
call
deploy
init
push
run
test
start
stop
```

### Commands that have changed:

_These nouns are now second-level commands._
```
apps
routes
calls
logs
context
```

_These verbs are now top-level commands._
```
config
create
delete
get
inspect
list
unset
update
use
```

### Commands that have been removed:
```
images 
```
As mention in [CLI Proposal](https://github.com/fnproject/cli/wiki/CLI-Proposal:--verb--noun--structure) 'All subcommands of 'fn images' exist as top-level commands, this makes the use of images redundant and will be deprecated'


### Examples:

```
fn [verb] [noun] <command>

fn config app <app> <key> <value>
fn create route <app> <route> <image>
fn delete route <app> <route>
fn get log <app> <call-id>
fn inspect route <app> <route>
fn list calls
fn unset config app <app> <key>
fn update route <app> <route>
fn use conetxt <context>
```




