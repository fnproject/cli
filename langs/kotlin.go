package langs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
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
	groupId          string
	pomType          string
}

func (h *KotlinLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}

// Runtime - return the correct runtime value for this helper.
func (h *KotlinLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *KotlinLangHelper) LangStrings() []string {
	return []string{"kotlin"}

}
func (h *KotlinLangHelper) Extensions() []string {
	return []string{".kt"}
}

// BuildFromImage returns the Docker image used to compile the Maven function project
func (h *KotlinLangHelper) BuildFromImage() (string, error) {

	fdkVersion, err := h.getFDKAPIVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("fnproject/fn-java-fdk-build:jdk11-%s", fdkVersion), nil
}

// RunFromImage returns the Docker image used to run the Kotlin function.
func (h *KotlinLangHelper) RunFromImage() (string, error) {
	fdkVersion, err := h.getFDKAPIVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("fnproject/fn-java-fdk:jre11-%s", fdkVersion), nil
}

// HasBoilerplate returns whether the Java runtime has boilerplate that can be generated.
func (h *KotlinLangHelper) HasBoilerplate() bool { return true }

// CustomMemory - no memory override here.
func (h *KotlinLangHelper) CustomMemory() uint64 { return 0 }

// GenerateBoilerplate will generate function boilerplate for a Java runtime.
// The default boilerplate is for a Maven project.
func (h *KotlinLangHelper) GenerateBoilerplate(path string) error {
	pathToPomFile := filepath.Join(path, "pom.xml")
	if exists(pathToPomFile) {
		return ErrBoilerplateExists
	}

	apiVersion, err := h.getFDKAPIVersion()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(pathToPomFile, []byte(kotlinPomFileContent(apiVersion, h.groupId, h.pomType)), os.FileMode(0644)); err != nil {
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
func kotlinPomFileContent(APIversion, groupId, pomType string) string {
	if groupId == "io.fnproject.com" || pomType == "maven" {
		return fmt.Sprintf(mavenKotlinPomFile, APIversion, groupId, groupId, groupId)
	} else {
		return fmt.Sprintf(bintrayKotlinPomFile, APIversion, groupId, groupId, groupId)
	}
}

func (lh *KotlinLangHelper) getFDKAPIVersion() (string, error) {

	if lh.latestFdkVersion != "" {
		return lh.latestFdkVersion, nil
	}

	const bintrayVersionURL = "https://api.bintray.com/search/packages/maven?repo=fnproject&g=com.fnproject.fn&a=fdk"
	const mavenComVersionUrl = "https://repo1.maven.org/maven2/com/fnproject/fn/fdk/maven-metadata.xml"
	const mavenIOVersionUrl = "https://repo1.maven.org/maven2/io/fnproject/fn/fdk/maven-metadata.xml"

	const versionEnv = "FN_JAVA_FDK_VERSION"
	fetchError := fmt.Errorf("Failed to fetch latest Java FDK javaVersion. Check your network settings or manually override the javaVersion by setting %s", versionEnv)
	version := os.Getenv(versionEnv)

	if version != "" {
		return version, nil
	}
	version, err := lh.getFDKLastestFromURL(mavenComVersionUrl, mavenIOVersionUrl, bintrayVersionURL)
	if err != nil {
		return "", fetchError
	}

	lh.latestFdkVersion = version
	return version, nil
}

func (lh *KotlinLangHelper) getFDKLastestFromURL(comURL string, ioURL string, bintrayURL string) (string, error) {
	var buf *bytes.Buffer
	var err error
	err = fmt.Errorf("All urls failed to respond ")

	//First search for com.fnproject.fn from Maven Central to get the latest version
	buf, err = lh.getURLResponse(comURL, false)
	if err == nil {
		version, e1 := lh.parseMavenResponse(*buf)
		if e1 == nil {
			lh.groupId = "com.fnproject.fn"
			lh.pomType = "maven"
			return version, e1
		}
	}

	//Second time search for io.fnproject.fn from Maven Central to get the latest version, if com.fnproject.fn fails
	buf, err = lh.getURLResponse(ioURL, false)
	if err == nil {
		version, e1 := lh.parseMavenResponse(*buf)
		if e1 == nil {
			lh.groupId = "io.fnproject.fn"
			lh.pomType = "maven"
			return version, e1
		}
	}

	//Third time search for com.fnproject.fn from Bintray to get the latest version, if both com.fnproject.fn and io.fnproject.fn fails
	buf, err = lh.getURLResponse(bintrayURL, true)
	if err == nil {
		version, e1 := lh.parseBintrayResponse(*buf)
		if e1 == nil {
			lh.groupId = "com.fnproject.fn"
			lh.pomType = "bintray"
			return version, e1
		}
	}

	//In all other case return error as latest FDK version is not identified
	return "", err
}

func (lh *KotlinLangHelper) getURLResponse(url string, inSecureSkipVerify bool) (*bytes.Buffer, error) {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	// nishalad95: bin tray TLS certs cause verification issues on OSX, skip TLS verification
	noVerifyTransport := &http.Transport{
		Proxy:                 defaultTransport.Proxy,
		DialContext:           defaultTransport.DialContext,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: inSecureSkipVerify},
	}
	client := &http.Client{Transport: noVerifyTransport}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to fetch response from URL %s Error: %v Status: %d", url, err, resp.StatusCode)
	}
	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (lh *KotlinLangHelper) parseMavenResponse(buf bytes.Buffer) (string, error) {
	type ParsedResponse struct {
		XMLName    xml.Name `xml:"metadata"`
		Text       string   `xml:",chardata"`
		GroupId    string   `xml:"groupId"`
		ArtifactId string   `xml:"artifactId"`
		Versioning struct {
			Text     string `xml:",chardata"`
			Latest   string `xml:"latest"`
			Release  string `xml:"release"`
			Versions struct {
				Text    string   `xml:",chardata"`
				Version []string `xml:"version"`
			} `xml:"versions"`
			LastUpdated string `xml:"lastUpdated"`
		} `xml:"versioning"`
	}
	var response ParsedResponse
	err := xml.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return "", err
	}

	if len(response.Versioning.Versions.Version) == 0 {
		return "", fmt.Errorf("Maven response is not valid")
	}
	version := response.Versioning.Latest
	return version, nil
}

