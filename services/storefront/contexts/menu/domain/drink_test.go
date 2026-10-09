package domain

import (
	"strings"
	"testing"
)

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
