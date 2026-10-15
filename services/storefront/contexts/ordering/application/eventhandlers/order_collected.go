package eventhandlers

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/model"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderCollectedIntegrationEventHandler durably requests an owner transition before acknowledging the fact.
type OrderCollectedIntegrationEventHandler struct {
	Commands a.CommandBusPort[orderingcontracts.RecordCollectionCommand]
}

func (h OrderCollectedIntegrationEventHandler) Handle(ctx context.Context, m a.Metadata, event model.OrderCollectedIntegrationEvent) (a.Outcome, error) {
	return h.Commands.Enqueue(ctx, m, orderingcontracts.RecordCollectionCommand{CustomerID: event.CustomerID})
}
