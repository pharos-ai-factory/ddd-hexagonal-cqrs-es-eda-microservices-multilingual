package commands

import (
	"context"
	"errors"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ChangeQuantityCommandHandler applies ChangeQuantity through one aggregate command transaction.
type ChangeQuantityCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.OrderWriteRepository
}

func (h ChangeQuantityCommandHandler) Execute(ctx context.Context, m a.CommandContext, c orderingcontracts.ChangeQuantityCommand) (result a.CommandResult, err error) {
	defer func() {
		var rejected d.ExpectedError
		if errors.As(err, &rejected) {
			detail := rejected.Rejection()
			if publishErr := h.Realtime.Publish(ctx, a.RealtimeEvent{Rejection: &detail}); publishErr != nil {
				err = publishErr
			}
		}
	}()
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
		if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: order.Snapshot()}); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("draft"), nil
}
