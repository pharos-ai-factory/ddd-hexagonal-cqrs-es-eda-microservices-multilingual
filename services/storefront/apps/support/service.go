package support

import (
	"context"
	"errors"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	realtime "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	config "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/config"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	workers "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/runtime"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Service struct {
	Context        context.Context
	Mux            *contract.Mux
	Databases      map[string]*postgres.Database
	RequestWorkers map[string]func(context.Context, string)
	URLs           map[string]string
	subscriptions  []broker.Subscription
	cancel         context.CancelFunc
	apiKey         string
}

func Open(contexts ...string) (*Service, error) {
	environment := os.Getenv("APP_ENV")
	if environment != "development" && environment != "local" {
		return nil, fmt.Errorf("this reference only runs in local or development environments")
	}
	apiKey, err := config.Secret("API_KEY")
	if err != nil {
		return nil, err
	}
	if len(apiKey) < 32 {
		return nil, fmt.Errorf("a development API key of at least 32 characters is required")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	service := &Service{Context: ctx, Mux: contract.NewMux("storefront", nil), Databases: map[string]*postgres.Database{}, URLs: map[string]string{}, RequestWorkers: map[string]func(context.Context, string){}, cancel: cancel, apiKey: apiKey}
	for _, owner := range contexts {
		databaseURL, err := config.Secret(strings.ToUpper(owner) + "_DATABASE_URL")
		if err != nil {
			service.Close()
			return nil, err
		}
		brokerURL, err := config.Secret(strings.ToUpper(owner) + "_BROKER_URL")
		if err != nil {
			service.Close()
			return nil, err
		}
		db, err := postgres.Open(ctx, databaseURL, owner, codec.Encode)
		if err != nil {
			service.Close()
			return nil, err
		}
		service.Databases[owner] = db
		if err = db.EnableRealtime(ctx, realtime.Encode); err != nil {
			service.Close()
			return nil, err
		}
		service.URLs[owner] = brokerURL
	}
	service.mountHTTP()
	return service, nil
}
func (s *Service) mountHTTP() {
	s.Mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { web.JSON(w, 200, map[string]string{"status": "ok"}) })
	s.Mux.HandleFunc("GET /diagnostics", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		databases := map[string]postgres.Diagnostics{}
		for owner, db := range s.Databases {
			state, err := db.Diagnostics(ctx)
			if err != nil {
				diagnostics.Record(owner, "diagnostics", "", "", err, false)
				web.JSON(w, 503, map[string]string{"code": "diagnostics_unavailable"})
				return
			}
			databases[owner] = state
		}
		w.Header().Set("Cache-Control", "no-store")
		web.JSON(w, 200, map[string]any{"databases": databases, "processWorkers": diagnostics.Snapshot(), "counterScope": "process lifetime; deadLetterTransfers are confirmed transfers, not queue depths"})
	})
}

func (s *Service) Close() {
	s.cancel()
	for _, db := range s.Databases {
		db.Close()
	}
}
func (s *Service) Run() error {
	defer s.Close()
	gatewayURL, err := config.Secret("REALTIME_GATEWAY_URL")
	if err != nil {
		return err
	}
	for owner, db := range s.Databases {
		realtimeKey, err := config.Secret(strings.ToUpper(owner) + "_REALTIME_KEY")
		if err != nil {
			return err
		}
		go workers.Relay(s.Context, db, s.URLs[owner], codec.Decode)
		go workers.Replies(s.Context, db, s.URLs[owner], owner)
		go workers.Realtime(s.Context, db, gatewayURL, realtimeKey)
	}
	for owner, run := range s.RequestWorkers {
		if s.Databases[owner] == nil {
			return fmt.Errorf("request worker bound outside its owner")
		}
		go run(s.Context, s.URLs[owner])
	}
	paused := "," + os.Getenv("PAUSED_CONSUMERS") + ","
	for _, sub := range s.subscriptions {
		if strings.Contains(paused, ","+sub.Binding.Consumer+",") {
			slog.Info("consumer intentionally paused", "consumer", sub.Binding.Consumer)
			continue
		}
		go workers.Consumer(s.Context, s.URLs[sub.Binding.Context], sub, codec.Decode, postgres.DerivedID)
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{Addr: address, Handler: web.Auth(s.apiKey, s.Mux.Handler()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		<-s.Context.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()
	slog.Info("reference service ready", "address", address)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func Subscribe[P any](s *Service, consumer string, target func(P) string, handle func(context.Context, a.Metadata, P) (a.Outcome, error)) {
	var definition model.Definition
	found := false
	for _, d := range model.Catalogue {
		for _, c := range d.Consumers {
			if c == consumer {
				definition = d
				found = true
			}
		}
	}
	if !found {
		panic("unregistered consumer: " + consumer)
	}
	owner := strings.SplitN(consumer, ".", 2)[0]
	binding := broker.Binding{Consumer: consumer, Event: definition.Name, Visibility: string(definition.Visibility), Context: owner}
	bindSubscription(s, binding, target, handle)
}

// SubscribePrivate binds owner-internal delivery explicitly in the composition;
// it does not add that message to the published integration catalogue.
func SubscribePrivate[P any](s *Service, binding broker.Binding, target func(P) string, handle func(context.Context, a.Metadata, P) (a.Outcome, error)) {
	if binding.Visibility != string(a.Private) || !strings.HasPrefix(binding.Event, binding.Context+".") {
		panic("invalid private owner binding")
	}
	bindSubscription(s, binding, target, handle)
}
func bindSubscription[P any](s *Service, binding broker.Binding, target func(P) string, handle func(context.Context, a.Metadata, P) (a.Outcome, error)) {
	if s.Databases[binding.Context] == nil || strings.SplitN(binding.Consumer, ".", 2)[0] != binding.Context {
		panic("consumer bound outside its context")
	}
	s.subscriptions = append(s.subscriptions, broker.Subscription{Binding: binding, Handle: func(ctx context.Context, metadata a.Metadata, message a.Message) error {
		payload, ok := message.Payload.(P)
		if !ok {
			return fmt.Errorf("consumer payload type disagrees with contract")
		}
		metadata.AggregateID = target(payload)
		result, err := handle(ctx, metadata, payload)
		if err != nil {
			return err
		}
		if result.Rejection != nil {
			return result.Rejection
		}
		return nil
	}})
}
func Main(build func() (*Service, error)) {
	s, err := build()
	if err == nil {
		err = s.Run()
	}
	if err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
