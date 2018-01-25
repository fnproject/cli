package langs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// KotlinLangHelper provides a set of helper methods for the lifecycle of Java Maven projects
type KotlinLangHelper struct {
	BaseHelper
}

func (h *KotlinLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *KotlinLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *KotlinLangHelper) Extensions() []string {
	return []string{".kt"}
}

func (lh *KotlinLangHelper) LangStrings() []string {
	return []string{"kotlin"}
}

// BuildFromImage returns the Docker image used to compile the kotlin function
func (lh *KotlinLangHelper) BuildFromImage() (string, error) {
	return "fnproject/kotlin:dev", nil
}

// RunFromImage returns the (Java) Docker image used to run the Kotlin function.
func (lh *KotlinLangHelper) RunFromImage() (string, error) {
	return "fnproject/fn-java-fdk:latest", nil
}

// HasBoilerplate returns whether the Java runtime has boilerplate that can be generated.
func (lh *KotlinLangHelper) HasBoilerplate() bool { return true }

// Java defaults to http
func (lh *KotlinLangHelper) DefaultFormat() string { return "http" }

// GenerateBoilerplate will generate function boilerplate for a Java runtime.
// project.
func (lh *KotlinLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	mkDirAndWriteFile := func(dir, filename, content string) error {
		fullPath := filepath.Join(wd, dir)
		if err = os.MkdirAll(fullPath, os.FileMode(0755)); err != nil {
			return err
		}

		fullFilePath := filepath.Join(fullPath, filename)
		return ioutil.WriteFile(fullFilePath, []byte(content), os.FileMode(0644))
	}

	err = mkDirAndWriteFile("src/main/kotlin", "Hello.kt", helloKotlinSrcBoilerplate)
	if err != nil {
		return err
	}

	testFile := filepath.Join(wd, "test.json")
	if exists(testFile) {
		return ErrBoilerplateExists
	}

	if err := ioutil.WriteFile(testFile, []byte(helloKotlinTestBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil

}

// Cmd returns the Java runtime Docker entrypoint that will be executed when the Kotlin function is executed.
func (lh *KotlinLangHelper) Cmd() (string, error) {
	return "HelloKt::hello", nil
}

// DockerfileCopyCmds returns the Docker COPY command to copy the compiled Kotlin function jar and its dependencies.
func (lh *KotlinLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function/*.jar /function/app/",
	}
}

// DockerfileBuildCmds returns the build stage steps to compile the Kotlin function project.
func (lh *KotlinLangHelper) DockerfileBuildCmds() []string {
	return []string{
		"ADD src /function/src",
		"RUN cd /function && kotlinc src/main/kotlin/Hello.kt -include-runtime -d function.jar",
	}
}

// HasPreBuild returns whether the Java Maven runtime has a pre-build step.
func (lh *KotlinLangHelper) HasPreBuild() bool { return true }

// PreBuild ensures that the expected the function is based is a maven project.
func (lh *KotlinLangHelper) PreBuild() error {
	return nil
}

func (lh *KotlinLangHelper) FixImagesOnInit() bool {
	return true
}

const (
	helloKotlinSrcBoilerplate = `class Input ( var name: String = "")
	
class Response( var message: String = "Hello World" )
	
fun hello(param: Input): Response {
	
	var response = Response()
	
	if (param.name.isNotEmpty()) {
		response.message = "Hello " + param.name.replace("\"", "") 
	}
	
	return response   
}`

	helloKotlinTestBoilerplate = `{
		"tests": [
			{
				"input": {
					"body": {
						"name": "Johnny"
					}
				},
				"output": {
					"body": {
						"message": "Hello Johnny"
					}
				}
			},
			{
				"input": {
					"body": {
						"name": ""
					}
				},
				"output": {
					"body": {
						"message": "Hello World"
					}
				}
			}
		]
	}`
)
