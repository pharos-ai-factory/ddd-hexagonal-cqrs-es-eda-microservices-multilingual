package backend

import (
	"context"
	"errors"
	mapping "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/backend/generated"
	"net/http"
	"strings"
	"time"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/security"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

type Requests interface {
	Call(context.Context, pb.Request) (pb.Reply, error)
}

type Config struct {
	Sessions a.Sessions
	CLIKey   string
	Origins  []string
	Owners   []string
	Requests Requests
}

func (c Config) Mount(mux *contract.Mux) {
	for _, owner := range c.Owners {
		registered := false
		for _, pattern := range mux.Patterns() {
			if mux.Owner(pattern) == owner {
				operation := mux.Operation(pattern)
				if !mapping.Supported(operation.OperationID) {
					panic("OpenAPI operation has no typed boundary mapping: " + operation.OperationID)
				}
				_, path, _ := strings.Cut(pattern, " ")
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					request, err := translate(r, owner, path, operation)
					if err != nil {
						var rejected requestError
						if errors.As(err, &rejected) {
							response.JSON(w, rejected.status, response.ErrorResponse{Code: rejected.code})
						} else {
							response.TemporarilyUnavailable.Write(w)
						}
						return
					}
					ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
					defer cancel()
					if c.Requests == nil {
						response.TemporarilyUnavailable.Write(w)
						return
					}
					reply, err := c.Requests.Call(ctx, request)
					if err != nil {
						response.TemporarilyUnavailable.Write(w)
						return
					}
					status, body, err := translateReply(request, reply, path)
					if err != nil {
						response.TemporarilyUnavailable.Write(w)
						return
					}
					response.JSON(w, status, body)
				})
				mux.Handle(pattern, c.authoriseBackend(handler))
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
