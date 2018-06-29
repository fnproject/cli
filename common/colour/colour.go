package colour

import (
	"os"

	tty "github.com/mattn/go-isatty"
)

var useColors bool

var Colours map[string]func(string) string

func init() {
	useColors = tty.IsTerminal(os.Stdout.Fd()) || tty.IsCygwinTerminal(os.Stdout.Fd())
	Colours = map[string]func(string) string{
		"b": Bold,
		"i": Italic,
	}
}

func Bold(text string) string {
	if useColors {
		return "\x1b[1m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func Italic(text string) string {
	if useColors {
		return "\x1b[3m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func BoldRed(text string) string {
	if useColors {
		return "\x1b[31;1m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func BrightRed(text string) string {
	if useColors {
		return "\x1b[91;21m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func UnderlineBrightRed(text string) string {
	if useColors {
		return "\x1b[91;4m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func BrightCyan(text string) string {
	if useColors {
		return "\x1b[96;21m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func Cyan(text string) string {
	if useColors {
		return "\x1b[36;21m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func BoldCyan(text string) string {
	if useColors {
		return "\x1b[36;1m" + text + "\x1b[0m"
	} else {
		return text
	}
}

func Yellow(text string) string {
	if useColors {
		return "\x1b[33;21m" + text + "\x1b[0m"
	} else {
		return text
	}
}
