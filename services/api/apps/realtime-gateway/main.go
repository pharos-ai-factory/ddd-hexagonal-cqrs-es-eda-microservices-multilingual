package main

import (
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/apps/support"
	"go.uber.org/fx"
	"log"
	"net/http"
	"os"
	"strings"
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

func server() (*http.Server, error) {
	if value := os.Getenv("APP_ENV"); value != "local" && value != "development" {
		return nil, fmt.Errorf("development environments only")
	}
	settings := realtime.Config{UpstreamURL: required("CENTRIFUGO_API_URL"), UpstreamKey: required("CENTRIFUGO_API_KEY"),
		SessionKey: required("SESSION_REALTIME_KEY"), PublisherKeys: map[string]string{}}
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		settings.PublisherKeys[owner] = required(strings.ToUpper(owner) + "_REALTIME_KEY")
	}
	handler, err := settings.Handler()
	if err != nil {
		return nil, err
	}
	server := &http.Server{Addr: ":8080", Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	return server, nil
}
func main() { fx.New(fx.Provide(server), fx.Invoke(support.Serve)).Run() }
