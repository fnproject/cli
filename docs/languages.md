## Language Plugins

The CLI now uses docker plugins for its language support. These plugins are 
written in Golang at present.

### Support
Currently, we provide support for the following language runtimes:
* java (also referred to as java9)
* java8
* go

More are being added all the time!

### Usage
To use an officially supported language plugin, simply pass the runtime name
listed above to the `--runtime` option on `fn init`, or ensure that the 
`runtime` option in `func.yaml` has been set to the correct runtime name.

If you're using your own plugin, pass its docker image name as the runtime to
`fn init`, or ensure that the `runtime` option in `func.yaml` has been set
to the name of your docker image e.g. `ollerhll/the-best-plugin:latest`.

### What can I do?
If we don't provide the language support you want, then you have three options:
* Create your own Dockerfile for building your function.
* Raise an issue on the project asking for language support. The more people
that ask, the more likely we are to provide the support!
* Write the plugin yourself!

### How to write a language plugin
Whether you're writing a plugin intended for official support, or just writing
one for personal use, it will need to obey the following rules presented in this
section.

To start, take a look at [the template](../examples/template.go), which you should
copy, and fill out, replacing `XXXXX` with a suitable name for your plugin.

#### Core
The plugin must implement two methods, `Init(flags map[string]string)`, and
`Build(flags map[string]string)`. These need to set the `LangInitialiser` and
`LangBuilder` structs in your plugin respectively.

```go
type langInitialiser struct {
    Cmd                 string
    Entrypoint          string
}
```
```go
type langBuilder struct {
    BuildImage          string
    RunImage            string
    IsMultiStage        bool
    DockerfileCopyCmds  []string
    DockerfileBuildCmds []string
}
```
> One or both of `Cmd`/`Entrypoint` must be assigned by the `init` command, as shown
> in the template.
> `IsMultiStage`, `BuildImage`, `RunImage`, `DockerfileCopyCmds`, and `DockerfileBuildCmds`
> should be set by the `build` command, as also shown in the template.

#### Extra Parameters
The `Init` and `Build` commands take a map as an argument, which will be filled
with additional information that may be needed by your plugin. Currently, this 
just contains the `runtime` variable set by the CLI option or the option in `func.yaml`.
This is subject to change, but all changes should be backwards compatible.

#### Side Effects
`Init` is also the appropriate place to generate any boilerplate code for your runtime.
The [Java example](../examples/javaPlugin.go.example) has a good example of this.

#### Naming
<!---
CHANGE?! 
-->
If you are writing an official plugin, it should be named `fnproject/lang-NAME:latest` 

#### Walkthrough example
To help demonstrate the process, these docs will now go through the process of 
adding a Golang plugin for the CLI.

First, I'm going to fill out the `Init` method. We want to set either `Cmd` or 
`Entrypoint`, and in this case we want to generate some boilerplate code.

So first we replace
```go
entrypoint := ""
```
with
```go
entrypoint := "./func"
```
This is the command that will be called to run the go binary `func`.

Next, we want to generate some boilerplate code, so we add a `GenerateBoilerplate()` function:
```go
func generateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.go")
	if exists(codeFile) {
		return ErrBoilerplateExists
	}
	testFile := filepath.Join(wd, "test.json")
	if exists(testFile) {
		return ErrBoilerplateExists
	}

	if err := ioutil.WriteFile(codeFile, []byte(helloGoSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(testFile, []byte(goTestBoilerPlate), os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

const (
	helloGoSrcBoilerplate = //GO SOURCE CODE FOR HELLO WORLD

	goTestBoilerPlate = //GO TEST SOURCE CODE
)
```

and now `Init` is done.

Next, we need to write the `Build` method. This is much easier, as it's just writing to variables.
We replace
```go
buildImage := ""
runImage := ""
isMultiStage := true
dockerFileCopyCmds := []string{}
dockerFileBuildCmds := []string{}
```
with
```go
buildImage := "funcy/go:dev"
runImage := "funcy/go"
isMultiStage := true
dockerFileCopyCmds := []string{"COPY --from=build-stage /go/src/func/func /function/",}
dockerFileBuildCmds := []string{"ADD . /go/src/func/", "RUN cd /go/src/func/func /function/",}
```
And now we're done with the code! All that's left is to build the binary and the docker image.
