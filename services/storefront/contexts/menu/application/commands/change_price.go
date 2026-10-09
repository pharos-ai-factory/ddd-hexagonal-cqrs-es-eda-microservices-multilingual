package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
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
	Editions a.AggregateCommandPort[d.EditionState]
}

func (h ChangePriceCommandHandler) Execute(ctx context.Context, m a.Metadata, c ChangePriceCommand) (a.Outcome, error) {
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if !s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("not_found", "The edition does not exist")
		}
		edition, err := d.RestoreEdition(s.State)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		if err = edition.ChangePrice(c.Code, c.Minor); err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		return a.Mutation[d.EditionState]{State: edition.Snapshot(), Changed: len(edition.Events()) > 0, Status: "draft"}, nil
	})
}
