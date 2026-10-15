package eventhandlers

import (
	"context"
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderCollectedIntegrationEventHandler durably requests an owner transition before acknowledging the fact.
type OrderCollectedIntegrationEventHandler struct {
	Commands a.CommandBusPort[commands.RecordCollectionCommand]
}

func (h OrderCollectedIntegrationEventHandler) Handle(ctx context.Context, m a.Metadata, event model.OrderCollectedIntegrationEvent) (a.Outcome, error) {
	return h.Commands.Enqueue(ctx, m, commands.RecordCollectionCommand{CustomerID: event.CustomerID})
}
