package postgres

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"testing"
)

func TestOrderReadViewRetainsLinesAndRejectsInvalidQuantity(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	state := d.OrderState{ID: id, CustomerID: id, EditionID: id, Currency: "EUR", Status: "placed", Lines: []d.LineState{{ID: id, Quantity: 2, Selection: d.Selection{OfferCode: "C1", Name: "Coffee", Minor: 300}}}}
	view, err := orderView(state)
	if err != nil {
		t.Fatal(err)
	}
	if view.Lines[0].Quantity != 2 {
		t.Fatal("lost line quantity")
	}
	view.Lines[0].Quantity = 1
	if state.Lines[0].Quantity != 2 {
		t.Fatal("view aliases stored lines")
	}
	state.Lines[0].Quantity = 6
	if _, err = orderView(state); err == nil {
		t.Fatal("invalid stored order accepted")
	}
}
