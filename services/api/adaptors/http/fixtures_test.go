package http

import test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"

func config() Config {
	return Config{Sessions: test.Sessions{}, OperatorPassword: "code", CLIKey: "cli",
		ProxySecret: "proxy", Origins: []string{"http://cafe.local"}}
}

var request = test.Request
