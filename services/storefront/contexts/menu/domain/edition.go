package domain

import (
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"regexp"
	"slices"
)

type OfferState struct {
	Code          string `json:"code"`
	DrinkID       string `json:"drinkId"`
	DrinkRevision uint64 `json:"drinkRevision"`
	Name          string `json:"name"`
	Minor         int64  `json:"minor"`
	Currency      string `json:"currency"`
}

// MenuOffer compares by its complete offered value, including the pinned revision.
type MenuOffer struct{ value OfferState }

func NewOffer(code string, drink DrinkState, minor int64, currency string) (MenuOffer, error) {
	if !regexp.MustCompile(`^[A-Z][A-Z0-9]{0,7}$`).MatchString(code) {
		return MenuOffer{}, core.Reject("invalid_offer_code", "Use 1–8 uppercase letters or digits")
	}
	if !drink.Published || drink.Revision == 0 {
		return MenuOffer{}, core.Reject("drink_not_published", "An offer requires a published drink revision")
	}
	if _, err := NewDrink(drink.ID, drink.Name); err != nil {
		return MenuOffer{}, err
	}
	if _, err := core.NewMoney(minor, currency); err != nil {
		return MenuOffer{}, err
	}
	return MenuOffer{OfferState{code, drink.ID, drink.Revision, drink.Name, minor, currency}}, nil
}
func (o MenuOffer) Equal(other MenuOffer) bool { return o == other }
func (o MenuOffer) Snapshot() OfferState       { return o.value }

type EditionState struct {
	ID       string       `json:"id"`
	Currency string       `json:"currency"`
	Status   string       `json:"status"`
	Offers   []OfferState `json:"offers"`
}
type MenuEdition struct {
	state EditionState
	facts []core.Fact
}

func NewEdition(id, currency string) (*MenuEdition, error) {
	if err := core.ValidateID(id); err != nil {
		return nil, err
	}
	if _, err := core.NewMoney(0, currency); err != nil {
		return nil, err
	}
	return &MenuEdition{state: EditionState{ID: id, Currency: currency, Status: "draft", Offers: []OfferState{}}}, nil
}
func RestoreEdition(state EditionState) (*MenuEdition, error) {
	if _, err := NewEdition(state.ID, state.Currency); err != nil {
		return nil, core.Corrupt(err.Error())
	}
	if state.Status != "draft" && state.Status != "published" {
		return nil, core.Corrupt("unknown edition status")
	}
	if len(state.Offers) > 20 || (state.Status == "published" && len(state.Offers) == 0) {
		return nil, core.Corrupt("invalid edition size")
	}
	seen := map[string]bool{}
	for _, offer := range state.Offers {
		drink := DrinkState{ID: offer.DrinkID, Name: offer.Name, Revision: offer.DrinkRevision, Published: true}
		if _, err := NewOffer(offer.Code, drink, offer.Minor, offer.Currency); err != nil {
			return nil, core.Corrupt(err.Error())
		}
		if offer.Currency != state.Currency || seen[offer.Code] {
			return nil, core.Corrupt("mixed currencies or duplicate offer codes")
		}
		seen[offer.Code] = true
	}
	state.Offers = slices.Clone(state.Offers)
	return &MenuEdition{state: state}, nil
}
func (e *MenuEdition) AddOffer(offer MenuOffer) error {
	if e.state.Status != "draft" {
		return core.Reject("edition_already_published", "Published menus cannot be edited")
	}
	if offer.value.Currency != e.state.Currency {
		return core.Reject("mixed_currencies", "All offers must use the edition currency")
	}
	for _, existing := range e.state.Offers {
		if existing.Code == offer.value.Code {
			return core.Reject("duplicate_offer_code", "Offer codes must be unique within the edition")
		}
	}
	if len(e.state.Offers) >= 20 {
		return core.Reject("menu_full", "The demonstration menu holds at most 20 offers")
	}
	e.state.Offers = append(e.state.Offers, offer.value)
	e.facts = append(e.facts, core.Fact{Name: "MenuOfferAdded", Data: offer.value})
	return nil
}
func (e *MenuEdition) ChangePrice(code string, minor int64) error {
	if e.state.Status != "draft" {
		return core.Reject("edition_already_published", "Published menus cannot be edited")
	}
	if _, err := core.NewMoney(minor, e.state.Currency); err != nil {
		return err
	}
	for i, offer := range e.state.Offers {
		if offer.Code == code {
			if offer.Minor == minor {
				return nil
			}
			offer.Minor = minor
			e.state.Offers[i] = offer
			e.facts = append(e.facts, core.Fact{Name: "MenuOfferChanged", Data: offer})
			return nil
		}
	}
	return core.Reject("offer_not_found", "The menu offer does not exist")
}
func (e *MenuEdition) Publish() error {
	if e.state.Status != "draft" {
		return core.Reject("edition_already_published", "The edition is already published")
	}
	if len(e.state.Offers) == 0 {
		return core.Reject("empty_menu", "A published menu must contain an offer")
	}
	e.state.Status = "published"
	e.facts = append(e.facts, core.Fact{Name: "MenuEditionPublished", Data: e.Snapshot()})
	return nil
}
func (e *MenuEdition) Snapshot() EditionState {
	s := e.state
	s.Offers = slices.Clone(s.Offers)
	return s
}
func (e *MenuEdition) Events() []core.Fact { return append([]core.Fact(nil), e.facts...) }
