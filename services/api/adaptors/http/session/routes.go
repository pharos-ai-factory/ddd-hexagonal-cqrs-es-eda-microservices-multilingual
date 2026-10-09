package session

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/security"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

// Config contains the settings and injected ports for this HTTP boundary.
type Config struct {
	Sessions         a.Sessions
	OperatorPassword string
	Origins          []string
}

func (c Config) Mount(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}) {
	mux.HandleFunc("POST /auth/login", c.login)
	mux.HandleFunc("GET /auth/session", c.session)
	mux.HandleFunc("POST /auth/logout", c.logout)
}

func (c Config) login(w http.ResponseWriter, r *http.Request) {
	if !security.Origin(r, c.Origins) {
		response.OriginDenied.Write(w)
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !security.Equal(input.Password, c.OperatorPassword) {
		response.InvalidAccessCode.Write(w)
		return
	}
	value, err := c.Sessions.Create(r.Context())
	if err != nil {
		response.SessionUnavailable.Write(w)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: security.CookieName, Value: value, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteStrictMode, Secure: strings.HasPrefix(r.Header.Get("Origin"), "https://"), MaxAge: 3600})
	response.JSON(w, 200, map[string]bool{"authenticated": true})
}

func (c Config) session(w http.ResponseWriter, r *http.Request) {
	principal, ok, err := c.Sessions.Authenticate(r.Context(), security.Token(r))
	if err != nil {
		response.SessionUnavailable.Write(w)
		return
	}
	if !ok {
		response.JSON(w, 401, map[string]bool{"authenticated": false})
		return
	}
	response.JSON(w, 200, principal)
}

func (c Config) logout(w http.ResponseWriter, r *http.Request) {
	if !security.Origin(r, c.Origins) {
		response.OriginDenied.Write(w)
		return
	}
	if err := c.Sessions.Revoke(r.Context(), security.Token(r)); err != nil {
		response.SessionUnavailable.Write(w)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: security.CookieName, Value: "", Path: "/", HttpOnly: true,
		SameSite: http.SameSiteStrictMode, MaxAge: -1})
	response.JSON(w, 200, map[string]bool{"authenticated": false})
}
