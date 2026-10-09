// Package model contains framework-independent published integration event DTOs.
package model

// Offer is an immutable published menu offer.
type Offer struct {
	Code          string `json:"code"`
	DrinkID       string `json:"drinkId"`
	DrinkRevision uint64 `json:"drinkRevision"`
	Name          string `json:"name"`
	Minor         int64  `json:"minor"`
	Currency      string `json:"currency"`
}

// MenuPublished carries the named fact at this module's domain or integration boundary.
type MenuPublished struct {
	EditionID string  `json:"editionId"`
	Currency  string  `json:"currency"`
	Offers    []Offer `json:"offers"`
}

// Line is an immutable published order line.
type Line struct {
	ID        string `json:"id"`
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Minor     int64  `json:"minor"`
}

// OrderPlaced carries the named fact at this module's domain or integration boundary.
type OrderPlaced struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	EditionID  string `json:"editionId"`
	Currency   string `json:"currency"`
	Lines      []Line `json:"lines"`
}

// DrinksReady carries the named fact at this module's domain or integration boundary.
type DrinksReady struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
}

// PickupOpened carries the named fact at this module's domain or integration boundary.
type PickupOpened struct {
	PickupID       string `json:"pickupId"`
	OrderID        string `json:"orderId"`
	CustomerID     string `json:"customerId"`
	CollectionCode string `json:"collectionCode"`
}

// OrderCollected carries the named fact at this module's domain or integration boundary.
type OrderCollected struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
}

// RewardIssued carries the named fact at this module's domain or integration boundary.
type RewardIssued struct {
	RewardID   string `json:"rewardId"`
	CustomerID string `json:"customerId"`
	Benefit    string `json:"benefit"`
	ExpiresAt  string `json:"expiresAt"`
}
