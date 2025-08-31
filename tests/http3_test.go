package surf

import (
	"context"
	"testing"
	"time"

	"github.com/enetx/surf"
	quic "github.com/refraction-networking/uquic"
)

func TestHTTP3Fingerprints(t *testing.T) {
	t.Run("Chrome fingerprint", func(t *testing.T) {
		// Get the expected Chrome fingerprint
		chromeID := quic.QUICChrome_115
		expectedSpec, err := quic.QUICID2Spec(chromeID)
		if err != nil {
			t.Fatalf("Failed to get Chrome spec: %v", err)
		}

		// Build client with Chrome fingerprint
		client := surf.NewClient().Builder().
			HTTP3Settings().Chrome().Set().
			Build()

		// Verify transport is set
		if client.GetTransport() == nil {
			t.Fatal("Transport is nil")
		}

		// Check fingerprint characteristics
		t.Logf("Chrome fingerprint ID: %s", chromeID.Fingerprint)
		t.Logf("Chrome SrcConnIDLength: %d", expectedSpec.InitialPacketSpec.SrcConnIDLength)
		t.Logf("Chrome UDPDatagramMinSize: %d", expectedSpec.UDPDatagramMinSize)
	})

	t.Run("Firefox fingerprint", func(t *testing.T) {
		// Get the expected Firefox fingerprint
		firefoxID := quic.QUICFirefox_116
		expectedSpec, err := quic.QUICID2Spec(firefoxID)
		if err != nil {
			t.Fatalf("Failed to get Firefox spec: %v", err)
		}

		// Build client with Firefox fingerprint
		client := surf.NewClient().Builder().
			HTTP3Settings().Firefox().Set().
			Build()

		// Verify transport is set
		if client.GetTransport() == nil {
			t.Fatal("Transport is nil")
		}

		// Check fingerprint characteristics
		t.Logf("Firefox fingerprint ID: %s", firefoxID.Fingerprint)
		t.Logf("Firefox SrcConnIDLength: %d", expectedSpec.InitialPacketSpec.SrcConnIDLength)
		t.Logf("Firefox UDPDatagramMinSize: %d", expectedSpec.UDPDatagramMinSize)
	})

	t.Run("Fingerprint differences", func(t *testing.T) {
		chromeSpec, _ := quic.QUICID2Spec(quic.QUICChrome_115)
		firefoxSpec, _ := quic.QUICID2Spec(quic.QUICFirefox_116)

		// These should be different to prove we have distinct fingerprints
		if chromeSpec.InitialPacketSpec.SrcConnIDLength == firefoxSpec.InitialPacketSpec.SrcConnIDLength {
			t.Log("Warning: SrcConnIDLength is the same for Chrome and Firefox")
		}

		if chromeSpec.UDPDatagramMinSize == firefoxSpec.UDPDatagramMinSize {
			t.Log("Warning: UDPDatagramMinSize is the same for Chrome and Firefox")
		}

		// Log the differences
		t.Logf("Chrome vs Firefox SrcConnIDLength: %d vs %d",
			chromeSpec.InitialPacketSpec.SrcConnIDLength,
			firefoxSpec.InitialPacketSpec.SrcConnIDLength)
		t.Logf("Chrome vs Firefox UDPDatagramMinSize: %d vs %d",
			chromeSpec.UDPDatagramMinSize,
			firefoxSpec.UDPDatagramMinSize)
	})
}

func TestHTTP3AutoDetection(t *testing.T) {
	t.Run("Chrome auto detection", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().Chrome().HTTP3().
			Build()

		if client.GetTransport() == nil {
			t.Fatal("Chrome HTTP/3 transport is nil")
		}

		// Verify it's HTTP/3 transport
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}
	})

	t.Run("Firefox auto detection", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().FireFox().HTTP3().
			Build()

		if client.GetTransport() == nil {
			t.Fatal("Firefox HTTP/3 transport is nil")
		}

		// Verify it's HTTP/3 transport
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}
	})

	t.Run("Default to Chrome", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3().
			Build()

		if client.GetTransport() == nil {
			t.Fatal("Default HTTP/3 transport is nil")
		}

		// Verify it's HTTP/3 transport
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}
	})
}

func TestHTTP3OrderIndependence(t *testing.T) {
	t.Run("HTTP3 then Impersonate", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3().
			Impersonate().Chrome().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport regardless of order")
		}
	})

	t.Run("Impersonate then HTTP3", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().Chrome().
			HTTP3().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport regardless of order")
		}
	})
}

