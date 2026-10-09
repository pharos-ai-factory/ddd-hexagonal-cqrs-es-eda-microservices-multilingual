package ports

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkWriteRepository loads and saves the authoritative aggregate inside its command transaction.
type DrinkWriteRepository = a.WriteRepository[*d.Drink]
