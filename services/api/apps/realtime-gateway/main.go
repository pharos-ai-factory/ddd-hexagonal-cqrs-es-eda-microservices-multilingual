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

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/config"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/realtime"
)

func required(name string) string {
	value, err := config.Secret(name)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func main() {
	if value := os.Getenv("APP_ENV"); value != "local" && value != "development" {
		log.Fatal("development environments only")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	settings := realtime.Config{UpstreamURL: required("CENTRIFUGO_API_URL"), UpstreamKey: required("CENTRIFUGO_API_KEY"),
		SessionKey: required("SESSION_REALTIME_KEY"), PublisherKeys: map[string]string{}}
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		settings.PublisherKeys[owner] = required(strings.ToUpper(owner) + "_REALTIME_KEY")
	}
	handler, err := settings.Handler()
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: ":8080", Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Print("Realtime authority gateway ready")
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
