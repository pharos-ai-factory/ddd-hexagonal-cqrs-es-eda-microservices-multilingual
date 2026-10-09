package main

import (
	"context"
	"fmt"
	runtimeconfig "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/config"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/apps/support"
	"go.uber.org/fx"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// APISettings contains validated development composition values, never business policy.
type APISettings map[string]string

func settings() (APISettings, error) {
	if value := os.Getenv("APP_ENV"); value != "local" && value != "development" {
		return nil, fmt.Errorf("development environments only")
	}
	values := APISettings{}
	for _, name := range []string{"VALKEY_ADDRESS", "SESSION_PASSWORD", "REALTIME_GATEWAY_URL", "SESSION_REALTIME_KEY", "API_BROKER_URL", "OPERATOR_PASSWORD", "API_KEY", "CONNECT_PROXY_SECRET", "WEB_ORIGINS"} {
		value, err := runtimeconfig.Secret(name)
		if err != nil {
			return nil, err
		}
		values[name] = value
	}
	return values, nil
}
func sessionStore(lifecycle fx.Lifecycle, scope *support.ResourceScope, settings APISettings) (*sessions.ValkeySessionStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store, err := sessions.Open(ctx, settings["VALKEY_ADDRESS"], settings["SESSION_PASSWORD"], settings["REALTIME_GATEWAY_URL"], settings["SESSION_REALTIME_KEY"])
	if err != nil {
		return nil, err
	}
	scope.Add(store.Close)
	return store, nil
}

// The store dependency ensures worker shutdown precedes store disposal in reverse hook order.
func workers(lifecycle fx.Lifecycle, scope *support.ResourceScope, _ *sessions.ValkeySessionStore) *support.WorkerGroup {
	group := support.NewWorkerGroup(lifecycle)
	scope.Add(group.Close)
	return group
}
func server(lifecycle fx.Lifecycle, group *support.WorkerGroup, store *sessions.ValkeySessionStore, requests *messaging.RabbitMQRequestClient, settings APISettings) *http.Server {
	owners := []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"}
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		group.Start(store.Run)
		group.Start(func(ctx context.Context) { requests.Run(ctx, settings["API_BROKER_URL"], owners) })
		return nil
	}})
	config := web.Config{Diagnostics: func(ctx context.Context) (any, error) { return store.Diagnostics(ctx) }, Sessions: store,
		OperatorPassword: settings["OPERATOR_PASSWORD"], CLIKey: settings["API_KEY"], ProxySecret: settings["CONNECT_PROXY_SECRET"],
		Origins: strings.Split(settings["WEB_ORIGINS"], ","), Owners: owners, Requests: requests}
	return &http.Server{Addr: ":8080", Handler: config.Handler(), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 30 * time.Second}
}
func composition() fx.Option {
	return fx.Options(fx.Provide(settings, sessionStore, workers, messaging.NewClient, server), fx.Invoke(support.Serve))
}
func main() {
	if err := support.Run(composition()); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
