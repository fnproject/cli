package langs

import (
	"os"
	"os/exec"
)

type DotNetLangHelper struct {
	BaseHelper
}

func (h *DotNetLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *DotNetLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *DotNetLangHelper) LangStrings() []string {
	return []string{"dotnet"}
}
func (lh *DotNetLangHelper) Extensions() []string {
	return []string{".cs", ".fs"}
}
func (lh *DotNetLangHelper) BuildFromImage() (string, error) {
	return "microsoft/dotnet:1.0.1-sdk-projectjson", nil
}
func (lh *DotNetLangHelper) RunFromImage() (string, error) {
	return "microsoft/dotnet:runtime", nil
}

func (lh *DotNetLangHelper) Entrypoint() (string, error) {
	return "dotnet dotnet.dll", nil
}

func (lh *DotNetLangHelper) HasPreBuild() bool {
	return true
}

// PreBuild for Go builds the binary so the final image can be as small as possible
func (lh *DotNetLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"docker", "run",
		"--rm", "-v",
		wd+":/dotnet", "-w", "/dotnet", "microsoft/dotnet:1.0.1-sdk-projectjson",
		"/bin/sh", "-c", "dotnet restore && dotnet publish -c release -b /tmp -o .",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return dockerBuildError(err)
	}
	return nil
}

func (lh *DotNetLangHelper) AfterBuild() error {
	return nil
}
