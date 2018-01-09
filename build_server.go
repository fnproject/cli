// This command builds a custom fn server with extensions compiled into it.
//
// NOTES:
// * We could just add extensions in the imports, but then there's no way to order them or potentially add extra config (although config should almost always be via env vars)

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

func buildServer() cli.Command {
	cmd := buildServerCmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "build-server",
		Usage:  "build custom fn server",
		Flags:  flags,
		Action: cmd.buildServer,
	}
}

type buildServerCmd struct {
	verbose bool
	noCache bool
}

func (b *buildServerCmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &b.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
		cli.StringFlag{
			Name:  "tag,t",
			Usage: "image name and optional tag",
		},
		cli.StringFlag{
			Name:  "fn-branch",
			Value: "master",
			Usage: "branch in github.com/fnproject/fn to build off",
		},
	}
}

// steps:
// • Yaml file with extensions listed
// • NO‎TE: All extensions should use env vars for config
// • ‎Generating main.go with extensions
// * Generate a Dockerfile that gets all the extensions (using dep)
// • ‎then generate a main.go with extensions
// • ‎compile, throw in another container like main dockerfile
func (b *buildServerCmd) buildServer(c *cli.Context) error {

	if c.String("tag") == "" {
		return errors.New("docker tag required")
	}

	fpath := "ext.yaml"
	bb, err := ioutil.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("could not open %s for parsing. Error: %v", fpath, err)
	}
	ef := &extFile{}
	err = yaml.Unmarshal(bb, ef)
	if err != nil {
		return err
	}

	err = os.MkdirAll("tmp", 0777)
	if err != nil {
		return err
	}
	err = os.Chdir("tmp")
	if err != nil {
		return err
	}
	err = generateMain(ef)
	if err != nil {
		return err
	}
	err = generateGopkg(c.String("fn-branch"))
	if err != nil {
		return err
	}
	err = generateDockerfile()
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	err = runBuild(c, dir, c.String("tag"), "Dockerfile", b.noCache)
	if err != nil {
		return err
	}
	fmt.Printf("Custom Fn server built successfully.\n")
	return nil
}

func generateMain(ef *extFile) error {
	tmpl, err := template.New("main").Parse(mainTmpl)
	if err != nil {
		return err
	}
	f, err := os.Create("main.go")
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, ef)
	if err != nil {
		return err
	}
	return nil
}

func generateGopkg(branch string) error {
	tmpl, err := template.New("gopkg").Parse(gopkgTmpl)
	if err != nil {
		return err
	}
	f, err := os.Create("Gopkg.toml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, &constraint{Branch: branch})
	if err != nil {
		return err
	}
	return nil
}

func generateDockerfile() error {
	if err := ioutil.WriteFile("Dockerfile", []byte(dockerFileTmpl), os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

type extFile struct {
	Extensions []*extInfo `yaml:"extensions"`
}

type extInfo struct {
	Name string `yaml:"name"`
	// will have version and other things down the road
}

var mainTmpl = `package main

import (
	"context"

	"github.com/fnproject/fn/api/server"
	
	{{- range .Extensions }}
		_ "{{ .Name }}"
	{{- end}}
)

func main() {
	ctx := context.Background()
	funcServer := server.NewFromEnv(ctx)
	{{- range .Extensions }}
		funcServer.AddExtensionByName("{{ .Name }}")
	{{- end}}
	funcServer.Start(ctx)
}
`

type constraint struct {
	Branch string
}

var gopkgTmpl = `
[[constraint]]
  name = "github.com/fnproject/fn"
  branch = "{{.Branch}}"
`

var dockerFileTmpl = `# build stage
FROM golang:1.9-alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
RUN go get -u github.com/golang/dep/cmd/dep
ENV D=/go/src/github.com/x/y
ADD main.go Gopkg.toml $D/
RUN cd $D && dep ensure
# RUN cd $D && go get
RUN cd $D && go build -o fnserver && cp fnserver /tmp/

# final stage
FROM fnproject/dind
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build-env /tmp/fnserver /app/fnserver
CMD ["./fnserver"]
`
