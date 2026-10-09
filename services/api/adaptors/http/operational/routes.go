package operational

import (
	"context"
	"net/http"
	"time"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/security"
)

// Config contains the settings and injected ports for this HTTP boundary.
type Config struct {
	CLIKey      string
	Diagnostics func(context.Context) (any, error)
}

func (c Config) Mount(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}) {
	mux.HandleFunc("GET /diagnostics", c.diagnostics)
	mux.HandleFunc("GET /healthz", health)
}

func health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]string{"status": "ok", "role": "client-api"})
}

func (c Config) diagnostics(w http.ResponseWriter, r *http.Request) {
	if c.CLIKey == "" || !security.Equal(r.Header.Get("Authorization"), "Bearer "+c.CLIKey) {
		response.Unauthorised.Write(w)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if c.Diagnostics == nil {
		response.DiagnosticsUnavailable.Write(w)
		return
	}
	result, err := c.Diagnostics(ctx)
	if err != nil {
		response.DiagnosticsUnavailable.Write(w)
		return
	}
	response.JSON(w, 200, result)
}
