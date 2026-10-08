package main

import (
	"embed"
	"encoding/json"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/config"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"log"
	"os"
	"strings"
)

//go:embed generated.json
var definitions embed.FS

func main() {
	if os.Getenv("APP_ENV") != "development" {
		log.Fatal("topology bootstrap is development-only")
	}
	var bindings []broker.Binding
	data, err := definitions.ReadFile("generated.json")
	if err != nil {
		log.Fatal(err)
	}
	var events []struct {
		Name       string
		Consumer   string
		Visibility string
	}
	if err = json.Unmarshal(data, &events); err != nil {
		log.Fatal(err)
	}
	for _, event := range events {
		bindings = append(bindings, broker.Binding{Consumer: event.Consumer, Event: event.Name, Visibility: event.Visibility, Context: strings.SplitN(event.Consumer, ".", 2)[0]})
	}
	url, err := config.Secret("BROKER_ADMIN_URL")
	if err != nil {
		log.Fatal(err)
	}
	if err := broker.DeclareRequests(url, []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"}); err != nil {
		log.Fatal(err)
	}
	if err := broker.Declare(url, bindings); err != nil {
		log.Fatal(err)
	}
}
