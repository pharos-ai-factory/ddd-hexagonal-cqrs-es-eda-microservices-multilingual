package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
)

func newBackendProxy(backend Target) *httputil.ReverseProxy {
	target, err := url.Parse(backend.URL)
	if err != nil || target.Host == "" || target.Scheme != "http" {
		panic("invalid internal service URL")
	}
	return &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(target)
			request.Out.URL.Path = strings.TrimPrefix(request.In.URL.Path, "/api")
			request.Out.Header = make(http.Header)
			for _, header := range []string{"Content-Type", "Idempotency-Key", "If-Match", "X-Correlation-ID"} {
				if value := request.In.Header.Get(header); value != "" {
					request.Out.Header.Set(header, value)
				}
			}
			request.Out.Header.Set("Authorization", "Bearer "+backend.Key)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			response.TemporarilyUnavailable.Write(w)
		},
	}
}
