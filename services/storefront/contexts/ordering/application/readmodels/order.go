package readmodels

// SelectionView is the application-owned read representation.
type SelectionView struct {
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Minor     int64  `json:"minor"`
}

// LineView is the application-owned read representation.
type LineView struct {
	ID        string        `json:"id"`
	Selection SelectionView `json:"selection"`
	Quantity  int           `json:"quantity"`
}

// OrderView is the application-owned read representation.
type OrderView struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customerId"`
	EditionID  string     `json:"editionId"`
	Currency   string     `json:"currency"`
	Status     string     `json:"status"`
	Lines      []LineView `json:"lines"`
}
