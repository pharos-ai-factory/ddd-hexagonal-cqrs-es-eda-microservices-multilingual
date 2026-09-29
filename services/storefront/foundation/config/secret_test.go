package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretFiles(t *testing.T) {
	const name = "CAFE_TEST_CONFIG_SECRET"
	t.Setenv(name+"_FILE", filepath.Join(t.TempDir(), "secret"))
	path := os.Getenv(name + "_FILE")
	if err := os.WriteFile(path, []byte("private-value\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if value, err := Secret(name); err != nil || value != "private-value" {
		t.Fatal("file secret was not resolved")
	}
	t.Run("ambiguous", func(t *testing.T) {
		t.Setenv(name, "")
		if _, err := Secret(name); err == nil {
			t.Fatal("accepted ambiguous sources")
		}
	})
	for _, content := range []string{"", " \n"} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Secret(name); err == nil {
			t.Fatal("accepted empty secret")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Secret(name); err == nil || strings.Contains(err.Error(), path) {
		t.Fatal("missing file must fail without revealing path")
	}
	t.Setenv(name+"_FILE", filepath.Dir(path))
	if _, err := Secret(name); err == nil {
		t.Fatal("accepted directory")
	}
}

func TestDirectSecret(t *testing.T) {
	t.Setenv("CAFE_TEST_DIRECT_SECRET", " value ")
	value, err := Secret("CAFE_TEST_DIRECT_SECRET")
	if err != nil || value != " value " {
		t.Fatal("direct secret changed")
	}
}
