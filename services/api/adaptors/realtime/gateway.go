// Package realtime restricts technical Centrifugo operations by workload identity.
package realtime

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const MaxRequestBytes = 1024 * 1024

type Config struct {
	UpstreamURL, UpstreamKey, SessionKey string
	PublisherKeys                        map[string]string
}

// Handler rejects overlapping credentials: a credential identifies one authority.
func (c Config) Handler() (http.Handler, error) {
	u, err := url.Parse(c.UpstreamURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("invalid realtime upstream")
	}
	seen := map[string]bool{}
	keys := []string{c.UpstreamKey, c.SessionKey}
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		keys = append(keys, c.PublisherKeys[owner])
	}
	if len(c.PublisherKeys) != 6 {
		return nil, fmt.Errorf("six publisher identities required")
	}
	for _, key := range keys {
		if key == "" || seen[key] {
			return nil, fmt.Errorf("distinct non-empty realtime credentials required")
		}
		seen[key] = true
	}
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("POST /api/publish", func(w http.ResponseWriter, r *http.Request) {
		owner := ""
		for context, key := range c.PublisherKeys {
			if sameKey(r.Header.Get("X-API-Key"), key) {
				owner = context
			}
		}
		if owner == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var publication struct {
			Channel string `json:"channel"`
			Data    string `json:"b64data"`
			ID      string `json:"idempotency_key"`
		}
		body, err := decode(w, r, &publication)
		if err != nil {
			http.Error(w, "invalid publication", http.StatusBadRequest)
			return
		}
		if publication.Channel != "cafe:"+owner {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		wire, err := base64.StdEncoding.Strict().DecodeString(publication.Data)
		if err != nil || !validEnvelope(wire, owner, publication.ID) {
			http.Error(w, "invalid publication envelope", http.StatusBadRequest)
			return
		}
		c.forward(w, r, client, body)
	})
	mux.HandleFunc("POST /api/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if !sameKey(r.Header.Get("X-API-Key"), c.SessionKey) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var command struct {
			User string `json:"user"`
		}
		body, err := decode(w, r, &command)
		if err != nil || command.User == "" {
			http.Error(w, "invalid disconnect", http.StatusBadRequest)
			return
		}
		c.forward(w, r, client, body)
	})
	return mux, nil
}

func sameKey(actual, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func decode(w http.ResponseWriter, r *http.Request, destination any) ([]byte, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBytes))
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(destination); err != nil {
		return nil, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("one request required")
	}
	return body, nil
}

func (c Config) forward(w http.ResponseWriter, r *http.Request, client *http.Client, body []byte) {
	request, err := http.NewRequestWithContext(r.Context(), "POST", strings.TrimRight(c.UpstreamURL, "/")+r.URL.Path, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", c.UpstreamKey)
	request.Header.Set("X-Centrifugo-Error-Mode", "transport")
	response, err := client.Do(request)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(response.Body, MaxRequestBytes))
}
