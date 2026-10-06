// Package httpcontract checks actual HTTP exchanges against the OpenAPI document.
package httpcontract

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

func Request(t *testing.T, document *openapi3.T, request *http.Request) {
	t.Helper()
	if err := RequestError(document, request); err != nil {
		t.Fatal(err)
	}
}

func Response(t *testing.T, document *openapi3.T, request *http.Request, response *http.Response) {
	t.Helper()
	defer response.Body.Close()
	if err := ResponseError(document, request, response); err != nil {
		t.Fatal(err)
	}
}

func ResponseError(document *openapi3.T, request *http.Request, response *http.Response) error {
	router, err := legacy.NewRouter(document)
	if err != nil {
		return err
	}
	route, params, err := router.FindRoute(request)
	if err != nil {
		return err
	}
	return openapi3filter.ValidateResponse(request.Context(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request, Route: route, PathParams: params},
		Status:                 response.StatusCode, Header: response.Header, Body: response.Body,
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
	})
}

func RequestError(document *openapi3.T, request *http.Request) error {
	router, err := legacy.NewRouter(document)
	if err != nil {
		return err
	}
	route, params, err := router.FindRoute(request)
	if err != nil {
		return err
	}
	return openapi3filter.ValidateRequest(request.Context(), &openapi3filter.RequestValidationInput{
		Request: request, Route: route, PathParams: params,
		Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	})
}
