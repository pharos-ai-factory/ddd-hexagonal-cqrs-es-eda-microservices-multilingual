package operational

import (
	"net/http"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func config() Config { return Config{CLIKey: "cli"} }

func newHandler(c Config) http.Handler {
	mux := http.NewServeMux()
	c.Mount(mux)
	return mux
}

var request = test.Request
