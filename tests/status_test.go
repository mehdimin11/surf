package surf_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/enetx/g"
	ehttp "github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
)

func TestStatusCodeSuccess(t *testing.T) {
	t.Parallel()

	successCodes := []int{200, 201, 202, 204, 206}

	for _, code := range successCodes {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode

		if int(statusCode) != code {
			t.Errorf("expected status code %d, got %d", code, statusCode)
		}

		if !statusCode.IsSuccess() {
			t.Errorf("expected status code %d to be success", code)
		}
	}
}

func TestStatusCodeClientError(t *testing.T) {
	t.Parallel()

	clientErrorCodes := []int{400, 401, 403, 404, 405, 409, 429}

	for _, code := range clientErrorCodes {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode

		if int(statusCode) != code {
			t.Errorf("expected status code %d, got %d", code, statusCode)
		}

		if !statusCode.IsClientError() {
			t.Errorf("expected status code %d to be client error", code)
		}

		if statusCode.IsSuccess() {
			t.Errorf("expected status code %d to not be success", code)
		}
	}
}

func TestStatusCodeServerError(t *testing.T) {
	t.Parallel()

	serverErrorCodes := []int{500, 501, 502, 503, 504, 505}

	for _, code := range serverErrorCodes {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode

		if int(statusCode) != code {
			t.Errorf("expected status code %d, got %d", code, statusCode)
		}

		if !statusCode.IsServerError() {
			t.Errorf("expected status code %d to be server error", code)
		}

		if statusCode.IsSuccess() {
			t.Errorf("expected status code %d to not be success", code)
		}
	}
}

func TestStatusCodeRedirect(t *testing.T) {
	t.Parallel()

	redirectCodes := []int{301, 302, 303, 307, 308}

	for _, code := range redirectCodes {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode

		if int(statusCode) != code {
			t.Errorf("expected status code %d, got %d", code, statusCode)
		}

		if !statusCode.IsRedirection() {
			t.Errorf("expected status code %d to be redirect", code)
		}

		if statusCode.IsSuccess() {
			t.Errorf("expected status code %d to not be success", code)
		}
	}
}

func TestStatusCodeInformational(t *testing.T) {
	t.Parallel()

	// Note: 1xx codes are difficult to test with httptest as they are handled
	// differently. We'll test the classification methods with known codes.

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

	statusCode := resp.Ok().StatusCode

	// Test that 200 is NOT informational
	if statusCode.IsInformational() {
		t.Error("expected 200 to not be informational")
	}
}

func TestStatusCodeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		code     int
		expected string
	}{
		{200, "200 OK"},
		{404, "404 Not Found"},
		{500, "500 Internal Server Error"},
	}

	for _, tc := range testCases {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(tc.code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode
		statusStr := fmt.Sprintf("%d %s", statusCode, statusCode.Text())

		if statusStr != tc.expected {
			t.Errorf("expected status string to be %s, got %s", tc.expected, statusStr)
		}
	}
}

func TestStatusCodeComparison(t *testing.T) {
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

	statusCode := resp.Ok().StatusCode

	// Test equality
	if statusCode != 200 {
		t.Error("expected status code to equal 200")
	}

	// Test type conversion
	if int(statusCode) != 200 {
		t.Error("expected status code to convert to int 200")
	}
}

func TestStatusCodeRetryLogic(t *testing.T) {
	t.Parallel()

	// Test that certain status codes should trigger retries
	retryCodes := []int{429, 500, 502, 503, 504}

	for _, code := range retryCodes {
		handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
			w.WriteHeader(code)
		}

		ts := httptest.NewServer(ehttp.HandlerFunc(handler))
		defer ts.Close()

		client := surf.NewClient()
		resp := client.Get(g.String(ts.URL)).Do()

		if resp.IsErr() {
			t.Fatal(resp.Err())
		}

		statusCode := resp.Ok().StatusCode

		// These codes should typically be retryable
		shouldRetry := statusCode.IsServerError() || statusCode == 429
		if !shouldRetry {
			t.Errorf("expected status code %d to be potentially retryable", code)
		}
	}
}

func TestStatusCodeEdgeCases(t *testing.T) {
	t.Parallel()

	// Test with 418 I'm a teapot
	handler := func(w ehttp.ResponseWriter, _ *ehttp.Request) {
		w.WriteHeader(418)
	}

	ts := httptest.NewServer(ehttp.HandlerFunc(handler))
	defer ts.Close()

	client := surf.NewClient()
	resp := client.Get(g.String(ts.URL)).Do()

	if resp.IsErr() {
		t.Fatal(resp.Err())
	}

	statusCode := resp.Ok().StatusCode

	if int(statusCode) != 418 {
		t.Errorf("expected status code 418, got %d", statusCode)
	}

	if !statusCode.IsClientError() {
		t.Error("expected 418 to be client error")
	}

	if statusCode.IsSuccess() {
		t.Error("expected 418 to not be success")
	}
}
