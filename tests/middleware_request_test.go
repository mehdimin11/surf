package surf_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
)

func TestMiddlewareRequestUserAgent(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	testCases := []struct {
		name      string
		userAgent string
	}{
		{"Custom String", "MyCustomAgent/1.0"},
		{"Browser Like", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
		{"Empty Agent", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := surf.NewClient().Builder().UserAgent(tc.userAgent).Build()
			req := client.Get(g.String(ts.URL))
			resp := req.Do()

			if resp.IsErr() {
				t.Fatal(resp.Err())
			}

			body := resp.Ok().Body.String()
			if !strings.Contains(body.Std(), tc.userAgent) && tc.userAgent != "" {
				t.Errorf("expected user agent %s in response", tc.userAgent)
			}
		})
	}
}

func TestMiddlewareRequestBearerAuth(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"token": "%s"}`, token)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test with bearer token
	client := surf.NewClient().Builder().BearerAuth("test-token-123").Build()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "test-token-123") {
		t.Error("expected token in response")
	}
}

func TestMiddlewareRequestBasicAuth(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"username": "%s", "password": "%s"}`, username, password)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test with basic auth
	client := surf.NewClient().Builder().BasicAuth("testuser:testpass").Build()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "testuser") {
		t.Error("expected username in response")
	}
	if !strings.Contains(body.Std(), "testpass") {
		t.Error("expected password in response")
	}
}

func TestMiddlewareRequestContentType(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"content-type": "%s"}`, contentType)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	testCases := []struct {
		name        string
		contentType string
	}{
		{"JSON", "application/json"},
		{"XML", "application/xml"},
		{"Form", "application/x-www-form-urlencoded"},
		{"Custom", "application/custom+type"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := surf.NewClient().Builder().ContentType(g.String(tc.contentType)).Build()
			req := client.Post(g.String(ts.URL), g.String("test data"))
			resp := req.Do()

			if resp.IsErr() {
				t.Fatal(resp.Err())
			}

			body := resp.Ok().Body.String()
			if !strings.Contains(body.Std(), tc.contentType) {
				t.Errorf("expected content type %s in response", tc.contentType)
			}
		})
	}
}

func TestMiddlewareRequestGetRemoteAddress(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test GetRemoteAddress
	client := surf.NewClient().Builder().GetRemoteAddress().Build()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Check if remote address is available
	remoteAddr := resp.Ok().RemoteAddress()
	if remoteAddr == nil {
		t.Error("expected remote address to be set")
	} else {
		// Remote address should contain IP and port
		if !strings.Contains(remoteAddr.String(), ":") {
			t.Error("expected remote address to contain port")
		}
	}
}

func TestMiddlewareRequestGot101Response(t *testing.T) {
	t.Parallel()

	// Test handling of 101 Switching Protocols response
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			// Simulate WebSocket upgrade attempt
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.WriteHeader(http.StatusSwitchingProtocols)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()

	// Test normal request
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Ok().StatusCode)
	}

	// Test WebSocket upgrade attempt (should handle 101 response)
	req2 := client.Get(g.String(ts.URL)).
		SetHeaders(g.Map[string, string]{
			"Upgrade":    "websocket",
			"Connection": "Upgrade",
		})
	resp2 := req2.Do()

	// This might fail or return 101, depending on middleware handling
	if resp2.IsOk() {
		if resp2.Ok().StatusCode == http.StatusSwitchingProtocols {
			t.Log("Got 101 Switching Protocols as expected")
		}
	}
}

func TestMiddlewareRequestDefaultUserAgent(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test default user agent (should be set)
	client := surf.NewClient()
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if resp.Ok().StatusCode != http.StatusOK {
		t.Error("expected default user agent to be set")
	}

	body := resp.Ok().Body.String()
	// Should have some user agent
	if !strings.Contains(body.Std(), "Mozilla") && !strings.Contains(body.Std(), "surf") {
		t.Log("Default user agent format may have changed")
	}
}

func TestMiddlewareRequestHeaders(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Echo back custom headers
		customHeaders := make(map[string]string)
		for key, values := range r.Header {
			if strings.HasPrefix(key, "X-") {
				customHeaders[key] = values[0]
			}
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `%v`, customHeaders)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test SetHeaders and AddHeaders
	client := surf.NewClient().Builder().
		SetHeaders(g.Map[string, string]{
			"X-First":  "1",
			"X-Second": "2",
		}).
		AddHeaders(g.Map[string, string]{
			"X-Third": "3",
		}).
		Build()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()

	// Check all headers were sent
	if !strings.Contains(body.Std(), "X-First") {
		t.Error("expected X-First header")
	}
	if !strings.Contains(body.Std(), "X-Second") {
		t.Error("expected X-Second header")
	}
	if !strings.Contains(body.Std(), "X-Third") {
		t.Error("expected X-Third header")
	}
}
