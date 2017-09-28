## Language Plugins

The CLI now uses docker plugins for its language support. These plugins can be
written in any language capable of encoding JSON structs, as detailed below.

### Support
Currently, we provide support for the following language runtimes:
* java (also referred to as java9)
* java8

More are being added all the time!

### Usage
To use an officially supported language plugin, simply pass the runtime name
listed above to the `--runtime` option on `fn init`, or ensure that the 
`runtime` option in `func.yaml` has been set to the correct runtime name.

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

#### Input
TBD

#### Output
The plugin must emit a JSON encoded struct to fill the following Golang struct:
```go
type langHelper struct {
    BuildImage          string
    RunImage            string
    IsMultiStage        bool
    DockerfileCopyCmds  []string
    DockerfileBuildCmds []string
    Cmd                 string
    Entrypoint          string
}
```
> Note: not all of these fields must be filled at once - in fact they should not be. 
> One or both of `Cmd`/`Entrypoint` must be assigned by the `init` command. 
> `IsMultiStage`, `BuildImage`, `RunImage`, `DockerfileCopyCmds`, and `DockerfileBuildCmds`
> should be set by the `build` command.

#### Side Effects
TBD

#### Naming
TBD
