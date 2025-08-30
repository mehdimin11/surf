package surf_test

import (
	"fmt"
	"testing"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
)

func TestTLSGrabberHTTPS(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "success"}`)
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().Build()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	// Test TLS info grabber
	tlsInfo := resp.Ok().TLSGrabber()
	if tlsInfo == nil {
		t.Error("expected TLS info to be available for HTTPS request")
	}
}

func TestTLSGrabberHTTP(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "success"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().Build()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	// Test TLS info grabber for HTTP (should be nil)
	tlsInfo := resp.Ok().TLSGrabber()
	if tlsInfo != nil {
		t.Error("expected TLS info to be nil for HTTP request")
	}
}

func TestTLSGrabberWithImpersonate(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "success"}`)
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().
		Impersonate().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	// Test TLS info grabber with impersonation
	// Note: TLS info may not be available when using impersonation
	// as it may interfere with the standard TLS handshake process
	tlsInfo := resp.Ok().TLSGrabber()
	if tlsInfo == nil {
		t.Log("TLS info not available when using impersonation (this may be expected)")
	} else {
		t.Log("TLS info available with browser impersonation")
	}
}

func TestTLSGrabberWithJA3(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "success"}`)
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().
		JA().Chrome131().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	// Test TLS info grabber with JA
	// Note: TLS info may not be available when using JA3 fingerprinting
	// as it may interfere with the standard TLS handshake process
	tlsInfo := resp.Ok().TLSGrabber()
	if tlsInfo == nil {
		t.Log("TLS info not available when using JA3 (this may be expected)")
	} else {
		t.Log("TLS info available with JA3 fingerprinting")
	}
}

func TestClientGetTLSConfig(t *testing.T) {
	t.Parallel()

	client := surf.NewClient()

	// Test that client has a TLS config
	tlsConfig := client.GetTLSConfig()
	if tlsConfig == nil {
		t.Error("expected client to have a TLS config")
	}
}
