package domain

import (
	"errors"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"reflect"
	"testing"
)

const id = "11111111-1111-4111-8111-111111111111"
const lineID = "22222222-2222-4222-8222-222222222222"

func assertRejected(t *testing.T, err error, code string) {
	t.Helper()
	var violation *core.Violation
	if !errors.As(err, &violation) || violation.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}
func TestOrderOwnsLineInvariants(t *testing.T) {
	order, err := New(id, id, id, "EUR")
	if err != nil {
		t.Fatal(err)
	}
	assertRejected(t, order.Place(), "empty_order")
	if err = order.AddLine(lineID, Selection{"C1", "Coffee", 300}, 4); err != nil {
		t.Fatal(err)
	}
	before := order.Snapshot()
	events := len(order.Events())
	assertRejected(t, order.ChangeQuantity(lineID, 6), "invalid_quantity")
	if !reflect.DeepEqual(before, order.Snapshot()) || len(order.Events()) != events {
		t.Fatal("rejection changed state or emitted a fact")
	}
	if err = order.ChangeQuantity(lineID, 5); err != nil {
		t.Fatal(err)
	}
	if order.Snapshot().Lines[0].ID != lineID {
		t.Fatal("quantity edit replaced the line identity")
	}
	err = order.AddLine(id, Selection{"T1", "Tea", 200}, 1)
	assertRejected(t, err, "too_many_drinks")
	var tooMany *TooManyDrinksDomainError
	if !errors.As(err, &tooMany) {
		t.Fatalf("expected the named order rule, got %v", err)
	}
	if err = order.Place(); err != nil {
		t.Fatal(err)
	}
	assertRejected(t, order.ChangeQuantity(lineID, 1), "order_already_placed")
	assertRejected(t, order.AddLine(id, Selection{"T1", "Tea", 200}, 1), "order_already_placed")
	snapshot := order.Snapshot()
	snapshot.Lines[0].Quantity = 99
	if order.Snapshot().Lines[0].Quantity != 5 {
		t.Fatal("snapshot mutated the aggregate")
	}
}
func TestNoOpProducesNoAdditionalFact(t *testing.T) {
	order, _ := New(id, id, id, "EUR")
	_ = order.AddLine(lineID, Selection{"C1", "Coffee", 300}, 1)
	count := len(order.Events())
	if err := order.ChangeQuantity(lineID, 1); err != nil {
		t.Fatal(err)
	}
	if len(order.Events()) != count {
		t.Fatal("no-op produced a fact")
	}
}
