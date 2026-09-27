package main

import (
	"context"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"log"
	"os"
	"time"
)

func main() {
	if os.Getenv("APP_ENV") != "development" || len(os.Args) != 2 {
		log.Fatal("usage: APP_ENV=development BROKER_URL=... replay <consumer>")
	}
	known := false
	for _, event := range model.Catalogue {
		for _, consumer := range event.Consumers {
			known = known || consumer == os.Args[1]
		}
	}
	if !known {
		log.Fatal("consumer is not in the event catalogue")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	replayed, err := broker.Replay(ctx, os.Getenv("BROKER_URL"), os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("replayed=%t\n", replayed)
}
