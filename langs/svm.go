package langs

import (
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"bufio"
	"unicode"
)

type SVMLangHelper struct {
	BaseHelper
	javaHelper JavaLangHelper
}

func (svm *SVMLangHelper) Handles(language string) bool {
	for _, id := range svm.LangStrings() {
		if (id == language) {
			return true
		}
	}
	return false
}

func (svm *SVMLangHelper) Runtime() string {
	return svm.LangStrings()[0]
}

func (svm *SVMLangHelper) LangStrings() []string {
	return []string {"svm"}
}

func (svm *SVMLangHelper) Extensions() []string {
	return []string{}
}

func (svm *SVMLangHelper) BuildFromImage() (string, error) {
	return "fnproject/fn-java-svm:dev", nil
}

func (svm *SVMLangHelper) RunFromImage() (string, error) {
	return "fnproject/fn-java-svm:latest", nil
}

func (svm *SVMLangHelper) IsMultiStage() bool {
	return svm.javaHelper.IsMultiStage();
}

func (svm *SVMLangHelper) HasBoilerplate() bool {
	return true
}

func (svm *SVMLangHelper) DefaultFormat() string {
	return "http"
}

func (svm *SVMLangHelper) GenerateBoilerplate(path string) error {
	err := svm.javaHelper.GenerateBoilerplate(path);
	if err != nil {
		return err;
	}
	wd, err := os.Getwd();
	if err != nil {
		return err
	}
	pomFile := filepath.Join(wd, "pom.xml")
	err = updatePomFile(pomFile);
	if err != nil {
		return err
	}
	err = createReflectionConfigurationFile(wd)
	return err
}


func createReflectionConfigurationFile(dir string) error {
	writeFile := func(path, content string) error {
		filePath := filepath.Join(dir, path)
		ownerPath := filepath.Dir(filePath)
		if err := os.MkdirAll(ownerPath, os.FileMode(0755)); err != nil {
			return err
		}
		if err := ioutil.WriteFile(
			filePath,
			[]byte(content),
			os.FileMode(0644)); err != nil {
			return err
		}
		return nil
	}
	return writeFile("src/main/conf/reflection.json", svmReflectionConfig)
}

func (svm *SVMLangHelper) Cmd() (string, error) {
	return "com.example.fn.HelloFunction::handleRequest", nil
}

func (svm *SVMLangHelper) DockerfileCopyCmds() []string {
	return []string {
		"COPY --from=build-stage /function/target/func /function",
	}
}

func (svm *SVMLangHelper) DockerfileBuildCmds() []string {
	javaCommands := svm.javaHelper.DockerfileBuildCmds()
	svmCommands := make([]string, 0)

	for _, command := range (javaCommands) {
		if !strings.HasPrefix(command, "RUN") {
			svmCommands = append(svmCommands, command)
		}
	}
	svmCommands = append(svmCommands, "ARG MVN_PROFILE=svm")
	svmCommands = append(svmCommands, "RUN mvn -P\"${MVN_PROFILE}\" -Dgraalvm.home=\"${GRAALVM_HOME}\" package")
	return svmCommands
}


func (svm *SVMLangHelper) HasPreBuild() bool {
	return svm.javaHelper.HasPreBuild()
}

func (svm *SVMLangHelper) PreBuild() error {
	return svm.javaHelper.PreBuild()
}

func (svm *SVMLangHelper) FixImagesOnInit() bool {
	return svm.javaHelper.FixImagesOnInit()
}

func updatePomFile (pomFilePath string) error {
	content, err := updatePomContent(pomFilePath)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pomFilePath, []byte(content), 0)
}

func updatePomContent(pomFilePath string) (string, error) {
	f, err := os.Open(pomFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	newPom := ""
	inserted := false
	lastIndentation := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !inserted {
			profilesEndIndex := strings.LastIndex(line, "</profiles>")
			if profilesEndIndex >= 0 {
				newPom = newPom + indent(svmProfile, lastIndentation)
				inserted = true;
			} else {
				projectEndIndex := strings.LastIndex(line, "</project>")
				if projectEndIndex >= 0 {
					newPom = newPom + indent("<profiles>\n", lastIndentation)
					newPom = newPom + indent(svmProfile, lastIndentation + 4)
					newPom = newPom + indent("</profiles>\n", lastIndentation)
					inserted = true;
				}
			}
			lastIndentation = indentationLevel(line)
		}
		newPom = newPom + line + "\n"
	}
	return newPom, nil
}

func indentationLevel(line string) int {
	res := 0
	for index, runeValue := range line {
		res = index
		if !unicode.IsSpace(runeValue) {
			break;
		}
	}
	return res
}

func indent(str string, indentLevel int) string {
	indent := strings.Repeat(" ", indentLevel)
	res := ""
	scanner := bufio.NewScanner(strings.NewReader(str))
	for scanner.Scan() {
		res += indent + scanner.Text() + "\n"
	}
	return res
}

const (
	defaultSVMJavaSupportedVersion = "1.8"

	svmReflectionConfig= `
[
  {
    "name" : "com.example.fn.HelloFunction",
    "methods" : [
        { "name" : "handleRequest" },
        { "name" : "<init>"}
    ]
  }
]
`
	svmProfile= `<profile>
    <id>svm</id>
    <dependencies>
        <dependency>
            <groupId>com.fnproject.fn</groupId>
            <artifactId>runtime</artifactId>
            <version>${fdk.version}</version>
        </dependency>
    </dependencies>
    <build>
        <plugins>
            <plugin>
                <groupId>org.codehaus.mojo</groupId>
                <artifactId>exec-maven-plugin</artifactId>
                <version>1.6.0</version>
                <executions>
                    <execution>
                        <goals>
                            <goal>exec</goal>
                        </goals>
                        <phase>package</phase>
                    </execution>
                </executions>
                <configuration>
                    <executable>${graalvm.home}/bin/native-image</executable>
                    <arguments>
                        <argument>--static</argument>
                        <argument>-H:Name=target/func</argument>
                        <argument>-H:+ReportUnsupportedElementsAtRuntime</argument>
                        <argument>-H:ReflectionConfigurationFiles=src/main/conf/reflection.json</argument>
                        <argument>-classpath</argument>
                        <classpath/>
                        <argument>com.fnproject.fn.runtime.EntryPoint</argument>
                    </arguments>
                </configuration>
            </plugin>
        </plugins>
    </build>
</profile>
`
)
