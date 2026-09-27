package domain

import (
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

type Selection struct {
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Minor     int64  `json:"minor"`
}
type LineState struct {
	ID        string    `json:"id"`
	Selection Selection `json:"selection"`
	Quantity  int       `json:"quantity"`
}

// OrderLine retains identity when quantity changes; Order controls its mutation.
type OrderLine struct {
	id        string
	selection Selection
	quantity  core.Quantity
	price     core.Money
}

func newLine(id string, selection Selection, quantity int, currency string) (OrderLine, error) {
	if err := core.ValidateID(id); err != nil {
		return OrderLine{}, err
	}
	q, err := core.NewQuantity(quantity)
	if err != nil {
		return OrderLine{}, err
	}
	price, err := core.NewMoney(selection.Minor, currency)
	if err != nil {
		return OrderLine{}, err
	}
	if selection.OfferCode == "" || selection.Name == "" {
		return OrderLine{}, core.Reject("invalid_selection", "A published selection is required")
	}
	return OrderLine{id: id, selection: selection, quantity: q, price: price}, nil
}
func (l OrderLine) SameIdentity(other OrderLine) bool { return l.id == other.id }
func (l OrderLine) Snapshot() LineState {
	return LineState{ID: l.id, Selection: l.selection, Quantity: l.quantity.Value()}
}

type State struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customerId"`
	EditionID  string      `json:"editionId"`
	Currency   string      `json:"currency"`
	Status     string      `json:"status"`
	Lines      []LineState `json:"lines"`
}
type Order struct {
	state State
	lines []OrderLine
	facts []core.Fact
}

func New(id, customer, edition, currency string) (*Order, error) {
	for _, value := range []string{id, customer, edition} {
		if err := core.ValidateID(value); err != nil {
			return nil, err
		}
	}
	if _, err := core.NewMoney(0, currency); err != nil {
		return nil, err
	}
	return &Order{state: State{ID: id, CustomerID: customer, EditionID: edition, Currency: currency, Status: "draft", Lines: []LineState{}}}, nil
}
func Restore(state State) (*Order, error) {
	order, err := New(state.ID, state.CustomerID, state.EditionID, state.Currency)
	if err != nil {
		return nil, core.Corrupt(err.Error())
	}
	if state.Status != "draft" && state.Status != "placed" {
		return nil, core.Corrupt("unknown order status")
	}
	seen := map[string]bool{}
	for _, snapshot := range state.Lines {
		line, err := newLine(snapshot.ID, snapshot.Selection, snapshot.Quantity, state.Currency)
		if err != nil {
			return nil, core.Corrupt(err.Error())
		}
		if seen[line.id] {
			return nil, core.Corrupt("duplicate line identity")
		}
		seen[line.id] = true
		order.lines = append(order.lines, line)
	}
	if order.total() > 5 || (state.Status == "placed" && order.total() == 0) {
		return nil, core.Corrupt("invalid order quantity")
	}
	order.state.Status = state.Status
	return order, nil
}
func (o *Order) AddLine(id string, selection Selection, quantity int) error {
	if o.state.Status != "draft" {
		return core.Reject("order_already_placed", "Placed orders cannot be edited")
	}
	line, err := newLine(id, selection, quantity, o.state.Currency)
	if err != nil {
		return err
	}
	for _, line := range o.lines {
		if line.id == id {
			return core.Reject("duplicate_line", "An order line already has this identity")
		}
	}
	if o.total()+line.quantity.Value() > 5 {
		return core.Reject("too_many_drinks", "An order contains at most five drinks")
	}
	o.lines = append(o.lines, line)
	o.facts = append(o.facts, core.Fact{Name: "OrderLineAdded", Data: line.Snapshot()})
	return nil
}
func (o *Order) ChangeQuantity(id string, quantity int) error {
	if o.state.Status != "draft" {
		return core.Reject("order_already_placed", "Placed orders cannot be edited")
	}
	q, err := core.NewQuantity(quantity)
	if err != nil {
		return err
	}
	for i, line := range o.lines {
		if line.id == id {
			if o.total()-line.quantity.Value()+q.Value() > 5 {
				return core.Reject("too_many_drinks", "An order contains at most five drinks")
			}
			if line.quantity.Value() == q.Value() {
				return nil
			}
			line.quantity = q
			o.lines[i] = line
			o.facts = append(o.facts, core.Fact{Name: "OrderLineQuantityChanged", Data: line.Snapshot()})
			return nil
		}
	}
	return core.Reject("line_not_found", "The order line does not exist")
}
func (o *Order) Place() error {
	if o.state.Status != "draft" {
		return core.Reject("order_already_placed", "The order is already placed")
	}
	if o.total() < 1 {
		return core.Reject("empty_order", "An order must contain at least one drink")
	}
	o.state.Status = "placed"
	o.facts = append(o.facts, core.Fact{Name: "OrderPlaced", Data: o.Snapshot()})
	return nil
}
func (o *Order) total() int {
	total := 0
	for _, line := range o.lines {
		total += line.quantity.Value()
	}
	return total
}
func (o *Order) Snapshot() State {
	s := o.state
	s.Lines = make([]LineState, 0, len(o.lines))
	for _, line := range o.lines {
		s.Lines = append(s.Lines, line.Snapshot())
	}
	return s
}
func (o *Order) Events() []core.Fact { return append([]core.Fact(nil), o.facts...) }
