package realtime

import (
	"net/http"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func config() Config {
	return Config{Sessions: test.Sessions{}, ProxySecret: "proxy", Origins: []string{"http://cafe.local"}}
}

func newHandler(c Config) http.Handler {
	mux := http.NewServeMux()
	c.Mount(mux)
	return mux
}

var request = test.Request