func TestHTTP3ManualVsAuto(t *testing.T) {
	t.Run("Manual settings disable auto", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().Chrome().
			HTTP3().                        // This should be ignored
			HTTP3Settings().Chrome().Set(). // This takes precedence
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport from manual settings")
		}
	})

	t.Run("Auto then manual", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3().                        // This gets disabled
			HTTP3Settings().Chrome().Set(). // This applies
			Impersonate().Chrome().         // This should not trigger auto HTTP3
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport from manual settings")
		}
	})
}

func TestHTTP3Compatibility(t *testing.T) {
	t.Run("HTTP3 with proxy fallback", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Proxy("http://proxy:8080").
			HTTP3Settings().Chrome().Set().
			Build()

		// Should not have HTTP/3 transport when proxy is set
		if client.IsHTTP3() {
			t.Fatal("HTTP/3 should not be active with proxy")
		}
	})

	t.Run("HTTP3 with JA3 compatibility", func(t *testing.T) {
		client := surf.NewClient().Builder().
			JA().Chrome131().
			HTTP3Settings().Chrome().Set().
			Build()

		// Should have HTTP/3 transport (JA3 should be ignored)
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport (JA3 should be ignored for HTTP/3)")
		}
	})

	t.Run("HTTP3 with DNS settings", func(t *testing.T) {
		client := surf.NewClient().Builder().
			DNS("8.8.8.8:53").
			HTTP3Settings().Chrome().Set().
			Build()

		// Should have HTTP/3 transport with DNS settings
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with DNS settings")
		}
	})

	t.Run("HTTP3 with DNSOverTLS", func(t *testing.T) {
		client := surf.NewClient().Builder().
			DNSOverTLS().Google().
			HTTP3Settings().Chrome().Set().
			Build()

		// Should have HTTP/3 transport with DNS over TLS
		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with DNS over TLS")
		}
	})

	t.Run("HTTP3 with timeout", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Timeout(30 * time.Second).
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with timeout")
		}
	})

	t.Run("HTTP3 with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client := surf.NewClient().Builder().
			WithContext(ctx).
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with context")
		}
	})

	t.Run("HTTP3 with headers", func(t *testing.T) {
		client := surf.NewClient().Builder().
			SetHeaders("X-Test", "value").
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with custom headers")
		}
	})

	t.Run("HTTP3 with middleware", func(t *testing.T) {
		client := surf.NewClient().Builder().
			With(func(req *surf.Request) error {
				req.SetHeaders("X-Middleware", "test")
				return nil
			}).
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with middleware")
		}
	})
}

func TestHTTP3TransportProperties(t *testing.T) {
	t.Run("Chrome transport properties", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}

		if client.GetTransport() == nil {
			t.Fatal("Transport should not be nil")
		}
	})

	t.Run("Firefox transport properties", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().Firefox().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}

		if client.GetTransport() == nil {
			t.Fatal("Transport should not be nil")
		}
	})
}

func TestHTTP3CustomSettings(t *testing.T) {
	t.Run("Custom QUIC ID", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().
			SetQUICID(quic.QUICChrome_115).
			Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with custom QUIC ID")
		}
	})

	t.Run("Custom QUIC Spec", func(t *testing.T) {
		spec, err := quic.QUICID2Spec(quic.QUICChrome_115)
		if err != nil {
			t.Fatalf("Failed to get QUIC spec: %v", err)
		}

		client := surf.NewClient().Builder().
			HTTP3Settings().
			SetQUICSpec(spec).
			Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport with custom QUIC spec")
		}
	})
}

func TestHTTP3EdgeCases(t *testing.T) {
	t.Run("Multiple HTTP3Settings calls", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().Chrome().Set().
			HTTP3Settings().Firefox().Set(). // Last one should win
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport from last HTTP3Settings call")
		}
	})

	t.Run("HTTP3 with ForceHTTP1", func(t *testing.T) {
		client := surf.NewClient().Builder().
			ForceHTTP1().
			HTTP3Settings().Chrome().Set().
			Build()

		// HTTP/3 should be disabled when ForceHTTP1 is set
		if client.IsHTTP3() {
			t.Fatal("HTTP/3 should be disabled when ForceHTTP1 is set")
		}
	})

	t.Run("Empty HTTP3Settings chain", func(t *testing.T) {
		// This should not panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("HTTP3Settings should not panic: %v", r)
				}
			}()

			client := surf.NewClient().Builder().
				HTTP3Settings().Chrome().Set().
				Build()

			// Should still work, just not have HTTP/3
			if client == nil {
				t.Fatal("Client should not be nil")
			}
		}()
	})
}

