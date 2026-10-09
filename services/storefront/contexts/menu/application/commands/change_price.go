package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ChangePriceCommand expresses the owner ChangePrice use case independently of transport.
type ChangePriceCommand struct {
	Code  string `json:"code"`
	Minor int64  `json:"minor"`
}

// ChangePriceCommandHandler applies ChangePrice through one aggregate command transaction.
type ChangePriceCommandHandler struct {
	Repository ports.EditionWriteRepository
}

func (h ChangePriceCommandHandler) Execute(ctx context.Context, m a.CommandContext, c ChangePriceCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, core.Reject("not_found", "The edition does not exist")
	}
	edition := s.State
	if err = edition.ChangePrice(c.Code, c.Minor); err != nil {
		return a.CommandResult{}, err
	}
	if len(edition.Events()) > 0 {
		if err = h.Repository.Save(ctx, edition); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("draft"), nil
}
