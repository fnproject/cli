package common

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func LoadCLIJSONInput(spec string, out interface{}) error {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "file://") {
		data, err := os.ReadFile(strings.TrimPrefix(trimmed, "file://"))
		if err != nil {
			return err
		}
		trimmed = string(data)
	}
	if err := json.Unmarshal([]byte(trimmed), out); err != nil {
		return fmt.Errorf("invalid --from-json payload: %w", err)
	}
	return nil
}
