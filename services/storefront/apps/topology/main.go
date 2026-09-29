package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/config"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"log"
	"os"
	"strings"
)

func main() {
	if os.Getenv("APP_ENV") != "development" {
		log.Fatal("topology bootstrap is development-only")
	}
	var bindings []broker.Binding
	for _, event := range model.Catalogue {
		for _, consumer := range event.Consumers {
			bindings = append(bindings, broker.Binding{Consumer: consumer, Event: event.Name, Visibility: string(event.Visibility), Context: strings.SplitN(consumer, ".", 2)[0]})
		}
	}
	url, err := config.Secret("BROKER_ADMIN_URL")
	if err != nil {
		log.Fatal(err)
	}
	if err := broker.Declare(url, bindings); err != nil {
		log.Fatal(err)
	}
}