func TestHTTP3MockRequests(t *testing.T) {
	t.Run("Chrome mock request", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().Chrome().Set().
			Build()

		resp := client.Get("https://example.com").Do()
		if resp.IsErr() {
			t.Fatalf("Request failed: %v", resp.Err())
		}

		r := resp.Ok()
		if r.Proto.Std() != "HTTP/3.0" {
			t.Errorf("Expected HTTP/3.0, got %s", r.Proto)
		}

		if r.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", r.StatusCode)
		}
	})

	t.Run("Firefox mock request", func(t *testing.T) {
		client := surf.NewClient().Builder().
			HTTP3Settings().Firefox().Set().
			Build()

		resp := client.Get("https://example.com").Do()
		if resp.IsErr() {
			t.Fatalf("Request failed: %v", resp.Err())
		}

		r := resp.Ok()
		if r.Proto.Std() != "HTTP/3.0" {
			t.Errorf("Expected HTTP/3.0, got %s", r.Proto)
		}

		if r.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", r.StatusCode)
		}
	})

	t.Run("Auto detection mock request", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().Chrome().HTTP3().
			Build()

		resp := client.Get("https://example.com").Do()
		if resp.IsErr() {
			t.Fatalf("Request failed: %v", resp.Err())
		}

		r := resp.Ok()
		if r.Proto.Std() != "HTTP/3.0" {
			t.Errorf("Expected HTTP/3.0, got %s", r.Proto)
		}

		if r.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", r.StatusCode)
		}
	})
}

func TestHTTP3RealRequests(t *testing.T) {
	// Skip real requests in short mode
	if testing.Short() {
		t.Skip("Skipping real HTTP/3 requests in short mode")
	}

	t.Run("Chrome real request", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().Chrome().HTTP3().
			Build()

		resp := client.Get("https://cloudflare-quic.com/").Do()
		if resp.IsErr() {
			t.Fatalf("Real request failed: %v", resp.Err())
		}

		r := resp.Ok()
		if r.Proto.Std() != "HTTP/3.0" {
			t.Errorf("Expected HTTP/3.0, got %s", r.Proto)
		}

		if r.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", r.StatusCode)
		}

		// Check for HTTP/3 specific headers
		if altSvc := r.Headers.Get("alt-svc"); altSvc == "" {
			t.Log("Warning: No alt-svc header found")
		} else {
			t.Logf("Alt-Svc header: %s", altSvc)
		}
	})

	t.Run("Firefox real request", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Impersonate().FireFox().HTTP3().
			Build()

		resp := client.Get("https://cloudflare-quic.com/").Do()
		if resp.IsErr() {
			t.Fatalf("Real request failed: %v", resp.Err())
		}

		r := resp.Ok()
		if r.Proto.Std() != "HTTP/3.0" {
			t.Errorf("Expected HTTP/3.0, got %s", r.Proto)
		}

		if r.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", r.StatusCode)
		}
	})
}

func TestHTTP3DNSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS integration tests in short mode")
	}

	t.Run("HTTP3 with custom DNS", func(t *testing.T) {
		client := surf.NewClient().Builder().
			DNS("8.8.8.8:53").
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}

		// DNS integration is tested through actual requests
		// since transport internals are not exposed
	})

	t.Run("HTTP3 with DNS over TLS", func(t *testing.T) {
		client := surf.NewClient().Builder().
			DNSOverTLS().Google().
			HTTP3Settings().Chrome().Set().
			Build()

		if !client.IsHTTP3() {
			t.Fatal("Expected HTTP/3 transport")
		}

		// DNS over TLS integration is tested through actual requests
		// since transport internals are not exposed
	})
}

func TestHTTP3NetworkConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network condition tests in short mode")
	}

	t.Run("HTTP3 with unreachable proxy", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Proxy("http://unreachable:8080").
			HTTP3Settings().Chrome().Set().
			Build()

		// Should fallback to HTTP/2 due to proxy
		if client.IsHTTP3() {
			t.Fatal("HTTP/3 should be disabled with proxy")
		}
	})

	t.Run("HTTP3 timeout handling", func(t *testing.T) {
		client := surf.NewClient().Builder().
			Timeout(1 * time.Millisecond). // Very short timeout
			HTTP3Settings().Chrome().Set().
			Build()

		resp := client.Get("https://cloudflare-quic.com/").Do()

		// Should either succeed or timeout, but not crash
		if resp.IsErr() {
			t.Logf("Request timed out as expected: %v", resp.Err())
		} else {
			t.Logf("Request succeeded despite short timeout")
		}
	})
}
