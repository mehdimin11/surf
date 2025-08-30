package surf_test

import (
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
)

func TestResponseBasics(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "value")
		w.Header().Set("Location", "https://example.com/redirect")
		http.SetCookie(w, &http.Cookie{
			Name:  "test",
			Value: "cookie",
		})
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "test response")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test GetResponse
	httpResp := response.GetResponse()
	if httpResp == nil {
		t.Error("GetResponse() returned nil")
	}

	// Test basic properties
	if response.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", response.StatusCode)
	}

	if response.Proto != "HTTP/1.1" {
		t.Errorf("expected proto HTTP/1.1, got %s", response.Proto)
	}

	if response.URL == nil {
		t.Error("URL is nil")
	}

	if response.Headers.Get("X-Test") != "value" {
		t.Error("expected X-Test header")
	}

	// Test Location
	if response.Location() != "https://example.com/redirect" {
		t.Errorf("expected Location header, got %s", response.Location())
	}

	// Test Referer (might be empty in test)
	_ = response.Referer()

	// Test UserAgent
	if response.UserAgent == "" {
		t.Error("UserAgent is empty")
	}

	// Test Time
	if response.Time == 0 {
		t.Error("Time is 0")
	}

	// Test ContentLength
	if response.ContentLength == 0 {
		t.Error("ContentLength is 0")
	}

	// Test Attempts
	if response.Attempts != 0 {
		t.Errorf("expected 0 attempts, got %d", response.Attempts)
	}

	// Test Cookies
	if len(response.Cookies) != 1 || response.Cookies[0].Name != "test" {
		t.Errorf("expected test cookie, got %v", response.Cookies)
	}

	// Test body content
	if !response.Body.Contains("test response") {
		t.Error("expected 'test response' in body")
	}
}

func TestResponseGetCookies(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "test1",
			Value: "value1",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "test2",
			Value: "value2",
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().Session().Build()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test GetCookies
	cookies := response.GetCookies(g.String(ts.URL))
	if len(cookies) < 2 {
		t.Errorf("expected at least 2 cookies, got %d", len(cookies))
	}

	// Test with invalid URL (should return nil)
	cookies = response.GetCookies("")
	if cookies != nil {
		t.Error("expected nil for invalid URL")
	}
}

func TestResponseSetCookies(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().Session().Build()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test SetCookies
	cookies := []*http.Cookie{
		{Name: "new1", Value: "value1"},
		{Name: "new2", Value: "value2"},
	}
	err := response.SetCookies(g.String(ts.URL), cookies)
	if err != nil {
		t.Fatal(err)
	}

	// Verify cookies were set
	setCookies := response.GetCookies(g.String(ts.URL))
	found := 0
	for _, cookie := range setCookies {
		if cookie.Name == "new1" || cookie.Name == "new2" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected 2 new cookies, found %d", found)
	}

	// Test with invalid URL
	err = response.SetCookies("", cookies)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestResponseSetCookiesWithoutJar(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Client without session (no cookie jar)
	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test SetCookies without jar
	cookies := []*http.Cookie{{Name: "test", Value: "value"}}
	err := response.SetCookies(g.String(ts.URL), cookies)
	if err == nil {
		t.Error("expected error when setting cookies without jar")
	}
}

func TestResponseRemoteAddress(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test with GetRemoteAddress enabled
	client := surf.NewClient().Builder().GetRemoteAddress().Build()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	remoteAddr := resp.Ok().RemoteAddress()
	if remoteAddr == nil {
		t.Error("expected remote address to be captured")
	} else {
		// Verify it's a TCP address
		if _, ok := remoteAddr.(*net.TCPAddr); !ok {
			t.Errorf("expected *net.TCPAddr, got %T", remoteAddr)
		}
	}

	// Test without GetRemoteAddress
	client = surf.NewClient()
	resp = client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	remoteAddr = resp.Ok().RemoteAddress()
	if remoteAddr != nil {
		t.Error("expected nil remote address when not enabled")
	}
}

func TestResponseTLSGrabber(t *testing.T) {
	t.Parallel()

	// Create HTTPS test server
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Test TLSGrabber
	tlsData := resp.Ok().TLSGrabber()
	if tlsData == nil {
		t.Error("expected TLS data for HTTPS connection")
	}
}

func TestResponseTLSGrabberHTTP(t *testing.T) {
	t.Parallel()

	// Create HTTP test server (not HTTPS)
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Test TLSGrabber for non-TLS connection
	tlsData := resp.Ok().TLSGrabber()
	if tlsData != nil {
		t.Error("expected nil TLS data for HTTP connection")
	}
}

func TestResponseWithRetry(t *testing.T) {
	t.Parallel()

	attemptCount := 0
	handler := func(w http.ResponseWriter, _ *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "success")
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().
		Retry(5, 10*time.Millisecond).
		Build()

	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test Attempts count
	if response.Attempts != 2 {
		t.Errorf("expected 2 retry attempts, got %d", response.Attempts)
	}

	if !response.Body.Contains("success") {
		t.Error("expected 'success' in body")
	}
}

func TestResponseHeaders(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Single", "value")
		w.Header().Add("X-Multiple", "value1")
		w.Header().Add("X-Multiple", "value2")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "test")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test single header
	if response.Headers.Get("X-Single") != "value" {
		t.Errorf("expected X-Single=value, got %s", response.Headers.Get("X-Single"))
	}

	// Test multiple headers
	values := response.Headers.Values("X-Multiple")
	if len(values) != 2 || values[0] != "value1" || values[1] != "value2" {
		t.Errorf("expected X-Multiple=[value1, value2], got %v", values)
	}

	// Test header contains
	if !response.Headers.Contains("Content-Type", "text/plain") {
		t.Error("expected Content-Type to contain text/plain")
	}

	// Test Clone
	clonedHeaders := response.Headers.Clone()
	if clonedHeaders.Get("X-Single") != "value" {
		t.Error("cloned headers missing X-Single")
	}
}

