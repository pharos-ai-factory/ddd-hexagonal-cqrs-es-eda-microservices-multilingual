package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ChangeQuantityCommand expresses the owner ChangeQuantity use case independently of transport.
type ChangeQuantityCommand struct {
	LineID   string `json:"lineId"`
	Quantity int    `json:"quantity"`
}

// ChangeQuantityCommandHandler applies ChangeQuantity through one aggregate command transaction.
type ChangeQuantityCommandHandler struct {
	Repository ports.OrderWriteRepository
}

func (h ChangeQuantityCommandHandler) Execute(ctx context.Context, m a.CommandContext, c ChangeQuantityCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, orderNotFound()
	}
	order := s.State
	if err = order.ChangeQuantity(c.LineID, c.Quantity); err != nil {
		return a.CommandResult{}, err
	}
	if len(order.Events()) > 0 {
		if err = h.Repository.Save(ctx, order); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("draft"), nil
}
