package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

const cookieName = "cafe_session"

type Backend struct{ URL, Key string }
type Config struct {
	Diagnostics                           func(context.Context) (any, error)
	Sessions                              a.Sessions
	OperatorPassword, CLIKey, ProxySecret string
	Origins                               []string
	Backends                              map[string]Backend
}

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func token(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
func (c Config) origin(r *http.Request) bool {
	for _, allowed := range c.Origins {
		if r.Header.Get("Origin") == allowed {
			return true
		}
	}
	return false
}
func (c Config) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /diagnostics", func(w http.ResponseWriter, r *http.Request) {
		if c.CLIKey == "" || !equal(r.Header.Get("Authorization"), "Bearer "+c.CLIKey) {
			jsonResponse(w, 401, map[string]string{"code": "unauthorised"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if c.Diagnostics == nil {
			jsonResponse(w, 503, map[string]string{"code": "diagnostics_unavailable"})
			return
		}
		result, err := c.Diagnostics(ctx)
		if err != nil {
			jsonResponse(w, 503, map[string]string{"code": "diagnostics_unavailable"})
			return
		}
		jsonResponse(w, 200, result)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, map[string]string{"status": "ok", "role": "client-api"})
	})
	mux.HandleFunc("POST /auth/login", func(w http.ResponseWriter, r *http.Request) {
		if !c.origin(r) {
			jsonResponse(w, 403, map[string]string{"code": "origin_denied"})
			return
		}
		var input struct {
			Password string `json:"password"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || !equal(input.Password, c.OperatorPassword) {
			jsonResponse(w, 401, map[string]string{"code": "invalid_access_code"})
			return
		}
		value, err := c.Sessions.Create(r.Context())
		if err != nil {
			jsonResponse(w, 503, map[string]string{"code": "session_unavailable"})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: value, Path: "/", HttpOnly: true,
			SameSite: http.SameSiteStrictMode, Secure: strings.HasPrefix(r.Header.Get("Origin"), "https://"), MaxAge: 3600})
		jsonResponse(w, 200, map[string]bool{"authenticated": true})
	})
	mux.HandleFunc("GET /auth/session", func(w http.ResponseWriter, r *http.Request) {
		principal, ok, err := c.Sessions.Authenticate(r.Context(), token(r))
		if err != nil {
			jsonResponse(w, 503, map[string]string{"code": "session_unavailable"})
			return
		}
		if !ok {
			jsonResponse(w, 401, map[string]bool{"authenticated": false})
			return
		}
		jsonResponse(w, 200, principal)
	})
	mux.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if !c.origin(r) {
			jsonResponse(w, 403, map[string]string{"code": "origin_denied"})
			return
		}
		if err := c.Sessions.Revoke(r.Context(), token(r)); err != nil {
			jsonResponse(w, 503, map[string]string{"code": "session_unavailable"})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", HttpOnly: true,
			SameSite: http.SameSiteStrictMode, MaxAge: -1})
		jsonResponse(w, 200, map[string]bool{"authenticated": false})
	})
	mux.HandleFunc("POST /api/realtime/connect", func(w http.ResponseWriter, r *http.Request) {
		if !equal(r.Header.Get("X-Cafe-Realtime-Proxy"), c.ProxySecret) {
			jsonResponse(w, 403, map[string]string{"code": "proxy_denied"})
			return
		}
		// Calculate before reading authority: every response authorised before
		// session deletion must expire within the worker's revocation fence.
		expires := time.Now().Add(a.ConnectLifetime).Unix()
		principal, ok, err := c.Sessions.Authenticate(r.Context(), token(r))
		if err != nil || !ok || !c.origin(r) {
			jsonResponse(w, 200, map[string]any{"disconnect": map[string]any{"code": 4501, "reason": "unauthorised"}})
			return
		}
		subs := map[string]any{}
		for _, channel := range a.AuthorisedChannels(principal) {
			subs[channel] = map[string]any{}
		}
		jsonResponse(w, 200, map[string]any{"result": map[string]any{
			"user": principal.Subject, "subs": subs, "expire_at": expires}})
	})
	mux.HandleFunc("POST /api/realtime/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !equal(r.Header.Get("X-Cafe-Realtime-Proxy"), c.ProxySecret) {
			jsonResponse(w, 403, map[string]string{"code": "proxy_denied"})
			return
		}
		// Centrifugo forwards the original connection's Cookie and Origin.
		// Refresh cannot create a connection or revive one already disconnected.
		expires := time.Now().Add(time.Minute).Unix()
		_, ok, err := c.Sessions.Authenticate(r.Context(), token(r))
		if err != nil || !ok || !c.origin(r) {
			jsonResponse(w, 200, map[string]any{"result": map[string]bool{"expired": true}})
			return
		}
		jsonResponse(w, 200, map[string]any{"result": map[string]int64{"expire_at": expires}})
	})
	for owner, backend := range c.Backends {
		target, err := url.Parse(backend.URL)
		if err != nil || target.Host == "" || target.Scheme != "http" {
			panic("invalid internal service URL")
		}
		proxy := &httputil.ReverseProxy{
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
				jsonResponse(w, 503, map[string]string{"code": "temporarily_unavailable"})
			},
		}
		mux.Handle("/api/v1/"+owner+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodPost {
				w.WriteHeader(405)
				return
			}
			cli := equal(r.Header.Get("Authorization"), "Bearer "+c.CLIKey)
			if !cli {
				_, ok, err := c.Sessions.Authenticate(r.Context(), token(r))
				if err != nil {
					jsonResponse(w, 503, map[string]string{"code": "session_unavailable"})
					return
				}
				if !ok {
					jsonResponse(w, 401, map[string]string{"code": "unauthorised"})
					return
				}
				if r.Method == http.MethodPost && !c.origin(r) {
					jsonResponse(w, 403, map[string]string{"code": "origin_denied"})
					return
				}
			}
			r.Body = http.MaxBytesReader(w, r.Body, 65536)
			proxy.ServeHTTP(w, r)
		}))
	}
	return mux
}
