package langhelper

import (
	"encoding/json"
	"flag"
	"os"
)

type LangHelperPlugin interface {
	//Main is called by main() of the binary; it should never require editing
	Main()
	//Init performs the necessary tasks for fn init - see docs
	Init(map[string]string)
	//Build performs the necessary tasks for fn build - see docs
	Build(map[string]string)
}

// BaseHelper is empty implementation of LangHelper for embedding in implementations.
type BaseHelper struct {
	LangHelpResults
	Plug LangHelperPlugin
}

type LangHelpResults struct {
	*LangBuilder
	*LangInitialiser
}

type LangInitialiser struct {
	Entrypoint string
	Cmd        string
}

type LangBuilder struct {
	BuildImage          string
	RunImage            string
	IsMultiStage        bool
	DockerfileCopyCmds  []string
	DockerfileBuildCmds []string
}

func (h *BaseHelper) Main() {
	//about a million possible improvements here
	helpCommand := flag.String("helpercommand", "", "init or build")
	rt := flag.String("runtime", "", "runtime")
	entrypoint := flag.String("entrypoint", "", "entrypoint of the function")
	cmd := flag.String("cmd", "", "command of the function")
	flag.Parse()
	if *helpCommand == "init" {
		h.Plug.Init(map[string]string{"runtime": *rt, "entrypoint": *entrypoint, "cmd": *cmd})
		h.Output(*h.LangInitialiser)
	} else if *helpCommand == "build" {
		h.Plug.Build((map[string]string{"runtime": *rt, "entrypoint": *entrypoint, "cmd": *cmd}))
		h.Output(*h.LangBuilder)
	} else {
		//reject
	}
}

func (h *BaseHelper) Init(map[string]string)  {}
func (h *BaseHelper) Build(map[string]string) {}

func (h *BaseHelper) Output(outputs interface{}) error {
	return json.NewEncoder(os.Stdout).Encode(outputs)
}

//func (h *BaseHelper) BuildFromImage() string        { return "" }
//func (h *BaseHelper) RunFromImage() string          { return h.BuildFromImage() }
//func (h *BaseHelper) IsMultiStage() bool            { return true }
//func (h *BaseHelper) DockerfileBuildCmds() []string { return []string{} }
//func (h *BaseHelper) DockerfileCopyCmds() []string  { return []string{} }
//func (h *BaseHelper) Entrypoint() string            { return "" }
//func (h *BaseHelper) Cmd() string                   { return "" }
//func (h *BaseHelper) PreBuild() error               { return nil }
//func (h *BaseHelper) AfterBuild() error             { return nil }
//func (h *BaseHelper) HasBoilerplate() bool          { return false }
//func (h *BaseHelper) GenerateBoilerplate() error    { return nil }

// exists checks if a file exists
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
