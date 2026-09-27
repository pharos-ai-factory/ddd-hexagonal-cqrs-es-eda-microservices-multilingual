package domain

import (
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"strings"
	"unicode/utf8"
)

type DrinkState struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Revision  uint64 `json:"revision"`
	Published bool   `json:"published"`
}
type Drink struct {
	state DrinkState
	facts []core.Fact
}

func NewDrink(id, name string) (*Drink, error) {
	if err := core.ValidateID(id); err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return nil, core.Reject("invalid_drink_name", "Drink name must contain 1–80 characters")
	}
	return &Drink{state: DrinkState{ID: id, Name: name}}, nil
}
func RestoreDrink(state DrinkState) (*Drink, error) {
	if _, err := NewDrink(state.ID, state.Name); err != nil {
		return nil, core.Corrupt(err.Error())
	}
	if state.Published && state.Revision == 0 {
		return nil, core.Corrupt("published drink has no revision")
	}
	return &Drink{state: state}, nil
}
func (d *Drink) Revise(name string) error {
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return core.Reject("invalid_drink_name", "Drink name must contain 1–80 characters")
	}
	if d.state.Name == name {
		return nil
	}
	d.state.Name = name
	d.state.Published = false
	d.facts = append(d.facts, core.Fact{Name: "DrinkRevised", Data: d.state})
	return nil
}
func (d *Drink) Publish() error {
	if d.state.Published {
		return core.Reject("drink_already_published", "This drink revision is already published")
	}
	d.state.Revision++
	d.state.Published = true
	d.facts = append(d.facts, core.Fact{Name: "DrinkPublished", Data: d.state})
	return nil
}
func (d *Drink) Snapshot() DrinkState { return d.state }
func (d *Drink) Events() []core.Fact  { return append([]core.Fact(nil), d.facts...) }
