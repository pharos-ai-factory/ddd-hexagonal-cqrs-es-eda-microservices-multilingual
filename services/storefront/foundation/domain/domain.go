// Package domain contains context-neutral tactical primitives.
package domain

import (
	"fmt"
	"regexp"
)

// Violation is an expected business rejection, independent of any transport.
type Violation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (v *Violation) Error() string        { return v.Code + ": " + v.Message }
func (v *Violation) Rejection() Violation { return *v }
func Reject(code, message string) error   { return &Violation{code, message} }

// ExpectedError is the only error shape a command port records as an outcome.
type ExpectedError interface {
	error
	Rejection() Violation
}

// DomainError marks a named aggregate rule, apart from generic rejections.
type DomainError interface {
	ExpectedError
	DomainError()
}

var identifier = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func ValidateID(id string) error {
	if !identifier.MatchString(id) {
		return Reject("invalid_id", "A canonical UUID is required")
	}
	return nil
}

// Fact has domain meaning only. Delivery identity and causality belong outside it.
type Fact struct {
	Name string
	Data any
}

// Quantity validates a positive item count as an immutable value object.
type Quantity struct{ value int }

func NewQuantity(value int) (Quantity, error) {
	if value < 1 || value > 5 {
		return Quantity{}, Reject("invalid_quantity", "Quantity must be between one and five")
	}
	return Quantity{value}, nil
}
func (q Quantity) Value() int { return q.value }

// Money uses integer minor units; floating-point prices are never accepted.
type Money struct {
	minor    int64
	currency string
}

func NewMoney(minor int64, currency string) (Money, error) {
	if minor < 0 || minor > 1_000_000 || !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(currency) {
		return Money{}, Reject("invalid_money", "A non-negative bounded amount and three-letter currency are required")
	}
	return Money{minor, currency}, nil
}
func (m Money) Minor() int64           { return m.minor }
func (m Money) Currency() string       { return m.currency }
func (m Money) Equal(other Money) bool { return m == other }

func Require(condition bool, code, message string) error {
	if !condition {
		return Reject(code, message)
	}
	return nil
}

// Corrupt distinguishes invalid persisted state from an expected command failure.
func Corrupt(reason string) error { return fmt.Errorf("corrupt aggregate state: %s", reason) }
