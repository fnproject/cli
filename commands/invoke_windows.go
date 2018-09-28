// +build windows

package commands

import (
	"io"
	"os"
)

func stdin() io.Reader {
	if isTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	return os.Stdin
}
