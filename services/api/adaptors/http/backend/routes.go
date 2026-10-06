package backend

import (
	"net/http"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/security"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

type Target struct{ URL, Key string }

type Config struct {
	Sessions a.Sessions
	CLIKey   string
	Origins  []string
	Backends map[string]Target
}

func (c Config) Mount(mux *contract.Mux) {
	for owner, backend := range c.Backends {
		handler := c.authoriseBackend(newBackendProxy(backend))
		registered := false
		for _, pattern := range mux.Patterns() {
			if mux.Owner(pattern) == owner {
				mux.Handle(pattern, handler)
				registered = true
			}
		}
		if !registered {
			panic("backend context absent from OpenAPI: " + owner)
		}
	}
}

func (c Config) authoriseBackend(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		cli := security.Equal(r.Header.Get("Authorization"), "Bearer "+c.CLIKey)
		if !cli {
			_, ok, err := c.Sessions.Authenticate(r.Context(), security.Token(r))
			if err != nil {
				response.SessionUnavailable.Write(w)
				return
			}
			if !ok {
				response.Unauthorised.Write(w)
				return
			}
			if r.Method == http.MethodPost && !security.Origin(r, c.Origins) {
				response.OriginDenied.Write(w)
				return
			}
		}
		r.Body = http.MaxBytesReader(w, r.Body, 65536)
		next.ServeHTTP(w, r)
	})
}
