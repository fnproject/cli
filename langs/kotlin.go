package langs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// KotlinLangHelper provides a set of helper methods for the lifecycle of Kotlin Maven projects
type KotlinLangHelper struct {
	BaseHelper
	latestFdkVersion string
}

func (h *KotlinLangHelper) Handles(lang string) bool {
	for _, s := range h.LangStrings() {
		if lang == s {
			return true
		}
	}
	return false
}
func (h *KotlinLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *KotlinLangHelper) LangStrings() []string {
	return []string{"kotlin"}

}
func (lh *KotlinLangHelper) Extensions() []string {
	return []string{".kt"}
}

// BuildFromImage returns the Docker image used to compile the Maven function project
func (lh *KotlinLangHelper) BuildFromImage() (string, error) {

	fdkVersion, err := lh.getFDKAPIVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("fnproject/fn-java-fdk-build:jdk9-%s", fdkVersion), nil
}

// RunFromImage returns the Docker image used to run the Kotlin function.
func (lh *KotlinLangHelper) RunFromImage() (string, error) {
	fdkVersion, err := lh.getFDKAPIVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("fnproject/fn-java-fdk:jdk9-%s", fdkVersion), nil
}

// HasBoilerplate returns whether the Java runtime has boilerplate that can be generated.
func (lh *KotlinLangHelper) HasBoilerplate() bool { return true }

// Kotlin defaults to http
func (lh *KotlinLangHelper) DefaultFormat() string { return "http" }

// GenerateBoilerplate will generate function boilerplate for a Java runtime.
// The default boilerplate is for a Maven project.
func (lh *KotlinLangHelper) GenerateBoilerplate(path string) error {
	pathToPomFile := filepath.Join(path, "pom.xml")
	if exists(pathToPomFile) {
		return ErrBoilerplateExists
	}

	apiVersion, err := lh.getFDKAPIVersion()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(pathToPomFile, []byte(kotlinPomFileContent(apiVersion)), os.FileMode(0644)); err != nil {
		return err
	}

	mkDirAndWriteFile := func(dir, filename, content string) error {
		fullPath := filepath.Join(path, dir)
		if err = os.MkdirAll(fullPath, os.FileMode(0755)); err != nil {
			return err
		}

		fullFilePath := filepath.Join(fullPath, filename)
		return ioutil.WriteFile(fullFilePath, []byte(content), os.FileMode(0644))
	}

	err = mkDirAndWriteFile("src/main/kotlin/", "HelloFunction.kt", helloKotlinSrcBoilerplate)
	if err != nil {
		return err
	}

	return mkDirAndWriteFile("src/test/kotlin/", "HelloFunctionTest.kt", helloKotlinTestBoilerplate)
}

// Cmd returns the Java runtime Docker entrypoint that will be executed when the function is executed.
func (lh *KotlinLangHelper) Cmd() (string, error) {
	return "com.fn.example.HelloFunctionKt::hello", nil
}

// DockerfileCopyCmds returns the Docker COPY command to copy the compiled Kotlin function jar and dependencies.
func (lh *KotlinLangHelper) DockerfileCopyCmds() []string {
	return []string{
		`COPY --from=build-stage /function/target/*.jar /function/app/`,
	}
}

// DockerfileBuildCmds returns the build stage steps to compile the Maven function project.
func (lh *KotlinLangHelper) DockerfileBuildCmds() []string {
	return []string{
		fmt.Sprintf(`ENV MAVEN_OPTS %s`, kotlinMavenOpts()),
		`ADD pom.xml /function/pom.xml`,
		`RUN ["mvn", "package", "dependency:copy-dependencies", "-DincludeScope=runtime", ` +
			`"-DskipTests=true", "-Dmdep.prependGroupId=true", "-DoutputDirectory=target", "--fail-never"]`,
		`ADD src /function/src`,
		`RUN ["mvn", "package"]`,
	}
}

// HasPreBuild returns whether the Java Maven runtime has a pre-build step.
func (lh *KotlinLangHelper) HasPreBuild() bool { return true }

// PreBuild ensures that the expected the function is based is a maven project.
func (lh *KotlinLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !exists(filepath.Join(wd, "pom.xml")) {
		return errors.New("Could not find pom.xml - are you sure this is a Maven project?")
	}

	return nil
}

func kotlinMavenOpts() string {
	var opts bytes.Buffer

	if parsedURL, err := url.Parse(os.Getenv("http_proxy")); err == nil {
		opts.WriteString(fmt.Sprintf("-Dhttp.proxyHost=%s ", parsedURL.Hostname()))
		opts.WriteString(fmt.Sprintf("-Dhttp.proxyPort=%s ", parsedURL.Port()))
	}

	if parsedURL, err := url.Parse(os.Getenv("https_proxy")); err == nil {
		opts.WriteString(fmt.Sprintf("-Dhttps.proxyHost=%s ", parsedURL.Hostname()))
		opts.WriteString(fmt.Sprintf("-Dhttps.proxyPort=%s ", parsedURL.Port()))
	}

	nonProxyHost := os.Getenv("no_proxy")
	opts.WriteString(fmt.Sprintf("-Dhttp.nonProxyHosts=%s ", strings.Replace(nonProxyHost, ",", "|", -1)))

	opts.WriteString("-Dmaven.repo.local=/usr/share/maven/ref/repository")

	return opts.String()
}

