package readmodels

// OfferView is the application-owned read representation.
type OfferView struct {
	Code          string `json:"code"`
	DrinkID       string `json:"drinkId"`
	DrinkRevision uint64 `json:"drinkRevision"`
	Name          string `json:"name"`
	Minor         int64  `json:"minor"`
	Currency      string `json:"currency"`
}

// EditionView is the application-owned read representation.
type EditionView struct {
	ID       string      `json:"id"`
	Currency string      `json:"currency"`
	Status   string      `json:"status"`
	Offers   []OfferView `json:"offers"`
}
