package run

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fnproject/cli/common"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	DefaultFormat    = "default"
	HttpFormat       = "http"
	JSONFormat       = "json"
	CloudEventFormat = "cloudevent"
	TestApp          = "myapp"
	TestRoute        = "/hello"
	LocalTestURL     = "http://localhost:8080"
)

func RunCommand() cli.Command {
	r := runCmd{}

	return cli.Command{
		Name:     "run",
		Usage:    "run a function locally",
		Aliases:  []string{"r"},
		Category: "DEVELOPMENT COMMANDS",
		Flags:    append(GetRunFlags(), []cli.Flag{}...),
		Action:   r.run,
	}
}

type runCmd struct{}

var RunFlags = []cli.Flag{
	cli.StringSliceFlag{
		Name:  "env, e",
		Usage: "select environment variables to be sent to function",
	},
	cli.StringSliceFlag{
		Name:  "link",
		Usage: "select container links for the function",
	},
	cli.StringFlag{
		Name:  "method",
		Usage: "http method for function",
	},
	cli.StringFlag{
		Name:  "format",
		Usage: "format to use. `default` and `http` (hot) formats currently supported.",
	},
	cli.IntFlag{
		Name:  "runs",
		Usage: "for hot functions only, will call the function `runs` times in a row.",
	},
	cli.Uint64Flag{
		Name:  "memory",
		Usage: "RAM to allocate for function, Units: MB",
	},
	cli.BoolFlag{
		Name:  "no-cache",
		Usage: "Don't use Docker cache for the build",
	},
	cli.StringFlag{
		Name:  "content-type",
		Usage: "The payload Content-Type for the function invocation.",
	},
	cli.StringSliceFlag{
		Name:  "build-arg",
		Usage: "set build time variables",
	},
}

func GetRunFlags() []cli.Flag {
	return RunFlags
}

// PreRun parses func.yaml, checks expected env vars and builds the function image.
func PreRun(c *cli.Context) (string, *common.FuncFile, []string, error) {
	wd := common.GetWd()
	// if image name is passed in, it will run that image
	path := c.Args().First() // TODO: should we ditch this?
	var err error
	var ff *common.FuncFile
	var fpath string

	if path != "" {
		fmt.Printf("Running function at: /%s\n", path)
		dir := filepath.Join(wd, path)
		err := os.Chdir(dir)
		if err != nil {
			return "", nil, nil, err
		}
		defer os.Chdir(wd) // todo: wrap this so we can log the error if changing back fails
		wd = dir
	}

	fpath, ff, err = common.FindAndParseFuncfile(wd)
	if err != nil {
		return fpath, nil, nil, err
	}

	// check for valid input
	envVars := c.StringSlice("env")
	// Check expected env vars defined in func file
	for _, expected := range ff.Expects.Config {
		n := expected.Name
		e := getEnvValue(n, envVars)
		if e != "" {
			continue
		}
		e = os.Getenv(n)
		if e != "" {
			envVars = append(envVars, kvEq(n, e))
			continue
		}
		if expected.Required {
			return "", ff, envVars, fmt.Errorf("required env var %s not found, please set either set it in your environment or pass in `-e %s=X` flag.", n, n)
		}
		fmt.Fprintf(os.Stderr, "info: optional env var %s not found.\n", n)
	}
	// get name from directory if it's not defined
	if ff.Name == "" {
		ff.Name = filepath.Base(filepath.Dir(fpath)) // todo: should probably make a copy of ff before changing it
	}

	buildArgs := c.StringSlice("build-arg")
	_, err = common.BuildFunc(c, fpath, ff, buildArgs, c.Bool("no-cache"))
	if err != nil {
		return fpath, nil, nil, err
	}
	return fpath, ff, envVars, nil
}

func getEnvValue(n string, envVars []string) string {
	for _, e := range envVars {
		// assuming has equals for now
		split := strings.Split(e, "=")
		if split[0] == n {
			return split[1]
		}
	}
	return ""
}