/*    TODO temporarily generate maven project boilerplate from hardcoded values.
      Will eventually move to using a maven archetype.*/
func kotlinPomFileContent(APIversion string) string {
	return fmt.Sprintf(kotlinPomFile, APIversion)
}

func (lh *KotlinLangHelper) getFDKAPIVersion() (string, error) {

	if lh.latestFdkVersion != "" {
		return lh.latestFdkVersion, nil
	}

	const versionURL = "https://api.bintray.com/search/packages/maven?repo=fnproject&g=com.fnproject.fn&a=fdk"
	const versionEnv = "FN_JAVA_FDK_VERSION"
	fetchError := fmt.Errorf("Failed to fetch latest Java FDK javaVersion from %v. Check your network settings or manually override the javaVersion by setting %s", versionURL, versionEnv)

	type parsedResponse struct {
		Version string `json:"latest_version"`
	}
	version := os.Getenv(versionEnv)
	if version != "" {
		return version, nil
	}
	resp, err := http.Get(versionURL)
	if err != nil || resp.StatusCode != 200 {
		return "", fetchError
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", fetchError
	}

	parsedResp := make([]parsedResponse, 1)
	err = json.Unmarshal(buf.Bytes(), &parsedResp)
	if err != nil {
		return "", fetchError
	}

	version = parsedResp[0].Version
	lh.latestFdkVersion = version
	return version, nil
}

func (lh *KotlinLangHelper) FixImagesOnInit() bool {
	return true
}

const (
	kotlinPomFile = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
		 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		 xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example.fn</groupId>
	<artifactId>hello</artifactId>
	<version>1.0.0</version>
	
	<properties>
		<kotlin.version>1.2.51</kotlin.version>
		<fdk.version>%s</fdk.version>
		<junit.version>4.12</junit.version>
		<project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
	</properties>

	<repositories>
		<repository>
			<id>fn-release-repo</id>
			<url>https://dl.bintray.com/fnproject/fnproject</url>
			<releases>
				<enabled>true</enabled>
			</releases>
			<snapshots>
				<enabled>false</enabled>
			</snapshots>
		</repository>
	</repositories>

	<dependencies>
		<dependency>
			<groupId>com.fnproject.fn</groupId>
			<artifactId>api</artifactId>
			<version>${fdk.version}</version>
		</dependency>
		<dependency>
			<groupId>org.jetbrains.kotlin</groupId>
			<artifactId>kotlin-stdlib</artifactId>
			<version>${kotlin.version}</version>
		</dependency>

		<dependency>
			<groupId>com.fnproject.fn</groupId>
			<artifactId>testing</artifactId>
			<version>${fdk.version}</version>
			<scope>test</scope>
		</dependency>
		<dependency>
			<groupId>org.jetbrains.kotlin</groupId>
			<artifactId>kotlin-test-junit</artifactId>
			<version>${kotlin.version}</version>
			<scope>test</scope>
		</dependency>
	</dependencies>
	
	<build>
		<sourceDirectory>${project.basedir}/src/main/kotlin</sourceDirectory>
		<testSourceDirectory>${project.basedir}/src/test/kotlin</testSourceDirectory>
		<plugins>
			<plugin>
				<artifactId>kotlin-maven-plugin</artifactId>
				<groupId>org.jetbrains.kotlin</groupId>
				<version>${kotlin.version}</version>
				<executions>
					<execution>
						<id>compile</id>
						<goals> <goal>compile</goal> </goals>
					</execution>
					<execution>
						<id>test-compile</id>
						<phase>compile</phase>
						<goals> <goal>test-compile</goal> </goals>
					</execution>
				</executions>
			</plugin>				
		</plugins>
	</build>
</project>
			
`

	helloKotlinSrcBoilerplate = `
package com.fn.example

fun hello(input: String) = when {
    input.isEmpty() -> ("Hello, world!")
        else -> ("Hello, ${input}")
}`

	helloKotlinTestBoilerplate = `package com.fn.example
import com.fnproject.fn.testing.*
import org.junit.*
import kotlin.test.assertEquals
	
class HelloFunctionTest {
	
	@Rule @JvmField
	val fn = FnTestingRule.createDefault()

	@Test
	fun ` + "`" + `should return default greeting` + "`" + `() {
		with (fn) {
			givenEvent().enqueue()
			thenRun("com.fn.example.HelloFunctionKt","hello")
			assertEquals("Hello, world!", getOnlyResult().getBodyAsString())
		}
	}
	
	@Test
	fun ` + "`" + `should return personalized greeting` + "`" + `() {
		with (fn) {
			givenEvent().withBody("Jhonny").enqueue()
			thenRun("com.fn.example.HelloFunctionKt","hello")
			assertEquals("Hello, Jhonny", getOnlyResult().getBodyAsString())
		}
	}
	
}`
)
