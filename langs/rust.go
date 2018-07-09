package langs

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type RustLangHelper struct {
	BaseHelper
}

func (h *RustLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *RustLangHelper) Runtime() string {
	return h.LangStrings()[0]
}
func (lh *RustLangHelper) LangStrings() []string {
	return []string{"rust"}
}
func (lh *RustLangHelper) Extensions() []string {
	return []string{".rs"}
}

func (lh *RustLangHelper) BuildFromImage() (string, error) {
	return "rust:1", nil
}

func (lh *RustLangHelper) RunFromImage() (string, error) {
	return "debian:stretch", nil
}

func (lh *RustLangHelper) HasBoilerplate() bool { return true }

func cargoTomlContent(username string) string {
	return `[package]
name = "func"
version = "0.1.0"
authors = ["` + username + `"]

[dependencies]
`
}

func mainContent() string {
	return `fn main() {
    println!("Hello, world!");
}
`
}

func (lh *RustLangHelper) GenerateBoilerplate(path string) error {
	username := os.Getenv("USER")
	if len(username) == 0 {
		username = "unknown"
	}

	pathToCargoToml := filepath.Join(path, "Cargo.toml")
	if exists(pathToCargoToml) {
		return ErrBoilerplateExists
	}
	if err := ioutil.WriteFile(pathToCargoToml, []byte(cargoTomlContent(username)), os.FileMode(0644)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(path, "src"), os.FileMode(0755)); err != nil {
		return err
	}
	pathToMain := filepath.Join(path, "src", "main.rs")
	if err := ioutil.WriteFile(pathToMain, []byte(mainContent()), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

func (lh *RustLangHelper) Entrypoint() (string, error) {
	return "/function/func", nil
}

func (lh *RustLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function/src/target/release/func /function/func",
	}
}

func (lh *RustLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	r = append(r, "ADD . /function/src/")
	r = append(r, "RUN cd /function/src/ && cargo build --release")
	return r
}

func (lh *RustLangHelper) HasPreBuild() bool {
	return true
}

func (lh *RustLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !exists(filepath.Join(wd, "Cargo.toml")) {
		return errors.New("Could not find Cargo.toml - are you sure this is a Rust Cargo project?")
	}

	return nil
}

func (lh *RustLangHelper) AfterBuild() error {
	return os.RemoveAll("target")
}
