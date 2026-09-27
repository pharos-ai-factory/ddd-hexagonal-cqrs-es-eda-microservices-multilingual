package sessions

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
	"github.com/redis/go-redis/v9"
)

const prefix = "cafe:auth:"
const Lifetime = time.Hour

type Store struct {
	client   *redis.Client
	url, key string
	http     *http.Client
}

func Open(ctx context.Context, address, password, centrifugoURL, centrifugoKey string) (*Store, error) {
	client := redis.NewClient(&redis.Options{Addr: address, Username: "sessions", Password: password,
		DialTimeout: 3 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second})
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return &Store{client: client, url: centrifugoURL, key: centrifugoKey, http: &http.Client{Timeout: 5 * time.Second}}, nil
}
func (s *Store) Close()          { _ = s.client.Close() }
func digest(token string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(token))) }
func (s *Store) Create(ctx context.Context) (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes[:])
	key := digest(token)
	data, _ := json.Marshal(a.Principal{Subject: a.OperatorID, Name: "Café operator", Role: "operator"})
	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, prefix+"session:"+key, data, Lifetime)
		pipe.ZAdd(ctx, prefix+"expirations", redis.Z{Score: float64(time.Now().Add(Lifetime).Unix()), Member: key})
		return nil
	})
	return token, err
}
func (s *Store) Authenticate(ctx context.Context, token string) (a.Principal, bool, error) {
	if len(token) != 43 {
		return a.Principal{}, false, nil
	}
	data, err := s.client.Get(ctx, prefix+"session:"+digest(token)).Bytes()
	if err == redis.Nil {
		return a.Principal{}, false, nil
	}
	if err != nil {
		return a.Principal{}, false, err
	}
	var principal a.Principal
	if err = json.Unmarshal(data, &principal); err != nil {
		return principal, false, err
	}
	return principal, len(a.AuthorisedChannels(principal)) > 0, nil
}
func (s *Store) Revoke(ctx context.Context, token string) error {
	_, authenticated, err := s.Authenticate(ctx, token)
	if err != nil || !authenticated {
		return err
	}
	key := digest(token)
	_, err = s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, prefix+"session:"+key)
		pipe.ZRem(ctx, prefix+"expirations", key)
		// The pending set is durable until Centrifugo accepts the disconnect.
		pipe.SAdd(ctx, prefix+"disconnects", a.OperatorID+"|"+key)
		return nil
	})
	return err
}
func (s *Store) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := s.client.ZRangeByScore(ctx, prefix+"expirations", &redis.ZRangeBy{
				Min: "-inf", Max: fmt.Sprint(time.Now().Unix()), Count: 100}).Result()
			if err != nil {
				continue
			}
			for _, key := range expired {
				_, _ = s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
					pipe.Del(ctx, prefix+"session:"+key)
					pipe.ZRem(ctx, prefix+"expirations", key)
					pipe.SAdd(ctx, prefix+"disconnects", a.OperatorID+"|"+key)
					return nil
				})
			}
			actors, err := s.client.SMembers(ctx, prefix+"disconnects").Result()
			if err != nil {
				continue
			}
			for _, work := range actors {
				// Each revoked session has its own work identity. Completing an
				// older disconnect cannot remove a concurrent revocation.
				// Start the fence only after observing durable revocation. Every
				// earlier connect grant then expires before the final disconnect.
				fence := prefix + "fence:" + work
				if err := s.client.SetNX(ctx, fence, time.Now().Add(a.ConnectLifetime).Unix(), Lifetime).Err(); err != nil {
					continue
				}
				deadline, err := s.client.Get(ctx, fence).Int64()
				// Centrifugo accepts expire_at == now, so cross the whole second.
				if err != nil || time.Now().Unix() <= deadline {
					continue
				}
				actor, _, _ := strings.Cut(work, "|")
				if err := s.disconnect(ctx, actor); err == nil {
					_, _ = s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
						pipe.SRem(ctx, prefix+"disconnects", work)
						pipe.Del(ctx, fence)
						return nil
					})
				}
			}
		}
	}
}
func (s *Store) disconnect(ctx context.Context, actor string) error {
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
	defer response.Body.Close()
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
