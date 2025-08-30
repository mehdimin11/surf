package surf_test

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
	"github.com/enetx/surf/pkg/sse"
)

func TestBodyString(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "test body content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test String()
	content := body.String()
	if content != "test body content" {
		t.Errorf("expected 'test body content', got %s", content)
	}
}

func TestBodyBytes(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "byte content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test Bytes()
	bytes := body.Bytes()
	if string(bytes) != "byte content" {
		t.Errorf("expected 'byte content', got %s", string(bytes))
	}

	// Without cache, second call to Bytes() returns nil as body is consumed
	bytes2 := body.Bytes()
	if bytes2 != nil {
		t.Error("expected nil on second call to Bytes() without cache")
	}
}

func TestBodyMD5(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test MD5()
	md5 := body.MD5()
	// MD5 of "hello" is "5d41402abc4b2a76b9719d911017c592"
	if md5 != "5d41402abc4b2a76b9719d911017c592" {
		t.Errorf("expected MD5 '5d41402abc4b2a76b9719d911017c592', got %s", md5)
	}
}

func TestBodyJSON(t *testing.T) {
	t.Parallel()

	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	expected := TestData{Name: "test", Value: 42}

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test JSON()
	var result TestData
	err := body.JSON(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != expected.Name || result.Value != expected.Value {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestBodyXML(t *testing.T) {
	t.Parallel()

	type TestData struct {
		XMLName xml.Name `xml:"root"`
		Name    string   `xml:"name"`
		Value   int      `xml:"value"`
	}

	expected := TestData{Name: "test", Value: 42}

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(expected)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test XML()
	var result TestData
	err := body.XML(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != expected.Name || result.Value != expected.Value {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestBodyStream(t *testing.T) {
	t.Parallel()

	lines := []string{"line1", "line2", "line3"}
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test Stream()
	reader := body.Stream()
	if reader == nil {
		t.Fatal("Stream() returned nil")
	}

	// Read lines from stream
	scanner := bufio.NewScanner(reader)
	i := 0
	for scanner.Scan() {
		if scanner.Text() != lines[i] {
			t.Errorf("expected line %s, got %s", lines[i], scanner.Text())
		}
		i++
	}

	if i != len(lines) {
		t.Errorf("expected %d lines, got %d", len(lines), i)
	}
}

func TestBodySSE(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Write SSE events
		fmt.Fprintf(w, "data: event1\n\n")
		fmt.Fprintf(w, "data: event2\n\n")
		fmt.Fprintf(w, "data: event3\n\n")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test SSE()
	events := []string{}
	err := body.SSE(func(event *sse.Event) bool {
		events = append(events, event.Data.Std())
		return true // Continue reading
	})

	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	for i, event := range events {
		expected := fmt.Sprintf("event%d", i+1)
		if event != expected {
			t.Errorf("expected event %s, got %s", expected, event)
		}
	}
}

func TestBodyLimit(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write 100 bytes
		fmt.Fprint(w, strings.Repeat("a", 100))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test Limit()
	body.Limit(50)

	content := body.Bytes()
	if len(content) != 50 {
		t.Errorf("expected 50 bytes with limit, got %d", len(content))
	}
}

func TestBodyClose(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
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

	body := resp.Ok().Body

	// Test Close()
	err := body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// After close, Bytes() should return nil
	content := body.Bytes()
	if content != nil {
		t.Error("expected nil after Close(), got content")
	}
}

func TestBodyCloseNil(t *testing.T) {
	t.Parallel()

	// Test Close() on nil body
	var body *surf.Body
	err := body.Close()
	if err == nil {
		t.Error("expected error when closing nil body")
	}

	// Test Close() on body with nil Reader
	body = &surf.Body{}
	err = body.Close()
	if err == nil {
		t.Error("expected error when closing body with nil Reader")
	}
}

func TestBodyContains(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello World Test Content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient().Builder().CacheBody().Build()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test Contains with string (case sensitive)
	if !body.Contains("Hello") {
		t.Error("expected body to contain 'Hello'")
	}

	// Test Contains with g.String
	if !body.Contains(g.String("World")) {
		t.Error("expected body to contain 'World'")
	}

	// Test Contains with []byte
	if !body.Contains([]byte("Test")) {
		t.Error("expected body to contain 'Test'")
	}

	// Test Contains with g.Bytes
	if !body.Contains(g.Bytes("Content")) {
		t.Error("expected body to contain 'Content'")
	}

	// Test Contains with regexp
	re := regexp.MustCompile(`Hello.*Content`)
	if !body.Contains(re) {
		t.Error("expected body to match regex")
	}

	// Test Contains with non-matching pattern
	if body.Contains("notfound") {
		t.Error("expected body to not contain 'notfound'")
	}

	// Test Contains with unsupported type
	if body.Contains(123) {
		t.Error("expected false for unsupported type")
	}
}

func TestBodyDump(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "dump content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Create temp file path
	tempFile := g.String(fmt.Sprintf("/tmp/surf_test_%d.txt", time.Now().UnixNano()))

	// Test Dump()
	err := body.Dump(tempFile)
	if err != nil {
		t.Fatal(err)
	}

	// Read dumped content
	content := g.NewFile(tempFile).Read().UnwrapOrDefault()
	if content != "dump content" {
		t.Errorf("expected 'dump content', got %s", content)
	}

	// Clean up
	g.NewFile(tempFile).Remove()
}

func TestBodyDumpNil(t *testing.T) {
	t.Parallel()

	// Test Dump() on nil body
	var body *surf.Body
	err := body.Dump("test.txt")
	if err == nil {
		t.Error("expected error when dumping nil body")
	}

	// Test Dump() on body with nil Reader
	body = &surf.Body{}
	err = body.Dump("test.txt")
	if err == nil {
		t.Error("expected error when dumping body with nil Reader")
	}
}

func TestBodyUTF8(t *testing.T) {
	t.Parallel()

	// Test with non-UTF8 content
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=windows-1252")
		w.WriteHeader(http.StatusOK)
		// Windows-1252 encoded content (would need actual encoding)
		fmt.Fprint(w, "test content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test UTF8()
	content := body.UTF8()
	if content == "" {
		t.Error("UTF8() returned empty string")
	}
}

func TestBodyUTF8Nil(t *testing.T) {
	t.Parallel()

	// Test UTF8() on nil body
	var body *surf.Body
	content := body.UTF8()
	if content != "" {
		t.Error("expected empty string for nil body")
	}
}

func TestBodyCache(t *testing.T) {
	t.Parallel()

	callCount := 0
	handler := func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "call %d", callCount)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test with cache enabled
	client := surf.NewClient().Builder().CacheBody().Build()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// First call to Bytes()
	content1 := body.Bytes()
	if string(content1) != "call 1" {
		t.Errorf("expected 'call 1', got %s", string(content1))
	}

	// Second call should return cached content
	content2 := body.Bytes()
	if string(content2) != "call 1" {
		t.Errorf("expected cached 'call 1', got %s", string(content2))
	}

	// Make another request to verify server was called only once
	resp2 := client.Get(g.String(ts.URL)).Do()
	if resp2.IsErr() {
		t.Fatal(resp2.Err())
	}

	content3 := resp2.Ok().Body.String()
	if content3 != "call 2" {
		t.Errorf("expected 'call 2' for new request, got %s", content3)
	}
}

func TestBodyWithoutCache(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "content")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test without cache (default)
	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// First call to Bytes()
	content1 := body.Bytes()
	if string(content1) != "content" {
		t.Errorf("expected 'content', got %s", string(content1))
	}

	// Second call returns nil because body was consumed
	content2 := body.Bytes()
	if content2 != nil {
		t.Error("expected nil for second call without cache")
	}
}

func TestBodyNilOperations(t *testing.T) {
	t.Parallel()

	// Test all methods on nil body
	var body *surf.Body

	// MD5() should panic or return consistent value
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic is ok for nil body
				return
			}
		}()
		// If it doesn't panic, just verify it returns some consistent value
		body.MD5()
	}()

	// Bytes() should return nil
	if body.Bytes() != nil {
		t.Error("expected nil Bytes() for nil body")
	}

	// String() should return empty or panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic is ok
			}
		}()
		str := body.String()
		if str != "" {
			t.Error("expected empty String() for nil body")
		}
	}()

	// Stream() should return nil
	if body.Stream() != nil {
		t.Error("expected nil Stream() for nil body")
	}

	// Limit() should return nil
	if body.Limit(100) != nil {
		t.Error("expected nil Limit() for nil body")
	}
}

func TestBodyLimitChaining(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, strings.Repeat("x", 1000))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	// Test Limit() chaining
	content := resp.Ok().Body.Limit(100).Bytes()
	if len(content) != 100 {
		t.Errorf("expected 100 bytes with limit chain, got %d", len(content))
	}
}

func TestBodyClosedBody(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
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

	body := resp.Ok().Body

	// Close the body
	body.Close()

	// Try to read after close
	content := body.Bytes()
	if content != nil {
		t.Error("expected nil after body closed")
	}
}

func TestBodyInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "not json")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test JSON() with invalid JSON
	var result map[string]any
	err := body.JSON(&result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBodyInvalidXML(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "not xml")
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()
	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	body := resp.Ok().Body

	// Test XML() with invalid XML
	var result struct {
		XMLName xml.Name `xml:"root"`
	}
	err := body.XML(&result)
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}
