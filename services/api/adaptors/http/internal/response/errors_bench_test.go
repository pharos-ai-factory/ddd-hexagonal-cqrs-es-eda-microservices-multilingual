package response

import (
	"net/http"
	"testing"
)

// Discard body bytes to measure response encoding without a growing buffer.
type benchmarkResponseWriter struct {
	header http.Header
	status int
	bytes  int
}

func (w *benchmarkResponseWriter) Header() http.Header    { return w.header }
func (w *benchmarkResponseWriter) WriteHeader(status int) { w.status = status }
func (w *benchmarkResponseWriter) Write(body []byte) (int, error) {
	w.bytes += len(body)
	return len(body), nil
}

func BenchmarkAPIErrorResponseEncoding(b *testing.B) {
	b.Run("Map", func(b *testing.B) {
		w := &benchmarkResponseWriter{header: make(http.Header)}
		b.ReportAllocs()
		for b.Loop() {
			JSON(w, http.StatusServiceUnavailable, map[string]string{"code": "session_unavailable"})
		}
	})
	b.Run("Struct", func(b *testing.B) {
		w := &benchmarkResponseWriter{header: make(http.Header)}
		b.ReportAllocs()
		for b.Loop() {
			SessionUnavailable.Write(w)
		}
	})
}
