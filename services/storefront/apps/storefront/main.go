package main

import (
	"log"
	"os"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/storefront/di"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
)

func main() {
	if err := support.Run(di.Module()); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
