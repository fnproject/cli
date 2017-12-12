package langs

type LambdaNodeHelper struct {
	BaseHelper
}

func (h *LambdaNodeHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *LambdaNodeHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *LambdaNodeHelper) LangStrings() []string {
	return []string{"lambda-nodejs4.3", "lambda-node-4"}
}
func (lh *LambdaNodeHelper) Extensions() []string {
	return []string{".py"}
}

func (lh *LambdaNodeHelper) BuildFromImage() (string, error) {
	return "fnproject/lambda:node-4", nil
}

func (lh *LambdaNodeHelper) RunFromImage() (string, error) {
	return "fnproject/lambda:node-4", nil
}

func (lh *LambdaNodeHelper) IsMultiStage() bool {
	return false
}

func (lh *LambdaNodeHelper) Cmd() (string, error) {
	return "func.handler", nil
}

func (h *LambdaNodeHelper) DockerfileBuildCmds() []string {
	r := []string{}
	if exists("package.json") {
		r = append(r,
			"ADD package.json /function/",
			"RUN npm install",
		)
	}
	// single stage build for this one, so add files
	r = append(r, "ADD . /function/")
	return r
}
