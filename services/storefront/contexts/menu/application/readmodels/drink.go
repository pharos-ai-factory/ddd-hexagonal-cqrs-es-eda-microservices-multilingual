package readmodels

// DrinkView is the application-owned read representation.
type DrinkView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Revision  uint64 `json:"revision"`
	Published bool   `json:"published"`
}
