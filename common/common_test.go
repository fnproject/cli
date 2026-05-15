/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"fmt"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/cli/langs"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

type ShellCommanderFactory func(name string, arg ...string) ShellCommand

type mockShellCommand struct {
	ShellCommand
	mock.Mock
	args         []string
	stdoutWriter io.Writer
}

func (m *mockShellCommand) Start() error                     { return m.Called().Error(0) }
func (m *mockShellCommand) Run() error                       { return m.Called().Error(0) }
func (m *mockShellCommand) Wait() error                      { return m.Called().Error(0) }
func (m *mockShellCommand) Kill() error                      { return m.Called().Error(0) }
func (m *mockShellCommand) SetStdOut(stdoutWriter io.Writer) { m.stdoutWriter = stdoutWriter }
func (m *mockShellCommand) SetStdErr(io.Writer)              {}

func TestValidateImageName(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr string
	}{
		{name: "docker.io/sally/img:0.0.1", expectedErr: ""},
		{name: "sally/img:0.0.1", expectedErr: ""},
		{name: "img:0.0.1", expectedErr: "image name must have a dockerhub owner or private registry. Be sure to set FN_REGISTRY env var, pass in --registry or configure your context file"},
		{name: "owner/img", expectedErr: "image name must have a tag"},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			errString := ""
			if err := ValidateFullImageName(c.name); err != nil {
				errString = err.Error()
			}
			if c.expectedErr != errString {
				t.Fatalf("expected %s but got %s", c.expectedErr, errString)
			}
		})
	}
}

func Test_proxyArgs(t *testing.T) {
	tests := []struct {
		name string
		set  []string
		want []string
	}{
		{"empty", []string{}, []string{}},
		{"populated", []string{"http_proxy", "https_proxy", "no_proxy", "foo"}, []string{
			"-e", "http_proxy=value_of_http_proxy",
			"-e", "https_proxy=value_of_https_proxy",
			"-e", "no_proxy=value_of_no_proxy"}},
		{"partial", []string{"http_proxy", "no_proxy", "foo"}, []string{
			"-e", "http_proxy=value_of_http_proxy",
			"-e", "no_proxy=value_of_no_proxy"}},
	}
	for _, tt := range tests {
		old := map[string]string{
			"http_proxy":  "",
			"https_proxy": "",
			"no_proxy":    "",
			"foo":         "",
		}
		for k, _ := range old {
			old[k] = os.Getenv(k)
			_ = os.Unsetenv(k)
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range tt.set {
				_ = os.Setenv(k, "value_of_"+k)
			}
			if got := proxyArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proxyArgs() = %v, want %v", got, tt.want)
			}
		})
		for k, v := range old {
			_ = os.Setenv(k, v)
		}
	}
}

func Test_writeTmpDockerfileV20180708(t *testing.T) {
	defer func() { ShellCommander = newExecShellCommander }()
	dir, _ := os.MkdirTemp("", fmt.Sprintf("%s_*", t.Name()))
	viper.SetDefault(config.ContainerEngineType, "docker")
	type args struct {
		helper                 langs.LangHelper
		dir                    string
		ff                     *FuncFileV20180708
		localDebug             bool
		baseFdkImageEntrypoint string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"python-debug-image",
			args{&langs.PythonLangHelper{Version: "3.12"}, dir, &pythonFuncFile, true, ""},
			pythonDebugDockerfile,
		},
		{"python-normal-image",
			args{&langs.PythonLangHelper{Version: "3.12"}, dir, &pythonFuncFile, false, ""},
			pythonDockerfile,
		},
		{"go-debug-image",
			args{&langs.GoLangHelper{Version: "1.24"}, dir, &goFuncFile, true, ""},
			goDebugDockerfile,
		},
		{"go-normal-image",
			args{&langs.GoLangHelper{Version: "1.24"}, dir, &goFuncFile, false, ""},
			goDockerfile,
		},
		{"java-debug-image",
			args{&langs.JavaLangHelper{Version: "17"}, dir, &javaFuncFile, true, javaFdkEntryPoint},
			fmt.Sprintf(javaDebugDockerfile, langs.MavenOptsForTest()),
		},
		{"java-normal-image",
			args{&langs.JavaLangHelper{Version: "17"}, dir, &javaFuncFile, false, javaFdkEntryPoint},
			fmt.Sprintf(javaDockerfile, langs.MavenOptsForTest()),
		},
		{"node-normal-image",
			args{&langs.NodeLangHelper{Version: "24"}, dir, &nodeFuncFile, false, ""},
			nodeDockerfile,
		},
		{"ruby-normal-image",
			args{&langs.RubyLangHelper{Version: "3.3"}, dir, &rubyFuncFile, false, ""},
			rubyDockerfile,
		},
		{"dotnet-normal-image",
			args{&langs.DotnetLangHelper{Version: "9.0"}, dir, &dotnetFuncFile, false, ""},
			dotnetDockerfile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			pullImageMockCommand := &mockShellCommand{}
			pullImageMockCommand.On("Start").Return(nil)
			pullImageMockCommand.On("Wait").Return(nil)
			inspectImageMockCommand := &mockShellCommand{}
			inspectImageMockCommand.On("Run").Run(
				func(args mock.Arguments) {
					_, _ = inspectImageMockCommand.stdoutWriter.Write([]byte(tt.args.baseFdkImageEntrypoint))
				},
			).Return(nil)
			ShellCommander = createMockShellCommander(t, pullImageMockCommand, inspectImageMockCommand)

			// Run
			actualDockerfile, err := writeTmpDockerfileV20180708(tt.args.helper, tt.args.dir, tt.args.ff, tt.args.localDebug)
			if err != nil {
				t.Errorf("writeTmpDockerfileV20180708() error = %v", err)
				return
			}

			// verify dockerfile content
			content, err := os.ReadFile(actualDockerfile)
			if err != nil {
				t.Errorf("Failed to read the generated dockerfile: %s. Error: %v", actualDockerfile, err)
			}
			if string(content) != tt.want {
				t.Errorf("writeTmpDockerfileV20180708() dockerfile content = %v, want %v", string(content), tt.want)
			}
			assert.Equal(t, []string{"pull", tt.args.ff.Run_image}, pullImageMockCommand.args)
			pullImageMockCommand.AssertCalled(t, "Start")
			pullImageMockCommand.AssertCalled(t, "Wait")
			assert.Equal(t, []string{"inspect", "-f", "'{{.Config.Entrypoint}}'", tt.args.ff.Run_image}, inspectImageMockCommand.args)
			inspectImageMockCommand.AssertCalled(t, "Run")
		})
	}
}

