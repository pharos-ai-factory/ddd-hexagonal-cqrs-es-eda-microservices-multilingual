package domain

// DrinkRevisedDomainEvent records the completed DrinkRevised transition.
type DrinkRevisedDomainEvent struct{ State DrinkState }

func (DrinkRevisedDomainEvent) DomainEvent() {}

// DrinkPublishedDomainEvent records the completed DrinkPublished transition.
type DrinkPublishedDomainEvent struct{ State DrinkState }

func (DrinkPublishedDomainEvent) DomainEvent() {}

// MenuOfferAddedDomainEvent records the completed MenuOfferAdded transition.
type MenuOfferAddedDomainEvent struct{ State OfferState }

func (MenuOfferAddedDomainEvent) DomainEvent() {}

// MenuOfferChangedDomainEvent records the completed MenuOfferChanged transition.
type MenuOfferChangedDomainEvent struct{ State OfferState }

func (MenuOfferChangedDomainEvent) DomainEvent() {}

// MenuEditionPublishedDomainEvent records the completed MenuEditionPublished transition.
type MenuEditionPublishedDomainEvent struct{ State EditionState }

func (MenuEditionPublishedDomainEvent) DomainEvent() {}
