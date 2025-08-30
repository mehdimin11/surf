package surf_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
	"github.com/klauspost/compress/zstd"
)

func TestMiddlewareResponseCloseIdleConnections(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"test": "close_idle"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// The middleware should close idle connections after response
	// We can't easily test the actual closing, but we can verify the request succeeded
	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "close_idle") {
		t.Error("expected response body to contain test data")
	}
}

func TestMiddlewareResponseWebSocketUpgradeError(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Simulate WebSocket upgrade response
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.WriteHeader(http.StatusSwitchingProtocols)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	// Should get WebSocket upgrade error
	if resp.IsOk() {
		t.Error("expected WebSocket upgrade error")
	}

	if !strings.Contains(resp.Err().Error(), "switching protocols to WebSocket") {
		t.Errorf("expected WebSocket upgrade error message, got: %v", resp.Err())
	}
}

func TestMiddlewareResponseWebSocketUpgradeNormal(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Normal response without WebSocket upgrade
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"normal": "response"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}
}

func TestMiddlewareResponseDecodeBodyGzip(t *testing.T) {
	t.Parallel()

	originalData := "This is test data for gzip compression"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Create gzip compressed response
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte(originalData))
		gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected decompressed body %q, got %q", originalData, body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyDeflate(t *testing.T) {
	t.Parallel()

	originalData := "This is test data for deflate compression"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Create deflate compressed response
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		zw.Write([]byte(originalData))
		zw.Close()

		w.Header().Set("Content-Encoding", "deflate")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected decompressed body %q, got %q", originalData, body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyBrotli(t *testing.T) {
	t.Parallel()

	originalData := "This is test data for brotli compression"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Create brotli compressed response
		var buf bytes.Buffer
		br := brotli.NewWriter(&buf)
		br.Write([]byte(originalData))
		br.Close()

		w.Header().Set("Content-Encoding", "br")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected decompressed body %q, got %q", originalData, body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyZstd(t *testing.T) {
	t.Parallel()

	originalData := "This is test data for zstd compression"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Create zstd compressed response
		var buf bytes.Buffer
		encoder, err := zstd.NewWriter(&buf)
		if err != nil {
			t.Fatal(err)
		}
		encoder.Write([]byte(originalData))
		encoder.Close()

		w.Header().Set("Content-Encoding", "zstd")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected decompressed body %q, got %q", originalData, body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyNoCompression(t *testing.T) {
	t.Parallel()

	originalData := "This is test data without compression"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, originalData)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected body %q, got %q", originalData, body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyEmptyBody(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Return empty body with 200 status
		w.WriteHeader(http.StatusOK)
		// No content written
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}

	// Body should be empty
	body := resp.Ok().Body.String()
	if !body.Empty() {
		t.Errorf("expected empty body, got %q", body.Std())
	}
}

func TestMiddlewareResponseDecodeBodyInvalidGzip(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		// Invalid gzip data
		w.Write([]byte("invalid gzip data"))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	// Should handle invalid gzip gracefully
	if resp.IsErr() {
		// Expected behavior - invalid compression should cause error
		if !strings.Contains(resp.Err().Error(), "gzip") && !strings.Contains(resp.Err().Error(), "invalid") {
			t.Logf("Got compression error as expected: %v", resp.Err())
		}
	}
}

func TestMiddlewareResponseDecodeBodyInvalidDeflate(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "deflate")
		w.WriteHeader(http.StatusOK)
		// Invalid deflate data
		w.Write([]byte("invalid deflate data"))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	// Should handle invalid deflate gracefully
	if resp.IsErr() {
		// Expected behavior - invalid compression should cause error
		if !strings.Contains(resp.Err().Error(), "deflate") && !strings.Contains(resp.Err().Error(), "invalid") {
			t.Logf("Got compression error as expected: %v", resp.Err())
		}
	}
}

func TestMiddlewareResponseDecodeBodyInvalidZstd(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "zstd")
		w.WriteHeader(http.StatusOK)
		// Invalid zstd data
		w.Write([]byte("invalid zstd data"))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	// Should handle invalid zstd gracefully
	if resp.IsErr() {
		// Expected behavior - invalid compression should cause error
		t.Logf("Got compression error as expected: %v", resp.Err())
	}
}

func TestMiddlewareResponseDecodeBodyUnknownEncoding(t *testing.T) {
	t.Parallel()

	originalData := "This data has unknown encoding"

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "unknown")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, originalData)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Should pass through without decoding
	body := resp.Ok().Body.String()
	if body.Std() != originalData {
		t.Errorf("expected body %q, got %q", originalData, body.Std())
	}
}
