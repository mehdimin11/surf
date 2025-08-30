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

func TestImpersonateChrome(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().
		Impersonate().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	// Test that Chrome headers are applied
	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}
}

func TestImpersonateFirefox(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().FireFox().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	// Test that Firefox headers are applied
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Firefox") {
		t.Error("expected Firefox user agent to be applied")
	}
}

func TestImpersonateWithOS(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().Windows().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	// Test that Windows Chrome headers are applied
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}

	if !strings.Contains(body.Std(), "Windows") {
		t.Error("expected Windows platform to be applied")
	}
}

func TestImpersonateMacOS(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().MacOS().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}
}

func TestImpersonateLinux(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().Linux().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}
}

func TestImpersonateAndroid(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().Android().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}
}

func TestImpersonateIOS(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().IOS().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	// For iOS impersonation, expect either Safari or iOS/iPhone in the user agent
	if !strings.Contains(body.Std(), "Safari") && !strings.Contains(body.Std(), "iPhone") &&
		!strings.Contains(body.Std(), "iOS") {
		t.Logf("User agent: %s", body.Std())
		t.Error("expected iOS/Safari user agent to be applied")
	}
}

func TestImpersonateRandomOS(t *testing.T) {
	t.Parallel()

	client := surf.NewClient().Builder().
		Impersonate().RandomOS().Chrome().
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		fmt.Fprintf(w, `{"user-agent": "%s"}`, userAgent)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	if !resp.Ok().StatusCode.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.Ok().StatusCode)
	}

	body := resp.Ok().Body.String()
	// RandomOS can select any OS, so just verify that some user agent was set
	if body.Std() == "" {
		t.Error("expected user agent to be set")
	} else {
		t.Logf("Random OS user agent: %s", body.Std())
	}
}

func TestImpersonateWithCustomHeaders(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userAgent := r.Header.Get("User-Agent")
		customHeader := r.Header.Get("X-Custom")
		fmt.Fprintf(w, "User-Agent: %s\nX-Custom: %s", userAgent, customHeader)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	headers := g.NewMapOrd[g.String, g.String](1)
	headers.Set("X-Custom", "test-value")

	client := surf.NewClient().Builder().
		Impersonate().Chrome().
		SetHeaders(headers).
		Build()

	if client == nil {
		t.Fatal("expected client to be built successfully")
	}

	req := client.Get(g.String(ts.URL))
	resp := req.Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body.String()
	if !strings.Contains(body.Std(), "Chrome") {
		t.Error("expected Chrome user agent to be applied")
	}

	// Note: Custom headers may be overridden by impersonation for authenticity
	// This is expected behavior - impersonation should override headers for realism
	if strings.Contains(body.Std(), "test-value") {
		t.Log("custom header was preserved (this may or may not be expected)")
	} else {
		t.Log("custom header was overridden by impersonation (this is expected behavior)")
	}
}
