package http

import (
	"context"
	"net/http"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/backend"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/operational"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/realtime"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/session"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

type Config struct {
	Diagnostics                           func(context.Context) (any, error)
	Sessions                              a.Sessions
	OperatorPassword, CLIKey, ProxySecret string
	Origins                               []string
	Owners                                []string
	Requests                              backend.Requests
}

func (c Config) Handler() http.Handler {
	owners := map[string]bool{}
	for _, owner := range c.Owners {
		owners[owner] = true
	}
	mux := contract.NewMux("api", owners)
	operational.Config{CLIKey: c.CLIKey, Diagnostics: c.Diagnostics}.Mount(mux)
	session.Config{Sessions: c.Sessions, OperatorPassword: c.OperatorPassword, Origins: c.Origins}.Mount(mux)
	realtime.Config{Sessions: c.Sessions, ProxySecret: c.ProxySecret, Origins: c.Origins}.Mount(mux)
	backend.Config{Sessions: c.Sessions, CLIKey: c.CLIKey, Origins: c.Origins, Owners: c.Owners, Requests: c.Requests}.Mount(mux)
	return mux.Handler()
}
