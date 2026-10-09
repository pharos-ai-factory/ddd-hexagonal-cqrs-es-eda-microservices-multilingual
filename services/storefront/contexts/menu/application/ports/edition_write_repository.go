package ports

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// EditionWriteRepository loads and saves the authoritative aggregate inside its command transaction.
type EditionWriteRepository = a.WriteRepository[*d.MenuEdition]
