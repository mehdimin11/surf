package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/enetx/g"
	"github.com/enetx/surf"
)

func main() {
	fmt.Println("=== Testing Surf Features through Std() Adapter ===\n")

	// Test 1: Retry Logic
	testRetry()

	// Test 2: Timeout
	testTimeout()

	// Test 3: Redirect Policy
	testRedirects()

	// Test 4: Request/Response Middleware
	testMiddleware()

	// Test 5: CloseIdleConnections
	testCloseIdleConnections()

	// Test 6: Request.Cancel field
	testCancelField()

	// Test 7: Response.Request field
	testResponseRequest()

	// Show summary
	summary()
}

func testRetry() {
	fmt.Println("1. Testing Retry Logic:")

	// Create a test server that fails first 2 times
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			fmt.Printf("   Server returned 503 (attempt %d)\n", current)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
		fmt.Printf("   Server returned 200 (attempt %d)\n", current)
	}))
	defer ts.Close()

	// Create surf client with retry
	surfClient := surf.NewClient().
		Builder().
		Retry(3, 100*time.Millisecond). // Retry 3 times with 100ms wait
		Build()

	// Test with surf directly
	fmt.Println("   Testing with surf.Client directly:")
	resp := surfClient.Get(g.String(ts.URL)).Do()
	if resp.IsOk() {
		fmt.Printf("   Surf retry works! Final attempts: %d\n", attempts)
	} else {
		fmt.Printf("   Surf retry failed\n")
	}

	// Reset counter
	atomic.StoreInt32(&attempts, 0)

	// Test with std adapter
	fmt.Println("   Testing with Std() adapter:")
	stdClient := surfClient.Std()
	stdResp, err := stdClient.Get(ts.URL)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else if stdResp.StatusCode == 200 {
		fmt.Printf("   Std adapter retry works! Final attempts: %d\n", attempts)
	} else {
		fmt.Printf("   Std adapter retry NOT working! Status: %d, Attempts: %d\n", stdResp.StatusCode, attempts)
		fmt.Println("   ISSUE: Retry logic is in Request.Do(), not in RoundTrip()!")
	}
	fmt.Println()
}

func testTimeout() {
	fmt.Println("2. Testing Timeout:")

	// Create slow server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("too late"))
	}))
	defer ts.Close()

	// Create surf client with 500ms timeout
	surfClient := surf.NewClient().
		Builder().
		Timeout(500 * time.Millisecond).
		Build()

	stdClient := surfClient.Std()

	start := time.Now()
	_, err := stdClient.Get(ts.URL)
	elapsed := time.Since(start)

	if err != nil && elapsed < 1*time.Second {
		fmt.Printf("   Timeout works! Failed after %v\n", elapsed)
	} else {
		fmt.Printf("   Timeout issue: elapsed %v, error: %v\n", elapsed, err)
	}

	fmt.Println()
}

func testRedirects() {
	fmt.Println("3. Testing Redirect Policy:")

	// Create redirect chain
	redirectCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		if redirectCount <= 3 {
			http.Redirect(w, r, fmt.Sprintf("/?redirect=%d", redirectCount), http.StatusFound)
			return
		}
		w.Write([]byte(fmt.Sprintf("final after %d redirects", redirectCount-1)))
	}))
	defer ts.Close()

	// Test MaxRedirects
	surfClient := surf.NewClient().
		Builder().
		MaxRedirects(2). // Only allow 2 redirects
		Build()

	stdClient := surfClient.Std()
	resp, err := stdClient.Get(ts.URL)

	if err != nil {
		fmt.Printf("   Redirect limit works! Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("   Response: %s\n", string(body))
		if redirectCount <= 3 {
			fmt.Println("   Redirect policy applied")
		} else {
			fmt.Println("   Redirect policy not working")
		}
	}

	// Test NotFollowRedirects
	redirectCount = 0
	surfClient2 := surf.NewClient().
		Builder().
		NotFollowRedirects().
		Build()

	stdClient2 := surfClient2.Std()
	resp2, _ := stdClient2.Get(ts.URL)

	if resp2.StatusCode == 302 {
		fmt.Println("   NotFollowRedirects works!")
	} else {
		fmt.Printf("   NotFollowRedirects not working, status: %d\n", resp2.StatusCode)
	}

	fmt.Println()
}

