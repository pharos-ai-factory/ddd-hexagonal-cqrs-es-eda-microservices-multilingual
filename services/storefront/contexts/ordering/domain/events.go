package domain

// OrderLineAddedDomainEvent records the completed OrderLineAdded transition.
type OrderLineAddedDomainEvent struct{ State LineState }

func (OrderLineAddedDomainEvent) DomainEvent() {}

// OrderLineQuantityChangedDomainEvent records the completed OrderLineQuantityChanged transition.
type OrderLineQuantityChangedDomainEvent struct{ State LineState }

func (OrderLineQuantityChangedDomainEvent) DomainEvent() {}

// OrderPlacedDomainEvent records the completed OrderPlaced transition.
type OrderPlacedDomainEvent struct{ State OrderState }

func (OrderPlacedDomainEvent) DomainEvent() {}
