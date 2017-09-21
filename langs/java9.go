package langs

import "errors"

type Java9LangHelper struct {
	BaseHelper
}

func (lh *Java9LangHelper) BuildFromImage() string {
	//FIXME: assumptions made here about dockerhub name
	//should this be a funcy: repo?
	return "fnproject/fn-java-fdk-build-1.9:latest"
}

func (lh *Java9LangHelper) RunFromImage() string {
	//FIXME: assumptions made here about dockerhub name
	return "fnproject/fn-java-fdk-1.9:latest"
}

// HasBoilerplate returns false as a stub until java9 boilerplate has been implemented
func (lh *Java9LangHelper) HasBoilerplate() bool {
	return false //FIXME @ollerhll
}

func (lh *Java9LangHelper) GenerateBoilerPlate() error {
	return errors.New("No boilerplate generation has been implemented for java9 yet.") //FIXME @ollerhll
}

// Cmd currently returns nil as boilerplate has not yet been implemented
func (lh *Java9LangHelper) Cmd() string {
	return nil //FIXME @ollerhll
}

// DockerfileBuildCmds currently returns nil as boilerplate has not yet been implemented
func (lh *Java9LangHelper) DockerfileBuildCmds() []string {
	return nil //FIXME @ollerhll
}

// HasPreBuild returns false as a stub until prebuild has been implemented
func (lh *Java9LangHelper) HasPreBuild() bool {
	return false //FIXME @ollerhll
}