func createMockShellCommander(t *testing.T, pullImageMockCommand *mockShellCommand, inspectImageMockCommand *mockShellCommand) ShellCommanderFactory {
	return func(name string, arg ...string) ShellCommand {
		if name != "docker" {
			t.Errorf("Unexpected shell command: %s", name)
		}
		switch arg[0] {
		case "pull":
			pullImageMockCommand.args = arg
			return pullImageMockCommand
		case "inspect":
			inspectImageMockCommand.args = arg
			return inspectImageMockCommand
		default:
			t.Errorf("Unexpected docker command: %s", arg[0])
		}
		return nil
	}
}

var pythonFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "python",
	Build_image:    "fnproject/python:3.12-dev",
	Run_image:      "fnproject/python:3.12",
	Cmd:            "",
	Entrypoint:     "/python/bin/fdk /function/func.py handler",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

var goFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "go",
	Build_image:    "fnproject/go:1.24-dev",
	Run_image:      "fnproject/go:1.24",
	Cmd:            "",
	Entrypoint:     "./func",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

var javaFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "java",
	Build_image:    "fnproject/fn-java-fdk-build:jdk17-1.1.7",
	Run_image:      "fnproject/fn-java-fdk-build:jre17-1.1.7",
	Cmd:            "com.example.fn.HelloFunction::handleRequest",
	Entrypoint:     "",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

var nodeFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "node",
	Build_image:    "fnproject/node:24-dev",
	Run_image:      "fnproject/node:24",
	Cmd:            "",
	Entrypoint:     "node func.js",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

var rubyFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "ruby",
	Build_image:    "fnproject/ruby:3.3-dev",
	Run_image:      "fnproject/ruby:3.3",
	Cmd:            "",
	Entrypoint:     "ruby func.rb",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

var dotnetFuncFile = FuncFileV20180708{
	Schema_version: 20180708,
	Name:           "test",
	Version:        "0.0.1",
	Runtime:        "dotnet",
	Build_image:    "fnproject/dotnet:9.0-1.0.57-dev",
	Run_image:      "fnproject/dotnet:9.0-1.0.57",
	Cmd:            "Function:Greeter:greet",
	Entrypoint:     "dotnet Function.dll",
	Content_type:   "",
	Type:           "",
	Memory:         128,
	Timeout:        nil,
	IDLE_timeout:   nil,
	Config:         nil,
	Annotations:    nil,
	Build:          nil,
	Expects:        Expects{},
	Triggers:       nil,
}

