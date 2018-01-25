package langs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PhpLangHelper struct {
	BaseHelper
}

func (h *PhpLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *PhpLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *PhpLangHelper) LangStrings() []string {
	return []string{"php"}
}
func (lh *PhpLangHelper) Extensions() []string {
	return []string{".php"}
}

func (lh *PhpLangHelper) BuildFromImage() (string, error) {
	return "fnproject/php:dev", nil
}

func (lh *PhpLangHelper) RunFromImage() (string, error) {
	return "fnproject/php:dev", nil
}

func (lh *PhpLangHelper) Entrypoint() (string, error) {
	return "php func.php", nil
}

func (lh *PhpLangHelper) HasPreBuild() bool {
	return true
}

func (lh *PhpLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !exists(filepath.Join(wd, "composer.json")) {
		return nil
	}

	pbcmd := fmt.Sprintf("docker run --rm -v %s:/worker -w /worker fnproject/php:dev composer install", wd)
	fmt.Println("Running prebuild command:", pbcmd)
	parts := strings.Fields(pbcmd)
	head := parts[0]
	parts = parts[1:]
	cmd := exec.Command(head, parts...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return dockerBuildError(err)
	}
	return nil
}

func (h *PhpLangHelper) IsMultiStage() bool {
	return false
}

func (h *PhpLangHelper) DockerfileBuildCmds() []string {
	return []string{"ADD . /function/"}
}
func (lh *PhpLangHelper) AfterBuild() error {
	return nil
}
