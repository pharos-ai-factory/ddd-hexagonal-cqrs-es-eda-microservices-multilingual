package openapi

import (
	"net/http"
	"strings"
	"testing"
)

func panicsWith(t *testing.T, text string, run func()) {
	t.Helper()
	defer func() {
		value := recover()
		if value == nil || !strings.Contains(value.(string), text) {
			t.Fatalf("expected panic containing %q, got %v", text, value)
		}
	}()
	run()
}

func TestRouteCoverageIsCheckedInBothDirections(t *testing.T) {
	for _, name := range specNames {
		t.Run(name, func(t *testing.T) {
			mux := NewMux(name, nil)
			panicsWith(t, "absent from OpenAPI", func() { mux.HandleFunc("GET /undocumented", func(http.ResponseWriter, *http.Request) {}) })
			panicsWith(t, "no HTTP handler", func() { mux.Handler() })
			for _, pattern := range mux.Patterns() {
				mux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
			}
			mux.Handler()
		})
	}
}

var specNames = []string{"storefront", "provider"}
