package color

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

var useColors bool

var Colors map[string]interface{}

func init() {
	useColors = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	Colors = map[string]interface{}{
		"bold":               Bold,
		"italic":             Italic,
		"join":               strings.Join,
		"cyan":               Cyan,
		"brightcyan":         BrightCyan,
		"boldcyan":           BoldCyan,
		"yellow":             Yellow,
		"red":                Red,
		"brightred":          BrightRed,
		"boldred":            BoldRed,
		"underlinebrightred": UnderlineBrightRed,
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

func Red(text string) string {
	if useColors {
		return "\x1b[31;21m" + text + "\x1b[0m"
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
