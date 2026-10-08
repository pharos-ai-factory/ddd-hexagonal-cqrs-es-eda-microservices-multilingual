package model

import "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"

type Definition struct {
	Name          string
	Owner         string
	Visibility    application.Visibility
	Consumers     []string
	AggregateKind string
}

// Catalogue is closed: an unknown or incorrectly scoped event cannot be routed.
var Catalogue = []Definition{
	{"menu.edition-published", "menu", application.Public, []string{"ordering.menu-directory"}, "edition"},
	{"ordering.order-placed", "ordering", application.Public, []string{"preparation.accept-order"}, "order"},
	{"preparation.drinks-ready", "preparation", application.Public, []string{"collection.open-pickup"}, "ticket"},
	{"collection.pickup-opened", "collection", application.Public, []string{"communication.pickup-notice"}, "pickup"},
	{"collection.order-collected", "collection", application.Public, []string{"loyalty.credit-collection"}, "pickup"},
	{"loyalty.reward-issued", "loyalty", application.Public, []string{"communication.reward-notice"}, "reward"},
}

func Lookup(name string) (Definition, bool) {
	for _, definition := range Catalogue {
		if definition.Name == name {
			return definition, true
		}
	}
	return Definition{}, false
}