const (
	pythonDebugDockerfile = `FROM fnproject/python:3.12-dev as build-stage
WORKDIR /function
RUN pip3 install --target /python/ --no-cache --no-cache-dir debugpy
RUN rm -rf /python/bin
ADD . /function/
RUN rm -fr /function/.pip_cache
FROM fnproject/python:3.12
WORKDIR /function
COPY --from=build-stage /python /python
COPY --from=build-stage /function /function
RUN chmod -R o+r /function
ENV PYTHONPATH=/function:/python
ENTRYPOINT ["python3.12", "-m", "debugpy", "--listen", "0.0.0.0:5678", "--wait-for-client", "/python/bin/fdk", "/function/func.py", "handler"]
`
	pythonDockerfile = `FROM fnproject/python:3.12-dev as build-stage
WORKDIR /function
ADD . /function/
RUN rm -fr /function/.pip_cache
FROM fnproject/python:3.12
WORKDIR /function
COPY --from=build-stage /python /python
COPY --from=build-stage /function /function
RUN chmod -R o+r /function
ENV PYTHONPATH=/function:/python
ENTRYPOINT ["/python/bin/fdk", "/function/func.py", "handler"]
`
	goDebugDockerfile = `FROM fnproject/go:1.24-dev as build-stage
WORKDIR /function
ADD . /go/src/func/
RUN go build -gcflags="all=-N -l" -o func -v
RUN go install github.com/go-delve/delve/cmd/dlv@latest
FROM fnproject/go:1.24
WORKDIR /function
COPY --from=build-stage /go/src/func/func /function/
COPY --from=build-stage /go/bin/dlv /function
ENTRYPOINT ["/function/dlv", "--listen=:5678", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./func"]
`
	goDockerfile = `FROM fnproject/go:1.24-dev as build-stage
WORKDIR /function
ADD . /go/src/func/
RUN go build -o func -v
FROM fnproject/go:1.24
WORKDIR /function
COPY --from=build-stage /go/src/func/func /function/
ENTRYPOINT ["./func"]
`
	javaDebugDockerfile = `FROM fnproject/fn-java-fdk-build:jdk17-1.1.7 as build-stage
WORKDIR /function
ENV MAVEN_OPTS %s
ADD pom.xml /function/pom.xml
RUN ["mvn", "package", "dependency:copy-dependencies", "-DincludeScope=runtime", "-DskipTests=true", "-Dmdep.prependGroupId=true", "-DoutputDirectory=target", "--fail-never"]
ADD src /function/src
RUN ["mvn", "package"]
FROM fnproject/fn-java-fdk-build:jre17-1.1.7
WORKDIR /function
COPY --from=build-stage /function/target/*.jar /function/app/
ENTRYPOINT ["/usr/local/openjdk-17/bin/java", "-XX:-UsePerfData", "-XX:+UseSerialGC", "-Xshare:auto", "-Djava.awt.headless=true", "-Djava.library.path=/function/runtime/lib", "-cp", "/function/app/*:/function/runtime/*:/function/app:/function/app/resources", "-agentlib:jdwp=transport=dt_socket,server=y,suspend=y,address=*:5678", "com.fnproject.fn.runtime.EntryPoint"]
CMD ["com.example.fn.HelloFunction::handleRequest"]
`
	javaDockerfile = `FROM fnproject/fn-java-fdk-build:jdk17-1.1.7 as build-stage
WORKDIR /function
ENV MAVEN_OPTS %s
ADD pom.xml /function/pom.xml
RUN ["mvn", "package", "dependency:copy-dependencies", "-DincludeScope=runtime", "-DskipTests=true", "-Dmdep.prependGroupId=true", "-DoutputDirectory=target", "--fail-never"]
ADD src /function/src
RUN ["mvn", "package"]
FROM fnproject/fn-java-fdk-build:jre17-1.1.7
WORKDIR /function
COPY --from=build-stage /function/target/*.jar /function/app/
ENTRYPOINT ["/usr/local/openjdk-17/bin/java", "-XX:-UsePerfData", "-XX:+UseSerialGC", "-Xshare:auto", "-Djava.awt.headless=true", "-Djava.library.path=/function/runtime/lib", "-cp", "/function/app/*:/function/runtime/*:/function/app:/function/app/resources", "com.fnproject.fn.runtime.EntryPoint"]
CMD ["com.example.fn.HelloFunction::handleRequest"]
`

	nodeDockerfile = `FROM fnproject/node:24-dev as build-stage
WORKDIR /function
FROM fnproject/node:24
WORKDIR /function
ADD . /function/
RUN chmod -R o+r /function
ENTRYPOINT ["node", "func.js"]
`
	rubyDockerfile = `FROM fnproject/ruby:3.3-dev as build-stage
WORKDIR /function
FROM fnproject/ruby:3.3
WORKDIR /function
COPY --from=build-stage /usr/lib/ruby/gems/ /usr/lib/ruby/gems/
COPY . /function/
RUN chmod -R o+r /function
ENTRYPOINT ["ruby", "func.rb"]
`
	dotnetDockerfile = `FROM fnproject/dotnet:9.0-1.0.57-dev as build-stage
WORKDIR /function
COPY . .
RUN dotnet sln add src/Function/Function.csproj tests/Function.Tests/Function.Tests.csproj
RUN dotnet build -c Release
RUN dotnet test -c Release
RUN dotnet publish src/Function/Function.csproj -c Release -o out
FROM fnproject/dotnet:9.0-1.0.57
WORKDIR /function
COPY --from=build-stage /function/out/ /function/
ENTRYPOINT ["dotnet", "Function.dll"]
CMD ["Function:Greeter:greet"]
`

	javaFdkEntryPoint = "'[/usr/local/openjdk-17/bin/java -XX:-UsePerfData -XX:+UseSerialGC -Xshare:auto -Djava.awt.headless=true -Djava.library.path=/function/runtime/lib -cp /function/app/*:/function/runtime/*:/function/app:/function/app/resources com.fnproject.fn.runtime.EntryPoint]'"
)
