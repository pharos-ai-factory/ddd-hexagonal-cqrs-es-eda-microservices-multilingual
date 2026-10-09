package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// PlaceOrderCommand expresses the owner use case independently of its transport.
type PlaceOrderCommand struct{}

// PlaceOrderCommandHandler applies PlaceOrder through one aggregate command transaction.
type PlaceOrderCommandHandler struct {
	Repository ports.OrderWriteRepository
}

func (h PlaceOrderCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ PlaceOrderCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, orderNotFound()
	}
	order := s.State
	if err = order.Place(); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, order); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("placed"), nil
}
