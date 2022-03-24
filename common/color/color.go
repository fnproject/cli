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
		"trim":               strings.TrimLeft,
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
