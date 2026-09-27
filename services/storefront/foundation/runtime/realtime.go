package runtime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	"net/http"
	"time"
)

func Realtime(ctx context.Context, db *postgres.Database, url, key string) {
	client := &http.Client{Timeout: 5 * time.Second}
	for ctx.Err() == nil {
		dispatch, found, err := db.ClaimRealtime(ctx)
		if err != nil || !found {
			if !Wait(ctx, 100*time.Millisecond) {
				return
			}
			continue
		}
		err = publishRealtime(ctx, client, url, key, dispatch)
		_ = db.FinishRealtime(ctx, dispatch, err)
		if err != nil && !Wait(ctx, time.Second) {
			return
		}
	}
}
func publishRealtime(ctx context.Context, client *http.Client, url, key string, d postgres.RealtimeDispatch) error {
	body, _ := json.Marshal(map[string]string{"channel": d.Channel, "b64data": base64.StdEncoding.EncodeToString(d.Body), "idempotency_key": d.ID})
	request, err := http.NewRequestWithContext(ctx, "POST", url+"/api/publish", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", key)
	request.Header.Set("X-Centrifugo-Error-Mode", "transport")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var result struct {
		Error json.RawMessage `json:"error"`
	}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if response.StatusCode != 200 || len(result.Error) > 0 {
		return fmt.Errorf("realtime publication not accepted")
	}
	return nil
}
