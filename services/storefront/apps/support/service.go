package support

import (
	"context"
	"errors"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	realtime "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	workers "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/runtime"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Service struct {
	Context       context.Context
	Mux           *http.ServeMux
	Databases     map[string]*postgres.Database
	URLs          map[string]string
	subscriptions []broker.Subscription
	cancel        context.CancelFunc
}

func Open(contexts ...string) (*Service, error) {
	environment := os.Getenv("APP_ENV")
	if environment != "development" && environment != "local" {
		return nil, fmt.Errorf("this reference only runs in local or development environments")
	}
	if len(os.Getenv("API_KEY")) < 32 {
		return nil, fmt.Errorf("a development API key of at least 32 characters is required")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	service := &Service{Context: ctx, Mux: http.NewServeMux(), Databases: map[string]*postgres.Database{}, URLs: map[string]string{}, cancel: cancel}
	for _, owner := range contexts {
		db, err := postgres.Open(ctx, os.Getenv(strings.ToUpper(owner)+"_DATABASE_URL"), owner, codec.Encode)
		if err != nil {
			service.Close()
			return nil, err
		}
		service.Databases[owner] = db
		if err = db.EnableRealtime(ctx, realtime.Encode); err != nil {
			service.Close()
			return nil, err
		}
		service.URLs[owner] = os.Getenv(strings.ToUpper(owner) + "_BROKER_URL")
	}
	service.Mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { web.JSON(w, 200, map[string]string{"status": "ok"}) })
	return service, nil
}
func (s *Service) Close() {
	s.cancel()
	for _, db := range s.Databases {
		db.Close()
	}
}
func (s *Service) Run() error {
	defer s.Close()
	for owner, db := range s.Databases {
		go workers.Relay(s.Context, db, s.URLs[owner], codec.Decode)
		go workers.Realtime(s.Context, db, os.Getenv("CENTRIFUGO_API_URL"), os.Getenv("CENTRIFUGO_API_KEY"))
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
	server := &http.Server{Addr: address, Handler: web.Auth(os.Getenv("API_KEY"), s.Mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		<-s.Context.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()
	slog.Info("reference service ready", "address", address)
	err := server.ListenAndServe()
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
	if s.Databases[owner] == nil {
		panic("consumer bound outside its context")
	}
	binding := broker.Binding{Consumer: consumer, Event: definition.Name, Visibility: string(definition.Visibility), Context: owner}
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
