
# Contributing to Fn CLI

We welcome all contributions

## How to contribute

1. Fork the repo
2. Fix an issue or create an issue and fix it
3. Create a Pull Request that fixes the issue
4. Sign the [CLA](http://www.oracle.com/technetwork/community/oca-486395.html)
5. Once processed, our CLA bot will automatically clear the CLA check on the PR
6. Good Job! Thanks for being awesome!

## Advice for Forkers ##

We do not advise cloning from your own fork (see [Getting the Repository](#getting-the-repository) below).

Instead we suggest:
1. Clone from the original repo
2. Add your fork as a remote
3. Push your changes to your fork
4. Create your Pull Requests as normal

## Documentation

When creating a Pull Request, make sure that you also update the documentation
accordingly.

Most of the time, when making some behavior more explicit or adding a feature,
documentation update is necessary.

You will either update a file inside docs/ or create one. Prefer the former over
the latter. If you are unsure, do not hesitate to open a PR with a comment
asking for suggestions on how to address the documentation part.

## How to build and get up and running ##

### Build Dependencies ###
- [Go](https://golang.org/doc/install)
- [Dep](https://github.com/golang/dep)

### Getting the Repository ###

`$ git clone https://github.com/fnproject/cli.git $GOPATH/src/github.com/fnproject/cli`

Note that Go will require the exact path given above in order to build (beware
forkers).

### Building ###

1.  Change to the correct directory (if not already there):

	`$ cd $GOPATH/src/github.com/fnproject/cli`

2.  Download required Go package dependencies:

	`$ make dep`

3.  Build and install:

	`$ make install`

### Testing ###

To test that your client has built correctly:

`$ fn --version`

It should return something like:

`fn version 0.4.57`

Congratulations! You're all set :-)
