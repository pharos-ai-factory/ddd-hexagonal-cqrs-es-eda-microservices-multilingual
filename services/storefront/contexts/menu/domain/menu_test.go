package domain

import (
	"errors"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"reflect"
	"strings"
	"testing"
)

const id = "11111111-1111-4111-8111-111111111111"

func rejection(t *testing.T, err error, code string) {
	t.Helper()
	var violation *core.Violation
	if !errors.As(err, &violation) || violation.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}
func TestEditionOwnsOfferedSnapshot(t *testing.T) {
	drink, _ := NewDrink(id, "Coffee")
	_, err := NewOffer("C1", drink.Snapshot(), 300, "EUR")
	rejection(t, err, "drink_not_published")
	_ = drink.Publish()
	offer, err := NewOffer("C1", drink.Snapshot(), 300, "EUR")
	if err != nil {
		t.Fatal(err)
	}
	edition, _ := NewEdition(id, "EUR")
	rejection(t, edition.Publish(), "empty_menu")
	if err = edition.AddOffer(offer); err != nil {
		t.Fatal(err)
	}
	before := edition.Snapshot()
	count := len(edition.Events())
	rejection(t, edition.AddOffer(offer), "duplicate_offer_code")
	foreign, _ := NewOffer("T1", drink.Snapshot(), 200, "GBP")
	rejection(t, edition.AddOffer(foreign), "mixed_currencies")
	if !reflect.DeepEqual(before, edition.Snapshot()) || len(edition.Events()) != count {
		t.Fatal("rejection changed the edition")
	}
	_ = drink.Revise("Renamed coffee")
	_ = drink.Publish()
	if edition.Snapshot().Offers[0].Name != "Coffee" || edition.Snapshot().Offers[0].DrinkRevision != 1 {
		t.Fatal("drink revision changed the pinned offer")
	}
	if err = edition.Publish(); err != nil {
		t.Fatal(err)
	}
	rejection(t, edition.ChangePrice("C1", 400), "edition_already_published")
	snapshot := edition.Snapshot()
	snapshot.Offers[0].Minor = 1
	if edition.Snapshot().Offers[0].Minor != 300 {
		t.Fatal("published snapshot can mutate the edition")
	}
}
func TestOfferValueEquality(t *testing.T) {
	drink, _ := NewDrink(id, "Coffee")
	_ = drink.Publish()
	a, _ := NewOffer("C1", drink.Snapshot(), 300, "EUR")
	b, _ := NewOffer("C1", drink.Snapshot(), 300, "EUR")
	c, _ := NewOffer("C1", drink.Snapshot(), 350, "EUR")
	if !a.Equal(b) || a.Equal(c) {
		t.Fatal("offer equality does not follow its value")
	}
}

func TestDrinkNameLimitCountsCharactersRatherThanUTF8Bytes(t *testing.T) {
	name := strings.Repeat("ή", 80)
	t.Run("create", func(t *testing.T) {
		drink, err := NewDrink(id, name)
		if err != nil {
			t.Fatalf("an 80-character drink name must be accepted: %v", err)
		}
		if drink.Snapshot().Name != name {
			t.Fatal("the name changed")
		}
	})
	t.Run("revise", func(t *testing.T) {
		drink, err := NewDrink(id, "Coffee")
		if err != nil {
			t.Fatal(err)
		}
		if err = drink.Revise(name); err != nil {
			t.Fatalf("the same character limit must apply to revision: %v", err)
		}
		rejection(t, drink.Revise(name+"ή"), "invalid_drink_name")
	})
}
