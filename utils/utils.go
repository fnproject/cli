package utils

import (
	"bytes"
	"strings"
)

func StringToSlice(in string) string {
	epvals := strings.Fields(in)
	var buffer bytes.Buffer
	for i, s := range epvals {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString("\"")
		buffer.WriteString(s)
		buffer.WriteString("\"")
	}
	return buffer.String()
}
