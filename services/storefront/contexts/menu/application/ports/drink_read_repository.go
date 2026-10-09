package ports

import (
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkReadRepository supplies versioned drink views for the query use cases.
type DrinkReadRepository interface {
	a.PagedQueryPort[view.DrinkView]
}
