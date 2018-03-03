
# Contributing to Fn CLI

We welcome all contributions

## How to contribute

1. Fork the repo
2. Fix an issue or create an issue and fix it
3. Create a Pull Request that fixes the issue
4. Sign the [CLA](http://www.oracle.com/technetwork/community/oca-486395.html)
5. Once processed, our CLA bot will automatically clear the CLA check on the PR
6. Good Job! Thanks for being awesome!

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

### Cloning the Repository ###

Whether cloning from your Fn CLI repo or your own fork, you need to take care with the location.

`$ mkdir $GOPATH/src/github.com/fnproject/cli `

`$ cd $GOPATH/src/github.com/fnproject/cli `

Then you can either

1. Clone from the Fn CLI repo:

	`$ git clone https://github.com/fnproject/cli.git`

2. Clone from your own fork:

	`$ git clone git@github.com:<YOUR-GITHUB_USERNAME>/cli.git`

### Building ###

1.  Change to the correct directory (if not already there):

	`$ cd $GOPATH/src/github.com/fnproject/cli`

2.  Download required Go package dependencies:

	`$ make dep`

3.  Build and install:

	`$ make install`

### Testing ###

To test that you're client has built correctly:

`$ fn --version`

It should return something like:

`fn version 0.4.57`

Congratulations! You're all set :-)