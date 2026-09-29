package application

import core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"

// OrderNotFoundApplicationError is an expected command-target failure.
type OrderNotFoundApplicationError struct{ core.ApplicationError }

func orderNotFound() error {
	return &OrderNotFoundApplicationError{ApplicationError: core.ApplicationError{
		Code: "not_found", Message: "The order does not exist",
	}}
}
