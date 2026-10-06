package realtime

import (
	"net/http"
	"time"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/security"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

type Config struct {
	Sessions    a.Sessions
	ProxySecret string
	Origins     []string
}

func (c Config) Mount(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}) {
	mux.HandleFunc("POST /api/realtime/connect", c.connect)
	mux.HandleFunc("POST /api/realtime/refresh", c.refresh)
}

func (c Config) connect(w http.ResponseWriter, r *http.Request) {
	if !security.Equal(r.Header.Get("X-Cafe-Realtime-Proxy"), c.ProxySecret) {
		response.ProxyDenied.Write(w)
		return
	}
	// Calculate before reading authority: every response authorised before
	// session deletion must expire within the worker's revocation fence.
	expires := time.Now().Add(a.ConnectLifetime).Unix()
	principal, ok, err := c.Sessions.Authenticate(r.Context(), security.Token(r))
	if err != nil || !ok || !security.Origin(r, c.Origins) {
		response.JSON(w, 200, map[string]any{"disconnect": map[string]any{"code": 4501, "reason": "unauthorised"}})
		return
	}
	subs := map[string]any{}
	for _, channel := range a.AuthorisedChannels(principal) {
		subs[channel] = map[string]any{}
	}
	response.JSON(w, 200, map[string]any{"result": map[string]any{
		"user": principal.Subject, "subs": subs, "expire_at": expires}})
}

func (c Config) refresh(w http.ResponseWriter, r *http.Request) {
	if !security.Equal(r.Header.Get("X-Cafe-Realtime-Proxy"), c.ProxySecret) {
		response.ProxyDenied.Write(w)
		return
	}
	// Centrifugo forwards the original connection's Cookie and Origin.
	// Refresh cannot create a connection or revive one already disconnected.
	expires := time.Now().Add(time.Minute).Unix()
	_, ok, err := c.Sessions.Authenticate(r.Context(), security.Token(r))
	if err != nil || !ok || !security.Origin(r, c.Origins) {
		response.JSON(w, 200, map[string]any{"result": map[string]bool{"expired": true}})
		return
	}
	response.JSON(w, 200, map[string]any{"result": map[string]int64{"expire_at": expires}})
}
