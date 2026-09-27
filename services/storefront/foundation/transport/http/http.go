package http

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func Auth(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+key)) != 1 {
			JSON(w, 401, map[string]string{"code": "unauthorised"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func Command[I any](name string, handler func(context.Context, a.Metadata, I) (a.Outcome, error), requiredFields ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input I
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64*1024))
		raw = bytes.TrimSpace(raw)
		if err != nil || len(raw) == 0 || raw[0] != '{' {
			JSON(w, 400, map[string]string{"code": "invalid_request"})
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			JSON(w, 400, map[string]string{"code": "invalid_request", "message": "A strictly shaped JSON command body is required"})
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			JSON(w, 400, map[string]string{"code": "invalid_request"})
			return
		}
		if len(requiredFields) > 0 {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				JSON(w, 400, map[string]string{"code": "invalid_request"})
				return
			}
			for _, field := range requiredFields {
				value := bytes.TrimSpace(fields[field])
				if len(value) == 0 || bytes.Equal(value, []byte("null")) {
					JSON(w, 400, map[string]string{"code": "invalid_request", "message": field + " must be supplied and non-null"})
					return
				}
			}
		}
		id := r.PathValue("id")
		key := r.Header.Get("Idempotency-Key")
		correlation := r.Header.Get("X-Correlation-ID")
		if correlation == "" {
			correlation = key
		}
		for _, value := range []string{id, key, correlation} {
			if err := d.ValidateID(value); err != nil {
				JSON(w, 400, err)
				return
			}
		}
		rawVersion := strings.Trim(r.Header.Get("If-Match"), `"`)
		version, err := strconv.ParseUint(rawVersion, 10, 64)
		if err != nil {
			JSON(w, 428, map[string]string{"code": "expected_version_required", "message": "Supply If-Match with the expected aggregate version; use 0 for creation"})
			return
		}
		metadata := a.Metadata{ID: key, AggregateID: id, Name: name, ExpectedVersion: &version, CorrelationID: correlation, Input: input}
		outcome, err := handler(r.Context(), metadata, input)
		if err != nil {
			JSON(w, 503, map[string]string{"code": "temporarily_unavailable"})
			return
		}
		status := 200
		if outcome.Rejection != nil {
			status = 422
			if outcome.Rejection.Code == "version_conflict" || outcome.Rejection.Code == "idempotency_conflict" {
				status = 409
			}
			if outcome.Rejection.Code == "not_found" {
				status = 404
			}
		}
		JSON(w, status, outcome)
	}
}
func Get[S any](handler func(context.Context, string) (a.Loaded[S], error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := d.ValidateID(r.PathValue("id")); err != nil {
			JSON(w, 400, err)
			return
		}
		item, err := handler(r.Context(), r.PathValue("id"))
		if err != nil {
			JSON(w, 503, map[string]string{"code": "temporarily_unavailable"})
			return
		}
		if !item.Exists {
			JSON(w, 404, map[string]string{"code": "not_found"})
			return
		}
		JSON(w, 200, item)
	}
}
func List[S any](handler func(context.Context) ([]a.Loaded[S], error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := handler(r.Context())
		if err != nil {
			JSON(w, 503, map[string]string{"code": "temporarily_unavailable"})
			return
		}
		JSON(w, 200, items)
	}
}
