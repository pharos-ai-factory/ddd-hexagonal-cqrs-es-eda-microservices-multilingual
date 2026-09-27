// A disposable test host for the real API handler, Valkey store and disconnect worker.
// The gate delays one successful connect response; it does not supply authentication.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
	"github.com/redis/go-redis/v9"
)

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal(name + " is required")
	}
	return value
}

func main() {
	if !strings.HasPrefix(required("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		log.Fatal("the revocation fixture requires a disposable integration project")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	target, err := url.Parse(required("CENTRIFUGO_API_URL"))
	if err != nil {
		log.Fatal(err)
	}
	var accepted, refreshed atomic.Int64
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(response *http.Response) error {
		if response.Request.URL.Path != "/api/disconnect" {
			return nil
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return err
		}
		response.Body = io.NopCloser(bytes.NewReader(body))
		var result struct {
			Error json.RawMessage `json:"error"`
		}
		if err = json.Unmarshal(body, &result); err != nil {
			return err
		}
		if response.StatusCode == http.StatusOK && len(result.Error) == 0 {
			accepted.Add(1)
		}
		return nil
	}
	observer := httptest.NewServer(proxy)
	defer observer.Close()
	store, err := sessions.Open(ctx, required("VALKEY_ADDRESS"), required("SESSION_PASSWORD"),
		observer.URL, required("CENTRIFUGO_API_KEY"))
	if err != nil {
		log.Fatal("fixture session store unavailable")
	}
	defer store.Close()
	go store.Run(ctx)
	client := redis.NewClient(&redis.Options{Addr: required("VALKEY_ADDRESS"), Username: "sessions",
		Password: required("SESSION_PASSWORD")})
	defer client.Close()
	api := (web.Config{Sessions: store, OperatorPassword: required("OPERATOR_PASSWORD"),
		CLIKey: required("API_KEY"), ProxySecret: required("CONNECT_PROXY_SECRET"),
		Origins: []string{required("WEB_ORIGIN")}}).Handler()
	var hold, captured, released, cancelled atomic.Bool
	release := make(chan struct{})
	var once sync.Once
	mux := http.NewServeMux()
	mux.HandleFunc("POST /test/hold", func(w http.ResponseWriter, r *http.Request) {
		hold.Store(true)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /test/release", func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(release) })
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /test/state", func(w http.ResponseWriter, r *http.Request) {
		pending, err := client.SCard(r.Context(), "cafe:auth:disconnects").Result()
		if err != nil {
			http.Error(w, "session evidence unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"captured": captured.Load(), "released": released.Load(), "cancelled": cancelled.Load(),
			"disconnects": accepted.Load(), "pending": pending,
			"refreshes": refreshed.Load(),
		})
	})
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/realtime/refresh" {
			recorded := httptest.NewRecorder()
			api.ServeHTTP(recorded, r)
			var response struct {
				Result struct {
					ExpireAt int64 `json:"expire_at"`
				}
			}
			if json.Unmarshal(recorded.Body.Bytes(), &response) == nil && response.Result.ExpireAt > 0 {
				refreshed.Add(1)
			}
			for key, values := range recorded.Header() {
				w.Header()[key] = values
			}
			w.WriteHeader(recorded.Code)
			_, _ = w.Write(recorded.Body.Bytes())
			return
		}
		if r.URL.Path != "/api/realtime/connect" || !hold.CompareAndSwap(true, false) {
			api.ServeHTTP(w, r)
			return
		}
		recorded := httptest.NewRecorder()
		api.ServeHTTP(recorded, r)
		var response struct {
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(recorded.Body.Bytes(), &response); err != nil || len(response.Result) == 0 {
			http.Error(w, "fixture did not capture an authorised response", http.StatusInternalServerError)
			return
		}
		captured.Store(true)
		select {
		case <-release:
			if r.Context().Err() != nil {
				cancelled.Store(true)
				return
			}
			for key, values := range recorded.Header() {
				w.Header()[key] = values
			}
			w.WriteHeader(recorded.Code)
			if _, err := w.Write(recorded.Body.Bytes()); err != nil {
				cancelled.Store(true)
			} else {
				released.Store(true)
			}
		case <-r.Context().Done():
			cancelled.Store(true)
		}
	}))
	server := &http.Server{Addr: required("LISTEN_ADDR"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		once.Do(func() { close(release) })
		shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = server.Shutdown(shutdown)
	}()
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("revocation fixture: %w", err))
	}
}
