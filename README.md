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




