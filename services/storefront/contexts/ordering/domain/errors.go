package domain

import core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"

// TooManyDrinksDomainError names the Order aggregate's five-drink invariant.
type TooManyDrinksDomainError struct{ *core.Violation }

var _ core.DomainError = (*TooManyDrinksDomainError)(nil)

func tooManyDrinks() error {
	return &TooManyDrinksDomainError{Violation: &core.Violation{
		Code: "too_many_drinks", Message: "An order contains at most five drinks",
	}}
}

func (e *TooManyDrinksDomainError) Unwrap() error { return e.Violation }
func (e *TooManyDrinksDomainError) DomainError()  {}
