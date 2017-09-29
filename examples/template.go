package XXXXXlanghelper //replace with "main"

import "github.com/fnproject/cli/langhelper"

type XXXXXHelper struct {
	langhelper.BaseHelper
}

func main() {
	//you probably don't want to change this
	helper := &XXXXXHelper{}
	helper.Plug = helper
	helper.Main()
}

func (h *XXXXXHelper) Init(flags map[string]string) {
	//change one or both of these values
	cmd := ""
	entrypoint := ""
	//leave this line as it is
	h.LangInitialiser = &langhelper.LangInitialiser{Cmd: cmd, Entrypoint: entrypoint}
}

func (h *XXXXXHelper) Build(flags map[string]string) {
	//change these as needed
	buildImage := ""
	runImage := ""
	isMultiStage := true
	dockerFileCopyCmds := []string{}
	dockerFileBuildCmds := []string{}
	//leave this line as it is
	h.LangBuilder = &langhelper.LangBuilder{
		BuildImage:          buildImage,
		RunImage:            runImage,
		IsMultiStage:        isMultiStage,
		DockerfileCopyCmds:  dockerFileCopyCmds,
		DockerfileBuildCmds: dockerFileBuildCmds,
	}
}
