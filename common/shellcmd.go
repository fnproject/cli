package common

import (
	"io"
	"os/exec"
)

type ShellCommand interface {
	Start() error
	Run() error
	Wait() error
	Kill() error
	SetStdOut(io.Writer)
	SetStdErr(io.Writer)
}

type shellCommandImpl struct {
	*exec.Cmd
}

func (s *shellCommandImpl) SetStdOut(w io.Writer) {
	s.Cmd.Stdout = w
}
func (s *shellCommandImpl) SetStdErr(w io.Writer) {
	s.Cmd.Stderr = w
}
func (s *shellCommandImpl) Kill() error {
	return s.Cmd.Process.Kill()
}

func newExecShellCommander(name string, arg ...string) ShellCommand {
	execCmd := exec.Command(name, arg...)
	return &shellCommandImpl{Cmd: execCmd}
}

var ShellCommander = newExecShellCommander
