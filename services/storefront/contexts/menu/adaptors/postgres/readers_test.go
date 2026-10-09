package postgres

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"strings"
	"testing"
)

func TestReadersValidateStoredAuthorityAndIsolateOffers(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	_, err := drinkView(d.DrinkState{ID: id, Name: "Coffee", Published: true})
	if err == nil || !strings.Contains(err.Error(), "corrupt aggregate state") {
		t.Fatalf("published drink with zero revision accepted: %v", err)
	}
	state := d.EditionState{ID: id, Currency: "EUR", Status: "published", Offers: []d.OfferState{{Code: "C1", DrinkID: id, DrinkRevision: 1, Name: "Coffee", Minor: 300, Currency: "EUR"}}}
	view, err := editionView(state)
	if err != nil {
		t.Fatal(err)
	}
	if view.Offers[0].Minor != 300 {
		t.Fatal("lost offered price")
	}
	view.Offers[0].Minor = 1
	if state.Offers[0].Minor != 300 {
		t.Fatal("view aliases stored offers")
	}
	state.Offers[0].Currency = "GBP"
	if _, err = editionView(state); err == nil || !strings.Contains(err.Error(), "corrupt aggregate state") {
		t.Fatalf("mixed currencies accepted: %v", err)
	}
}