func (lh *KotlinLangHelper) parseBintrayResponse(buf bytes.Buffer) (string, error) {
	type parsedResponse struct {
		Version string `json:"latest_version"`
	}
	parsedResp := make([]parsedResponse, 1)
	err := json.Unmarshal(buf.Bytes(), &parsedResp)
	if err != nil {
		return "", err
	}
	version := parsedResp[0].Version

	return version, nil
}

func (lh *KotlinLangHelper) FixImagesOnInit() bool {
	return true
}

const (
	mavenKotlinPomFile = `<?xml version="1.0" encoding="UTF-8"?>
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

	<dependencies>
		<dependency>
			<groupId>%s</groupId>
			<artifactId>api</artifactId>
			<version>${fdk.version}</version>
		</dependency>
		<dependency>
			<groupId>org.jetbrains.kotlin</groupId>
			<artifactId>kotlin-stdlib</artifactId>
			<version>${kotlin.version}</version>
		</dependency>

        <dependency>
            <groupId>%s</groupId>
            <artifactId>testing-core</artifactId>
            <version>${fdk.version}</version>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>%s</groupId>
            <artifactId>testing-junit4</artifactId>
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

	bintrayKotlinPomFile = `<?xml version="1.0" encoding="UTF-8"?>
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
            <artifactId>testing-core</artifactId>
            <version>${fdk.version}</version>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>com.fnproject.fn</groupId>
            <artifactId>testing-junit4</artifactId>
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

fun hello(input: String): String {
    println("Inside Kotlin Hello World function")
    return when {
        input.isEmpty() -> ("Hello, world!")
            else -> ("Hello, ${input}")
    }
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