func (r *runCmd) run(c *cli.Context) error {
	_, ff, envVars, err := PreRun(c)
	if err != nil {
		return err
	}
	// means no memory specified through CLI args
	// memory from func.yaml applied
	if c.Uint64("memory") != 0 {
		ff.Memory = c.Uint64("memory")
	}
	return RunFF(ff, Stdin(), os.Stdout, os.Stderr, c.String("method"), envVars, c.StringSlice("link"), c.String("format"), c.Int("runs"), c.String("content-type"))
}

// TODO: share all this stuff with the Docker driver in server or better yet, actually use the Docker driver
func RunFF(ff *common.FuncFile, stdin io.Reader, stdout, stderr io.Writer, method string, envVars []string, links []string, format string, runs int, contentType string) error {
	sh := []string{"docker", "run", "--rm", "-i", fmt.Sprintf("--memory=%dm", ff.Memory)}

	var env []string    // env for the shelled out docker run command
	var runEnv []string // env to pass into the container via -e's
	callID := "12345678901234567890123456"
	if contentType == "" {
		if ff.ContentType != "" {
			contentType = ff.ContentType
		} else {
			contentType = "text/plain"
		}
	}

	fmt.Println("Content type: ", contentType)
	to := int32(30)
	if ff.Timeout != nil {
		to = *ff.Timeout
	}
	deadline := time.Now().Add(time.Duration(to) * time.Second)
	deadlineS := deadline.Format(time.RFC3339)

	if method == "" {
		if stdin == nil {
			method = "GET"
		} else {
			method = "POST"
		}
	}
	if format == "" {
		if ff.Format != "" {
			format = ff.Format
		} else {
			format = DefaultFormat
		}
	}

	// Add expected env vars that service will add
	// Full set here: https://github.com/fnproject/fn/pull/660#issuecomment-356157279
	runEnv = append(runEnv, kvEq("FN_TYPE", "sync"))
	runEnv = append(runEnv, kvEq("FN_FORMAT", format))
	runEnv = append(runEnv, kvEq("FN_PATH", TestRoute))
	runEnv = append(runEnv, kvEq("FN_MEMORY", fmt.Sprintf("%d", ff.Memory)))
	runEnv = append(runEnv, kvEq("FN_APP_NAME", TestApp))

	// add user defined envs
	runEnv = append(runEnv, envVars...)

	var requestURL string
	if requestURL = viper.GetString("api-url"); requestURL == "" {
		requestURL = LocalTestURL
	}

	requestURL = requestURL + "/" + TestApp + TestRoute

	if format == DefaultFormat {
		runEnv = append(runEnv, kvEq("FN_REQUEST_URL", requestURL))
		runEnv = append(runEnv, kvEq("FN_CALL_ID", callID))
		runEnv = append(runEnv, kvEq("FN_METHOD", method))
		runEnv = append(runEnv, kvEq("FN_DEADLINE", deadlineS))
	}

	if runs <= 0 {
		runs = 1
	}

	// NOTE: 'run' does 'sync'

	handler := handle(format)

	// TODO: we really should handle multiple runs better, clearly delimiting them,
	// as well as at least an option for outputting the entire http / json blob.

	stdin, stdout, err := handler(runConfig{
		runs:        runs,
		method:      method,
		url:         requestURL,
		contentType: contentType,
		callID:      callID,
		deadline:    deadlineS,
		stdin:       stdin,
		stdout:      stdout,
	})

	if err != nil {
		return err
	}

	for _, l := range links {
		sh = append(sh, "--link", l)
	}

	dockerenv := []string{"DOCKER_TLS_VERIFY", "DOCKER_HOST", "DOCKER_CERT_PATH", "DOCKER_MACHINE_NAME"}
	for _, e := range dockerenv {
		env = append(env, fmt.Sprint(e, "=", os.Getenv(e)))
	}

	for _, e := range runEnv {
		sh = append(sh, "-e", e)
	}

	sh = append(sh, ff.ImageName())
	cmd := exec.Command(sh[0], sh[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// cmd.Env = env
	return cmd.Run()
}

type runConfig struct {
	runs        int
	method      string
	url         string
	contentType string
	callID      string
	deadline    string
	stdin       io.Reader
	stdout      io.Writer
}

type handlerFunc func(runConfig) (stdin io.Reader, stdout io.Writer, err error)

func handle(format string) handlerFunc {
	switch format {
	case HttpFormat:
		return handleHTTP
	case JSONFormat:
		return handleJSON
	case CloudEventFormat:
		return handleCloudEvent
	default:
		return handleDefault
	}
}

func handleDefault(conf runConfig) (io.Reader, io.Writer, error) {
	return conf.stdin, conf.stdout, nil
}

func handleHTTP(conf runConfig) (io.Reader, io.Writer, error) {
	var b bytes.Buffer
	for i := 0; i < conf.runs; i++ {
		// making new request each time since Write closes the body
		// TODO: this isn't do the headers like recent changes on the server side
		// let's swap out stdin for http formatted message
		req, err := http.NewRequest(conf.method, conf.url, conf.stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating http request: %v", err)
		}
		req.Header.Set("Content-Type", conf.contentType) // TODO this isn't a thing (see add headers)

		req.Header.Set("FN_REQUEST_URL", conf.url)
		req.Header.Set("FN_CALL_ID", conf.callID)
		req.Header.Set("FN_METHOD", conf.method)
		req.Header.Set("FN_DEADLINE", conf.deadline)
		err = req.Write(&b)
		if err != nil {
			return nil, nil, fmt.Errorf("error writing to byte buffer: %v", err)
		}
	}

	body := b.String()
	stdin := strings.NewReader(body)
	stdout := stdoutHTTP(conf.stdout)
	return stdin, stdout, nil
}

// TODO run should not be setup like this and this should get shot. hot patching...
func stdoutHTTP(stdout io.Writer) io.Writer {
	pr, pw := io.Pipe()

	go func() {
		buf := bufio.NewReader(pr)
		for {
			resp, err := http.ReadResponse(buf, nil)
			if err != nil {
				fmt.Println("error reading http", err)
				return
			}
			io.Copy(stdout, resp.Body)
			resp.Body.Close()
		}
	}()
	return pw
}

func handleJSON(conf runConfig) (io.Reader, io.Writer, error) {
	var b strings.Builder
	for i := 0; i < conf.runs; i++ {
		body, err := createJSONInput(conf.callID, conf.contentType, conf.deadline, conf.method, conf.url, conf.stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating input: %v", err)
		}
		b.WriteString(body)
		b.Write([]byte("\n"))
	}
	stdin := strings.NewReader(b.String())
	stdout := stdoutJSON(conf.stdout)
	return stdin, stdout, nil
}

func handleCloudEvent(conf runConfig) (io.Reader, io.Writer, error) {
	var b strings.Builder
	for i := 0; i < conf.runs; i++ {
		body, err := createCloudEventInput(conf.callID, conf.contentType, conf.deadline, conf.method, conf.url, conf.stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating input: %v", err)
		}
		b.WriteString(body)
		b.Write([]byte("\n"))
	}
	stdin := strings.NewReader(b.String())
	stdout := stdoutCloudEvent(conf.stdout)
	return stdin, stdout, nil
}

func extractEnvVar(e string) ([]string, string) {
	kv := strings.Split(e, "=")
	name := toEnvName("HEADER", kv[0])
	sh := []string{"-e", name}
	var v string
	if len(kv) > 1 {
		v = kv[1]
	} else {
		v = os.Getenv(kv[0])
	}
	return sh, kvEq(name, v)
}

func kvEq(k, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

// From server.toEnvName()
func toEnvName(envtype, name string) string {
	name = strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return fmt.Sprintf("%s_%s", envtype, name)
}
