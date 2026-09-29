package http

import (
	"context"
	"encoding/base64"
	"errors"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func pageRequest(values url.Values, resource string) (*a.PageRequest, error) {
	if !values.Has("limit") && !values.Has("cursor") {
		return nil, nil
	}
	invalid := errors.New("invalid pagination")
	if len(values["limit"]) != 1 || len(values["cursor"]) > 1 {
		return nil, invalid
	}
	raw := values.Get("limit")
	limit, err := strconv.Atoi(raw)
	if err != nil || strconv.Itoa(limit) != raw || limit < 1 || limit > 100 {
		return nil, invalid
	}
	request := &a.PageRequest{Limit: limit}
	if values.Has("cursor") {
		cursor := values.Get("cursor")
		if len(cursor) > 1024 {
			return nil, invalid
		}
		decoded, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != cursor {
			return nil, invalid
		}
		prefix := "1|" + resource + "|"
		if !strings.HasPrefix(string(decoded), prefix) {
			return nil, invalid
		}
		request.After = strings.TrimPrefix(string(decoded), prefix)
		if d.ValidateID(request.After) != nil {
			return nil, invalid
		}
	}
	return request, nil
}

// PagedList retains the complete array response unless pagination is requested.
func PagedList[S any](list func(context.Context) ([]a.Loaded[S], error), page func(context.Context, a.PageRequest) (a.Page[S], error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := pageRequest(r.URL.Query(), r.URL.Path)
		if err != nil {
			JSON(w, 400, map[string]string{"code": "invalid_pagination"})
			return
		}
		if request == nil {
			List(list)(w, r)
			return
		}
		result, err := page(r.Context(), *request)
		if err != nil {
			JSON(w, 503, map[string]string{"code": "temporarily_unavailable"})
			return
		}
		var cursor *string
		if result.NextID != "" {
			value := base64.RawURLEncoding.EncodeToString([]byte("1|" + r.URL.Path + "|" + result.NextID))
			cursor = &value
		}
		JSON(w, 200, struct {
			Items      []a.Loaded[S] `json:"items"`
			NextCursor *string       `json:"nextCursor"`
		}{result.Items, cursor})
	}
}
