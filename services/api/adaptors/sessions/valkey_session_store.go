package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
	"github.com/redis/go-redis/v9"
)

const prefix = "cafe:auth:"
const Lifetime = time.Hour

// ValkeySessionStore implements durable sessions and revocation work in Valkey.
type ValkeySessionStore struct {
	diagnosticsMu sync.Mutex
	failures      uint64
	lastFailure   *Failure
	client        *redis.Client
	url, key      string
	http          *http.Client
}

func Open(ctx context.Context, address, password, centrifugoURL, centrifugoKey string) (*ValkeySessionStore, error) {
	client := redis.NewClient(&redis.Options{Addr: address, Username: "sessions", Password: password,
		DialTimeout: 3 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &ValkeySessionStore{client: client, url: centrifugoURL, key: centrifugoKey, http: &http.Client{Timeout: 5 * time.Second}}, nil
}
func (s *ValkeySessionStore) Close() { _ = s.client.Close() }
func digest(token string) string     { return fmt.Sprintf("%x", sha256.Sum256([]byte(token))) }
func (s *ValkeySessionStore) Create(ctx context.Context) (string, error) {
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
func (s *ValkeySessionStore) Authenticate(ctx context.Context, token string) (a.Principal, bool, error) {
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
func (s *ValkeySessionStore) Revoke(ctx context.Context, token string) error {
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
