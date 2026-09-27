// The development provider accepts messages into a durable local mailbox.
// It never contacts an email service or sends a message to another person.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Accepted struct {
	Receipt     string     `json:"receipt"`
	Fingerprint string     `json:"fingerprint"`
	Message     a.Delivery `json:"message"`
}

func main() {
	if os.Getenv("APP_ENV") != "development" || len(os.Getenv("API_KEY")) < 32 {
		log.Fatal("development configuration is required")
	}
	root := os.Getenv("DATA_DIR")
	if root == "" {
		root = "/data"
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		log.Fatal(err)
	}
	var lock sync.Mutex
	failures := 0
	lostResponses := 0
	mux := http.NewServeMux()
	mux.HandleFunc("POST /messages", func(w http.ResponseWriter, r *http.Request) {
		var input a.Delivery
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			web.JSON(w, 400, map[string]string{"code": "invalid_request"})
			return
		}
		if err := d.ValidateID(input.ID); err != nil || r.Header.Get("Idempotency-Key") != input.ID {
			web.JSON(w, 400, map[string]string{"code": "invalid_id"})
			return
		}
		lock.Lock()
		defer lock.Unlock()
		data, _ := json.Marshal(input)
		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])
		path := filepath.Join(root, input.ID+".json")
		if existing, err := os.ReadFile(path); err == nil {
			var accepted Accepted
			if json.Unmarshal(existing, &accepted) != nil {
				web.JSON(w, 500, map[string]string{"code": "corrupt_receipt"})
				return
			}
			if accepted.Fingerprint != hash {
				web.JSON(w, 409, map[string]string{"code": "identity_conflict"})
				return
			}
			web.JSON(w, 200, accepted)
			return
		}
		if failures > 0 {
			failures--
			web.JSON(w, 503, map[string]string{"code": "simulated_outage"})
			return
		}
		accepted := Accepted{Receipt: "mailbox-" + input.ID, Fingerprint: hash, Message: input}
		encoded, _ := json.Marshal(accepted)
		if err := persist(root, path, encoded); err != nil {
			web.JSON(w, 500, map[string]string{"code": "storage_unavailable"})
			return
		}
		if lostResponses > 0 {
			lostResponses--
			web.JSON(w, 503, map[string]string{"code": "simulated_lost_acceptance_response"})
			return
		}
		web.JSON(w, 200, accepted)
	})
	mux.HandleFunc("POST /failures", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Count         int `json:"count"`
			LostResponses int `json:"lostResponses"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			web.JSON(w, 400, map[string]string{"code": "invalid_request"})
			return
		}
		lock.Lock()
		failures = input.Count
		lostResponses = input.LostResponses
		lock.Unlock()
		web.JSON(w, 200, input)
	})
	mux.HandleFunc("GET /messages", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		paths, err := filepath.Glob(filepath.Join(root, "*.json"))
		if err != nil {
			web.JSON(w, 500, map[string]string{"code": "storage_unavailable"})
			return
		}
		items := []Accepted{}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var item Accepted
			if json.Unmarshal(data, &item) == nil {
				items = append(items, item)
			}
		}
		web.JSON(w, 200, items)
	})
	server := http.Server{Addr: ":8080", Handler: web.Auth(os.Getenv("API_KEY"), mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	log.Fatal(server.ListenAndServe())
}
func persist(root, path string, data []byte) error {
	file, err := os.CreateTemp(root, "pending-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return err
	}
	directory, err := os.Open(root)
	if err != nil {
		return err
	}
	defer directory.Close()
	if err = directory.Sync(); err != nil {
		return fmt.Errorf("sync mailbox directory: %w", err)
	}
	return nil
}
