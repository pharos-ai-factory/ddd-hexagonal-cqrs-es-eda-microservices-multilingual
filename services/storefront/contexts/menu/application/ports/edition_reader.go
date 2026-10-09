package ports

import (
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// EditionReader supplies versioned edition views for the query use cases.
type EditionReader interface {
	a.PagedQueryPort[view.EditionView]
}
