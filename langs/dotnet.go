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

package langs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DotnetLangHelper struct {
	BaseHelper
	Version string
}

func (h *DotnetLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}

func (h *DotnetLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *DotnetLangHelper) CustomMemory() uint64 {
	return 0
}

func (lh *DotnetLangHelper) LangStrings() []string {
	return []string{"dotnet", fmt.Sprintf("dotnet%s", lh.Version)}
}

func (lh *DotnetLangHelper) Extensions() []string {
	return []string{".cs"}
}

func (lh *DotnetLangHelper) BuildFromImage() (string, error) {
	fdkVersion, _ := lh.GetLatestFDKVersion()
	return fmt.Sprintf("fnproject/dotnet:%s-%s-dev", lh.Version, fdkVersion), nil
}

func (lh *DotnetLangHelper) RunFromImage() (string, error) {
	fdkVersion, _ := lh.GetLatestFDKVersion()
	return fmt.Sprintf("fnproject/dotnet:%s-%s", lh.Version, fdkVersion), nil
}

func (h *DotnetLangHelper) DockerfileBuildCmds() []string {
	r := []string{"COPY . ."}
	r = append(r, "RUN dotnet sln add src/Function/Function.csproj tests/Function.Tests/Function.Tests.csproj")
	r = append(r, "RUN dotnet build -c Release")
	r = append(r, "RUN dotnet test -c Release")
	r = append(r, "RUN dotnet publish src/Function/Function.csproj -c Release -o out")
	return r
}

func (h *DotnetLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function/out/ /function/",
	}
}

func (h *DotnetLangHelper) Entrypoint() (string, error) {
	return "dotnet Function.dll", nil
}

func (lh *DotnetLangHelper) HasBoilerplate() bool { return true }

func (lh *DotnetLangHelper) GenerateBoilerplate(path string) error {
	slnFile := filepath.Join(path, "Function.sln")
	if exists(slnFile) {
		return errors.New("Solution file already exists, canceling init")
	}
	if err := ioutil.WriteFile(slnFile, []byte(slnFileBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	csprojFile := "Function.csproj"
	fdkVersion, _ := lh.GetLatestFDKVersion()

	if err := mkdirAndWriteFile(path, "src/Function", csprojFile, fmt.Sprintf(srcCsprojBoilerplate, fdkVersion)); err != nil {
		return err
	}

	codeFile := filepath.Join(path, "src/Function", "Program.cs")
	if err := ioutil.WriteFile(codeFile, []byte(helloDotnetSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	testCsprojFile := "Function.Tests.csproj"
	if err := mkdirAndWriteFile(path, "tests/Function.Tests", testCsprojFile, testsCsprojBoilerplate); err != nil {
		return err
	}

	testFile := filepath.Join(path, "tests/Function.Tests", "ProgramTest.cs")
	if err := ioutil.WriteFile(testFile, []byte(helloDotnetTestBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

func (h *DotnetLangHelper) Cmd() (string, error) {
	return "Function:Greeter:greet", nil
}

func (h *DotnetLangHelper) GetLatestFDKVersion() (string, error) {
	return getLatestFDKVersionFromGithub("fnproject/fdk-dotnet")
}

const (
	helloDotnetSrcBoilerplate = `using Fnproject.Fn.Fdk;

using System.Runtime.CompilerServices;
[assembly:InternalsVisibleTo("Function.Tests")]
namespace Function {
	class Greeter {
		public string greet(string input) {
			return string.Format("Hello {0}!",
				input.Length == 0 ? "World" : input.Trim());
		}

		static void Main(string[] args) { Fdk.Handle(args[0]); }
	}
}
`

	helloDotnetTestBoilerplate = `using Function;
using NUnit.Framework;

namespace Function.Tests {
	public class GreeterTest {
		[Test]
		public void TestGreetValid() {
			Greeter greeter = new Greeter();
			string response = greeter.greet("Dotnet");
			Assert.AreEqual("Hello Dotnet!", response);
		}

		[Test]
		public void TestGreetEmpty() {
			Greeter greeter = new Greeter();
			string response = greeter.greet("");
			Assert.AreEqual("Hello World!", response);
		}
	}
}
`

	slnFileBoilerplate = `Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio 15
VisualStudioVersion = 15.0.26124.0
MinimumVisualStudioVersion = 15.0.26124.0
Global
  GlobalSection(SolutionConfigurationPlatforms) = preSolution
 	 Debug|Any CPU = Debug|Any CPU
 	 Debug|x64 = Debug|x64
 	 Debug|x86 = Debug|x86
 	 Release|Any CPU = Release|Any CPU
 	 Release|x64 = Release|x64
 	 Release|x86 = Release|x86
  EndGlobalSection
  GlobalSection(SolutionProperties) = preSolution
 	 HideSolutionNode = FALSE
  EndGlobalSection
EndGlobal
`

	srcCsprojBoilerplate = `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
  <OutputType>Exe</OutputType>
  <TargetFramework>netcoreapp3.1</TargetFramework>
  </PropertyGroup>

  <ItemGroup>
  <PackageReference Include="Fnproject.Fn.Fdk" Version="%s" />
  </ItemGroup>
</Project>
`

	testsCsprojBoilerplate = `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
  <TargetFramework>netcoreapp3.1</TargetFramework>

  <IsPackable>false</IsPackable>
  </PropertyGroup>

  <ItemGroup>
  <PackageReference Include="NUnit" Version="3.12.0" />
  <PackageReference Include="NUnit3TestAdapter" Version="3.16.1" />
  <PackageReference Include="Microsoft.NET.Test.Sdk" Version="16.5.0"/>
 	 <ProjectReference Include="..\..\src\Function\Function.csproj" />
  </ItemGroup>

</Project>
`
)

func (h *DotnetLangHelper) FixImagesOnInit() bool {
	return true
}
