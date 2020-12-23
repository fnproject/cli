package context

import (
	"github.com/fnxproject/cli/config"
)

// Info holds the information found in the context YAML file
type Info struct {
	Current bool   `json:"current"`
	Name    string `json:"name"`
	*config.ContextFile
}

// NewInfo creates an instance of the contextInfo
// by parsing the provided context YAML file. This is used
// for outputting the context information
func NewInfo(name string, isCurrent bool, contextFile *config.ContextFile) *Info {
	return &Info{
		Name:        name,
		Current:     isCurrent,
		ContextFile: contextFile,
	}
}
