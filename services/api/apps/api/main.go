package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
)

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal(name + " is required")
	}
	return value
}
func main() {
	if value := os.Getenv("APP_ENV"); value != "local" && value != "development" {
		log.Fatal("development environments only")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	store, err := sessions.Open(ctx, required("VALKEY_ADDRESS"), required("SESSION_PASSWORD"),
		required("CENTRIFUGO_API_URL"), required("CENTRIFUGO_API_KEY"))
	if err != nil {
		log.Fatal("session store unavailable")
	}
	defer store.Close()
	go store.Run(ctx)
	backends := map[string]web.Backend{}
	for service, owners := range map[string][]string{
		"storefront": {"menu", "ordering"}, "operations": {"preparation", "collection"},
		"engagement": {"loyalty", "communication"},
	} {
		prefix := strings.ToUpper(service)
		for _, owner := range owners {
			backends[owner] = web.Backend{URL: required(prefix + "_URL"), Key: required(prefix + "_API_KEY")}
		}
	}
	config := web.Config{Sessions: store, OperatorPassword: required("OPERATOR_PASSWORD"), CLIKey: required("API_KEY"),
		ProxySecret: required("CONNECT_PROXY_SECRET"), Origins: strings.Split(required("WEB_ORIGINS"), ","), Backends: backends}
	server := &http.Server{Addr: ":8080", Handler: config.Handler(), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Print("Go client API ready")
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
