// Package openapi binds HTTP routes to the embedded, authoritative wire contract.
package openapi

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed generated/*.openapi.json
var specifications embed.FS

func Document(name string) *openapi3.T {
	data, err := specifications.ReadFile("generated/" + name + ".openapi.json")
	if err != nil {
		panic(err)
	}
	document, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		panic(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		panic(err)
	}
	return document
}

// Mux checks both directions: every handler is declared, and every enabled
// operation has a handler. Nil owners includes every operation in the document.
type Mux struct {
	mux      *http.ServeMux
	document *openapi3.T
	declared map[string]bool
	bound    map[string]bool
}

func NewMux(name string, owners map[string]bool) *Mux {
	m := &Mux{mux: http.NewServeMux(), document: Document(name), declared: map[string]bool{}, bound: map[string]bool{}}
	for path, item := range m.document.Paths.Map() {
		owner, _ := item.Extensions["x-owner"].(string)
		if owners != nil && owner != "" && !owners[owner] {
			continue
		}
		for method := range item.Operations() {
			m.declared[method+" "+path] = true
		}
	}
	return m
}

func (m *Mux) Patterns() []string {
	patterns := make([]string, 0, len(m.declared))
	for pattern := range m.declared {
		patterns = append(patterns, pattern)
	}
	sort.Strings(patterns)
	return patterns
}

func (m *Mux) Owner(pattern string) string {
	_, path, _ := strings.Cut(pattern, " ")
	owner, _ := m.document.Paths.Value(path).Extensions["x-owner"].(string)
	return owner
}

func (m *Mux) Operation(pattern string) *openapi3.Operation {
	method, path, _ := strings.Cut(pattern, " ")
	return m.document.Paths.Value(path).GetOperation(method)
}

func (m *Mux) Handle(pattern string, handler http.Handler) {
	if !m.declared[pattern] {
		panic(fmt.Sprintf("HTTP route absent from OpenAPI: %s", pattern))
	}
	m.mux.Handle(pattern, handler)
	m.bound[pattern] = true
}
func (m *Mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handler))
}

func (m *Mux) Handler() http.Handler {
	for _, pattern := range m.Patterns() {
		if !m.bound[pattern] {
			panic(fmt.Sprintf("OpenAPI operation has no HTTP handler: %s", pattern))
		}
	}
	return m.mux
}
