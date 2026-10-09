package application

// DrinkPublished is an owner-private application delivery value.
type DrinkPublished struct {
	DrinkID  string `json:"drinkId"`
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
}
