package surf_test

import (
	"net/http"
	"testing"

	"github.com/enetx/g"
	ehttp "github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
	"github.com/enetx/surf/header"
)

func TestHeadersBasic(t *testing.T) {
	t.Parallel()

	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "test-value")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	headers := resp.Ok().Headers

	// Test header access
	if !headers.Contains("Content-Type", "application/json") {
		t.Error("expected Content-Type header to contain application/json")
	}

	if headers.Get("Content-Type").Std() != "application/json" {
		t.Errorf("expected Content-Type to be application/json, got %s", headers.Get("Content-Type"))
	}

	if headers.Get("X-Custom").Std() != "test-value" {
		t.Errorf("expected X-Custom to be test-value, got %s", headers.Get("X-Custom"))
	}
}

func TestHeaderConstants(t *testing.T) {
	t.Parallel()

	// Test that header constants are available
	if header.ACCEPT == "" {
		t.Error("expected ACCEPT header constant to be available")
	}

	if header.CONTENT_TYPE == "" {
		t.Error("expected CONTENT_TYPE header constant to be available")
	}

	if header.USER_AGENT == "" {
		t.Error("expected USER_AGENT header constant to be available")
	}

	if header.AUTHORIZATION == "" {
		t.Error("expected AUTHORIZATION header constant to be available")
	}
}

func TestHeadersIteration(t *testing.T) {
	t.Parallel()

	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.Header().Set("X-Header-1", "value1")
		w.Header().Set("X-Header-2", "value2")
		w.Header().Set("X-Header-3", "value3")
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	headers := resp.Ok().Headers
	foundHeaders := 0

	// Iterate through headers manually since Headers doesn't have ForEach
	for name := range headers {
		if g.String(name).StartsWith("X-Header-") {
			foundHeaders++
		}
	}

	if foundHeaders != 3 {
		t.Errorf("expected 3 custom headers, found %d", foundHeaders)
	}
}

func TestHeadersSetOnRequest(t *testing.T) {
	t.Parallel()

	var receivedHeaders map[string]string

	handler := func(w ehttp.ResponseWriter, r *ehttp.Request) {
		receivedHeaders = make(map[string]string)
		for name, values := range r.Header {
			if len(values) > 0 {
				receivedHeaders[name] = values[0]
			}
		}
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()

	headers := g.NewMapOrd[g.String, g.String](3)
	headers.Set("X-Custom-1", "value1")
	headers.Set("X-Custom-2", "value2")
	headers.Set(header.CONTENT_TYPE, "application/json")

	resp := client.Get(g.String(ts.URL)).SetHeaders(headers).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Check that headers were sent
	if receivedHeaders["X-Custom-1"] != "value1" {
		t.Error("expected X-Custom-1 header to be sent")
	}

	if receivedHeaders["X-Custom-2"] != "value2" {
		t.Error("expected X-Custom-2 header to be sent")
	}

	if receivedHeaders["Content-Type"] != "application/json" {
		t.Error("expected Content-Type header to be sent")
	}
}

func TestHeadersCaseInsensitive(t *testing.T) {
	t.Parallel()

	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	headers := resp.Ok().Headers

	// Test case-insensitive access
	if !headers.Contains("content-type", "application/json") {
		t.Error("expected headers to be case-insensitive")
	}

	if !headers.Contains("CONTENT-TYPE", "application/json") {
		t.Error("expected headers to be case-insensitive")
	}

	if headers.Get("content-type").Empty() {
		t.Error("expected case-insensitive header access")
	}
}

func TestHeadersEmpty(t *testing.T) {
	t.Parallel()

	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	headers := resp.Ok().Headers

	// Test accessing non-existent header
	if !headers.Get("Non-Existent-Header").Empty() {
		t.Error("expected non-existent header to return empty string")
	}

	if headers.Contains("Non-Existent-Header", "any") {
		t.Error("expected Contains to return false for non-existent header")
	}
}

func TestHeadersMultipleValues(t *testing.T) {
	t.Parallel()

	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.Header().Add("X-Multi", "value1")
		w.Header().Add("X-Multi", "value2")
		w.Header().Add("X-Multi", "value3")
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	headers := resp.Ok().Headers

	// Test that multi-value headers are handled
	if !headers.Contains("X-Multi", "value1") {
		t.Error("expected X-Multi header to contain value1")
	}

	multiValue := headers.Get("X-Multi")
	if multiValue.Empty() {
		t.Error("expected X-Multi header to have value")
	}
}
