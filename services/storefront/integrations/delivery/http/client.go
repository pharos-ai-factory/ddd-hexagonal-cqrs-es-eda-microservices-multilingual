package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"io"
	"net/http"
	"time"
)

// HTTPNotificationDeliveryClient implements the outbound request port for this named adaptor package.
type HTTPNotificationDeliveryClient struct {
	URL  string
	Key  string
	HTTP *http.Client
}

func New(url, key string) *HTTPNotificationDeliveryClient {
	return &HTTPNotificationDeliveryClient{URL: url, Key: key, HTTP: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (c *HTTPNotificationDeliveryClient) Deliver(ctx context.Context, delivery a.Delivery) (string, error) {
	body, err := json.Marshal(delivery)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", delivery.ID)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("delivery provider returned HTTP %d", res.StatusCode)
	}
	var result struct {
		Receipt string `json:"receipt"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&result); err != nil {
		return "", err
	}
	if result.Receipt == "" {
		return "", fmt.Errorf("provider acceptance has no receipt")
	}
	return result.Receipt, nil
}
