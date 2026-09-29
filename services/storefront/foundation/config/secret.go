// Package config resolves runtime configuration without exposing secret values.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Secret accepts either NAME or NAME_FILE. File contents may end in one newline.
// Resolution is deliberately strict: an invalid file never falls back to NAME.
func Secret(name string) (string, error) {
	value, direct := os.LookupEnv(name)
	path, file := os.LookupEnv(name + "_FILE")
	if direct && file {
		return "", fmt.Errorf("%s and %s_FILE are mutually exclusive", name, name)
	}
	if file {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("%s_FILE cannot be read", name)
		}
		value = strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required and must not be empty", name)
	}
	return value, nil
}
