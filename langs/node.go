package langs

type NodeLangHelper struct {
	BaseHelper
}

func (lh *NodeLangHelper) BuildFromImage() (string, error) {
	return "fnproject/node:dev", nil
}
func (lh *NodeLangHelper) RunFromImage() (string, error) {
	return "fnproject/node", nil
}

func (lh *NodeLangHelper) Entrypoint() (string, error) {
	return "node func.js", nil
}

func (h *NodeLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	if exists("package.json") {
		r = append(r,
			"ADD package.json /function/",
			"RUN npm install",
		)
	}
	return r
}

func (h *NodeLangHelper) DockerfileCopyCmds() []string {
	r := []string{"ADD . /function/"}
	if exists("package.json") {
		r = append(r, "COPY --from=build-stage /function/node_modules/ /function/node_modules/")
	}
	return r
}
