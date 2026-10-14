package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *ValkeySessionStore) disconnect(ctx context.Context, actor string) error {
	body, _ := json.Marshal(map[string]string{"user": actor})
	req, err := http.NewRequestWithContext(ctx, "POST", s.url+"/api/disconnect", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.key)
	req.Header.Set("X-Centrifugo-Error-Mode", "transport")
	response, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	var result struct {
		Error json.RawMessage `json:"error"`
	}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if response.StatusCode != 200 || len(result.Error) > 0 {
		return fmt.Errorf("disconnect not accepted")
	}
	return nil
}
