package gateway

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressMiddlewareGzipsResponseWhenAccepted(t *testing.T) {
	body := "gateway response body"
	handler := CompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Fatalf("expected Vary: Accept-Encoding, got %q", got)
	}

	gz, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("expected gzipped response body: %v", err)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("failed to read gzipped response body: %v", err)
	}
	if string(decompressed) != body {
		t.Fatalf("expected decompressed body %q, got %q", body, string(decompressed))
	}
}

func TestCompressMiddlewareLeavesResponseUncompressedWithoutGzip(t *testing.T) {
	body := "plain gateway response"
	handler := CompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("expected no content encoding, got %q", got)
	}
	if got := rec.Body.String(); got != body {
		t.Fatalf("expected body %q, got %q", body, got)
	}
}

func TestAcceptsGzipRejectsExplicitZeroQuality(t *testing.T) {
	for _, acceptEncoding := range []string{"br, gzip;q=0", "gzip; q=0.0"} {
		if acceptsGzip(acceptEncoding) {
			t.Fatalf("expected %q to be rejected", acceptEncoding)
		}
	}
}
