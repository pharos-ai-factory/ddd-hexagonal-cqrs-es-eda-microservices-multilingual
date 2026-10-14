package application

// DrinkPublishedDomainEvent is an owner-private application delivery value.
type DrinkPublishedDomainEvent struct {
	DrinkID  string `json:"drinkId"`
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
}