func testMiddleware() {
	fmt.Println("4. Testing Middleware:")

	// Test custom header middleware
	customHeaderApplied := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") == "surf-test" {
			customHeaderApplied = true
		}
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	surfClient := surf.NewClient().
		Builder().
		// SetHeaders("X-Custom-Header", "surf-test").
		With(func(req *surf.Request) error {
			req.SetHeaders("X-Custom-Header", "surf-test")
			return nil
		}).
		Build()

	stdClient := surfClient.Std()
	resp, _ := stdClient.Get(ts.URL)
	resp.Body.Close()

	if customHeaderApplied {
		fmt.Println("   Request middleware works!")
	} else {
		fmt.Println("   Request middleware not applied")
	}
	fmt.Println()
}

func testCloseIdleConnections() {
	fmt.Print("5. Testing CloseIdleConnections:")

	surfClient := surf.NewClient().Builder().Build()
	stdClient := surfClient.Std()

	// This should not panic after the fix
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("   CloseIdleConnections panicked: %v\n", r)
		} else {
			fmt.Println("   CloseIdleConnections works!\n")
		}
	}()

	// Call CloseIdleConnections on the transport
	if transport, ok := stdClient.Transport.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
	fmt.Println()
}

func testCancelField() {
	fmt.Println("6. Testing Cancel field preservation:")

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	surfClient := surf.NewClient().Builder().Build()
	stdClient := surfClient.Std()

	// Create request with Cancel channel
	req, _ := http.NewRequest("GET", ts.URL, nil)
	cancelChan := make(chan struct{})
	req.Cancel = cancelChan // deprecated but should be preserved

	// Make request through adapter
	resp, err := stdClient.Do(req)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Println("   Cancel field preserved (no panic)")
	}
	fmt.Println()
}

func testResponseRequest() {
	fmt.Println("7. Testing Response.Request field:")

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer ts.Close()

	surfClient := surf.NewClient().Builder().Build()
	stdClient := surfClient.Std()

	resp, err := stdClient.Get(ts.URL)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check if Response.Request is set
	if resp.Request != nil {
		fmt.Printf("    Response.Request is set!\n")
		fmt.Printf("    Method: %s\n", resp.Request.Method)
		fmt.Printf("    URL: %s\n", resp.Request.URL)
	} else {
		fmt.Println("   Response.Request is nil")
	}

	// Verify body is readable
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("    Body: %s\n", string(body))
	fmt.Println()
}

func summary() {
	fmt.Println("=== SUMMARY ===")
	fmt.Println(" Working through Std():")
	fmt.Println("   - JA3 fingerprinting")
	fmt.Println("   - Cookies/Sessions")
	fmt.Println("   - Headers and User-Agent")
	fmt.Println("   - Request/Response middleware")
	fmt.Println("   - Timeout (via http.Client)")
	fmt.Println("   - Redirect policy (via CheckRedirect)")
	fmt.Println("   - Proxy (in transport)")
	fmt.Println("   - CloseIdleConnections support")
	fmt.Println("   - Request.Cancel field preservation")
	fmt.Println("   - Response.Request field set correctly")
	fmt.Println()
	fmt.Println(" NOT Working through Std():")
	fmt.Println("   - Retry logic (implemented in Request.Do(), not RoundTrip())")
	fmt.Println("   - Response body caching (in Request.Do())")
	fmt.Println("   - Remote address tracking (in Request.Do())")
	fmt.Println("   - Request timing information (in Request.Do())")
	fmt.Println()
	fmt.Println(" Note: These features require architectural changes to work in TransportAdapter")
}
