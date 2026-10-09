package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/config"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"log"
	"os"
	"time"
)

func main() {
	if os.Getenv("APP_ENV") != "development" || len(os.Args) != 2 {
		log.Fatal("usage: APP_ENV=development BROKER_URL=... replay <consumer>")
	}
	var entries []struct {
		Consumer string
		Command  bool
	}
	if err := json.Unmarshal(topology, &entries); err != nil {
		log.Fatal(err)
	}
	known := false
	for _, entry := range entries {
		known = known || entry.Consumer == os.Args[1] || (entry.Command && entry.Consumer+".command" == os.Args[1])
	}
	if !known {
		log.Fatal("consumer is not in the event catalogue")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url, err := config.Secret("BROKER_URL")
	if err != nil {
		log.Fatal(err)
	}
	replayed, err := broker.Replay(ctx, url, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("replayed=%t\n", replayed)
}

//go:embed generated.json
var topology []byte
