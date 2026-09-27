package support

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func FeatureOptions(t *testing.T, name, path, tags string) *godog.Options {
	t.Helper()
	format := "pretty"
	if directory := os.Getenv("BDD_REPORT_DIR"); directory != "" {
		if !filepath.IsAbs(directory) {
			t.Fatal("BDD_REPORT_DIR must be absolute")
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		format += ",cucumber:" + filepath.Join(directory, name+".json")
	}
	return &godog.Options{Format: format, Paths: []string{path}, Tags: tags, Strict: true, TestingT: t}
}
