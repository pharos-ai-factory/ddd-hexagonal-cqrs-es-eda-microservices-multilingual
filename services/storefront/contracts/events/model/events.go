// Package model contains framework-independent published and private event DTOs.
package model

type DrinkPublished struct {
	DrinkID  string `json:"drinkId"`
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
}

type Offer struct {
	Code          string `json:"code"`
	DrinkID       string `json:"drinkId"`
	DrinkRevision uint64 `json:"drinkRevision"`
	Name          string `json:"name"`
	Minor         int64  `json:"minor"`
	Currency      string `json:"currency"`
}

type MenuPublished struct {
	EditionID string  `json:"editionId"`
	Currency  string  `json:"currency"`
	Offers    []Offer `json:"offers"`
}

type Line struct {
	ID        string `json:"id"`
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Minor     int64  `json:"minor"`
}

type OrderPlaced struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	EditionID  string `json:"editionId"`
	Currency   string `json:"currency"`
	Lines      []Line `json:"lines"`
}

type DrinksReady struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
}

type PickupOpened struct {
	PickupID       string `json:"pickupId"`
	OrderID        string `json:"orderId"`
	CustomerID     string `json:"customerId"`
	CollectionCode string `json:"collectionCode"`
}

type OrderCollected struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
}

// RewardEarned is private to Loyalty despite using the common durable transport.
type RewardEarned struct {
	AccountID string `json:"accountId"`
	GrantID   string `json:"grantId"`
	Benefit   string `json:"benefit"`
	ValidDays int    `json:"validDays"`
}

type RewardIssued struct {
	RewardID   string `json:"rewardId"`
	CustomerID string `json:"customerId"`
	Benefit    string `json:"benefit"`
	ExpiresAt  string `json:"expiresAt"`
}

type NotificationRequested struct {
	NotificationID string `json:"notificationId"`
}