func TestResponseURL(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL + "/path?query=value")).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	if response.URL == nil {
		t.Fatal("URL is nil")
	}

	// Verify URL components
	if response.URL.Path != "/path" {
		t.Errorf("expected path /path, got %s", response.URL.Path)
	}

	if response.URL.Query().Get("query") != "value" {
		t.Errorf("expected query=value, got %s", response.URL.Query().Get("query"))
	}

	parsedURL, _ := url.Parse(ts.URL)
	if response.URL.Host != parsedURL.Host {
		t.Errorf("expected host %s, got %s", parsedURL.Host, response.URL.Host)
	}
}

func TestResponseRedirect(t *testing.T) {
	t.Parallel()

	redirectCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if redirectCount == 0 {
			redirectCount++
			http.Redirect(w, r, "/redirected", http.StatusFound)
			return
		}
		w.Header().Set("X-Redirected", "true")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "redirected")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Check final URL after redirect
	if response.URL.Path != "/redirected" {
		t.Errorf("expected path /redirected after redirect, got %s", response.URL.Path)
	}

	// Check response from redirected page
	if !response.Body.Contains("redirected") {
		t.Error("expected 'redirected' in body")
	}

	if response.Headers.Get("X-Redirected") != "true" {
		t.Error("expected X-Redirected header after redirect")
	}
}

func TestResponseClientMethods(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "first")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	// Test that response has Client methods
	if response.GetClient() == nil {
		t.Error("GetClient() returned nil")
	}

	if response.GetDialer() == nil {
		t.Error("GetDialer() returned nil")
	}

	if response.GetTransport() == nil {
		t.Error("GetTransport() returned nil")
	}

	if response.GetTLSConfig() == nil {
		t.Error("GetTLSConfig() returned nil")
	}

	// Test chaining - make another request using the response
	resp2 := response.Get(g.String(ts.URL + "/second")).Do()
	if resp2.IsErr() {
		t.Fatal(resp2.Err())
	}

	// Should use the same client
	if resp2.Ok().GetClient() != response.GetClient() {
		t.Error("expected same client for chained request")
	}
}

func TestResponseWithMiddleware(t *testing.T) {
	t.Parallel()

	var middlewareCalled bool

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().
		With(func(resp *surf.Response) error {
			middlewareCalled = true
			// Verify response properties are available in middleware
			if resp.StatusCode != 200 {
				t.Errorf("expected status 200 in middleware, got %d", resp.StatusCode)
			}
			return nil
		}).
		Build()

	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !middlewareCalled {
		t.Error("response middleware was not called")
	}
}

func TestResponseNilBody(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()

	// HEAD request should have nil body
	resp := client.Head(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	response := resp.Ok()

	if response.Body != nil {
		t.Error("expected nil body for HEAD request")
	}

	// GetResponse should still work
	if response.GetResponse() == nil {
		t.Error("GetResponse() returned nil")
	}
}

func TestResponseTLSProperties(t *testing.T) {
	t.Parallel()

	// Use httptest TLS server since custom cert parsing fails
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer server.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(server.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Test TLS is present
	httpResp := resp.Ok().GetResponse()
	if httpResp.TLS == nil {
		t.Error("expected TLS connection state")
	}
}
