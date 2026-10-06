// Validate the live owner's wire responses through the Go API after the journey.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
)

type client struct {
	base, key string
	router    routers.Router
	http      *http.Client
}

func identity() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	value[6] = value[6]&15 | 64
	value[8] = value[8]&63 | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[:4], value[4:6], value[6:8], value[8:10], value[10:])
}

func (c client) exchange(method, path string, body []byte, key, version string, validRequest bool, want ...int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", "Bearer "+c.key)
	r.Header.Set("Content-Type", "application/json")
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
		r.Header.Set("X-Correlation-ID", key)
	}
	if version != "" {
		r.Header.Set("If-Match", version)
	}
	route, params, err := c.router.FindRoute(r)
	if err != nil {
		return nil, err
	}
	input := &openapi3filter.RequestValidationInput{Request: r, Route: route, PathParams: params, Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}}
	if validRequest {
		if err := openapi3filter.ValidateRequest(ctx, input); err != nil {
			return nil, err
		}
	}
	response, err := c.http.Do(r)
	if err != nil {
		return nil, fmt.Errorf("%s %s unavailable", method, path)
	}
	defer response.Body.Close()
	allowed := false
	for _, status := range want {
		if response.StatusCode == status {
			allowed = true
		}
	}
	if !allowed {
		return nil, fmt.Errorf("%s %s status=%d", method, path, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, err
	}
	err = openapi3filter.ValidateResponse(ctx, &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.StatusCode, Header: response.Header, Body: io.NopCloser(bytes.NewReader(data)), Options: &openapi3filter.Options{IncludeResponseStatus: true}})
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	return data, nil
}

func run() error {
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		return fmt.Errorf("live HTTP conformance requires a disposable project")
	}
	base, key := os.Getenv("CAFE_HTTP_API_URL"), os.Getenv("CAFE_HTTP_API_KEY")
	if base == "" || key == "" {
		return fmt.Errorf("HTTP conformance configuration required")
	}
	document := contract.Document("api")
	router, err := legacy.NewRouter(document)
	if err != nil {
		return err
	}
	c := client{base, key, router, &http.Client{Timeout: 10 * time.Second}}
	operations := 0
	for path, item := range document.Paths.Map() {
		if item.Extensions["x-owner"] == nil {
			continue
		}
		for method, operation := range item.Operations() {
			operations++
			if method == "POST" {
				body, err := json.Marshal(operation.RequestBody.Value.Content["application/json"].Example)
				if err != nil {
					return err
				}
				target := strings.ReplaceAll(path, "{id}", identity())
				command := identity()
				if _, err = c.exchange(method, target, body, command, "0", true, 200, 404, 422); err != nil {
					return err
				}
				if _, err = c.exchange(method, target, body, command, "", false, 428); err != nil {
					return err
				}
				if operation.OperationID == "createDrink" {
					if _, err = c.exchange(method, target, []byte(`{"name":"Changed"}`), command, "0", true, 409); err != nil {
						return err
					}
				}
				continue
			}
			if strings.Contains(path, "{id}") {
				if _, err = c.exchange(method, strings.ReplaceAll(path, "{id}", identity()), nil, "", "", true, 404); err != nil {
					return err
				}
				continue
			}
			data, err := c.exchange(method, path, nil, "", "", true, 200)
			if err != nil {
				return err
			}
			var items []struct{ State struct{ ID string } }
			if err = json.Unmarshal(data, &items); err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("journey supplied no samples for %s", path)
			}
			for _, item := range items {
				if _, err = c.exchange("GET", path+"/"+item.State.ID, nil, "", "", true, 200); err != nil {
					return err
				}
			}
			cursor := ""
			seen := map[string]bool{}
			for {
				query := path + "?limit=1"
				if cursor != "" {
					query += "&cursor=" + cursor
				}
				data, err = c.exchange("GET", query, nil, "", "", true, 200)
				if err != nil {
					return err
				}
				var page struct {
					Items      []struct{ State struct{ ID string } }
					NextCursor *string
				}
				if err = json.Unmarshal(data, &page); err != nil {
					return err
				}
				for _, item := range page.Items {
					if seen[item.State.ID] {
						return fmt.Errorf("duplicate HTTP page item")
					}
					seen[item.State.ID] = true
				}
				if page.NextCursor == nil {
					break
				}
				if *page.NextCursor == cursor {
					return fmt.Errorf("HTTP cursor did not advance")
				}
				cursor = *page.NextCursor
			}
			if len(seen) != len(items) {
				return fmt.Errorf("HTTP pagination omitted items for %s", path)
			}
			if _, err = c.exchange("GET", path+"?limit=0", nil, "", "", false, 400); err != nil {
				return err
			}
		}
	}
	fmt.Printf("OpenAPI conformance passed for %d live business operations across all six contexts\n", operations)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
