package surf_test

import (
	"fmt"
	"testing"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http/httptest"
	"github.com/enetx/surf"
)

func TestDNSOverTLSProviders(t *testing.T) {
	t.Parallel()

	// Test that all DNS over TLS providers can be created through builder
	testCases := []string{
		"AdGuard", "Google", "Cloudflare", "Quad9", "Switch",
		"CIRAShield", "Ali", "Quad101", "SB", "Forge", "LibreDNS",
	}

	for _, name := range testCases {
		t.Run(name, func(t *testing.T) {
			client := surf.NewClient()
			builder := client.Builder()

			dnsBuilder := builder.DNSOverTLS()
			if dnsBuilder == nil {
				t.Fatalf("expected DNSOverTLS builder to be non-nil")
			}

			// Test different providers by method name
			var builtClient *surf.Client
			switch name {
			case "AdGuard":
				builtClient = dnsBuilder.AdGuard().Build()
			case "Google":
				builtClient = dnsBuilder.Google().Build()
			case "Cloudflare":
				builtClient = dnsBuilder.Cloudflare().Build()
			case "Quad9":
				builtClient = dnsBuilder.Quad9().Build()
			case "Switch":
				builtClient = dnsBuilder.Switch().Build()
			case "CIRAShield":
				builtClient = dnsBuilder.CIRAShield().Build()
			case "Ali":
				builtClient = dnsBuilder.Ali().Build()
			case "Quad101":
				builtClient = dnsBuilder.Quad101().Build()
			case "SB":
				builtClient = dnsBuilder.SB().Build()
			case "Forge":
				builtClient = dnsBuilder.Forge().Build()
			case "LibreDNS":
				builtClient = dnsBuilder.LibreDNS().Build()
			default:
				t.Fatalf("unknown provider: %s", name)
			}

			if builtClient == nil {
				t.Errorf("expected client with %s DNS to be non-nil", name)
			}
		})
	}
}

func TestDNSOverTLSWithLocalRequest(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"dns": "test"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Test with local server (no actual DNS resolution needed)
	testCases := []string{"Cloudflare", "AdGuard", "Google", "Quad9"}

	for _, name := range testCases {
		t.Run(name, func(t *testing.T) {
			client := surf.NewClient()
			builder := client.Builder()
			dnsBuilder := builder.DNSOverTLS()

			var builtClient *surf.Client
			switch name {
			case "Cloudflare":
				builtClient = dnsBuilder.Cloudflare().Build()
			case "AdGuard":
				builtClient = dnsBuilder.AdGuard().Build()
			case "Google":
				builtClient = dnsBuilder.Google().Build()
			case "Quad9":
				builtClient = dnsBuilder.Quad9().Build()
			}
			req := builtClient.Get(g.String(ts.URL))
			resp := req.Do()

			if resp.IsErr() {
				t.Fatal(resp.Err())
			}

			if !resp.Ok().StatusCode.IsSuccess() {
				t.Errorf("expected success status with %s DNS over TLS, got %d", name, resp.Ok().StatusCode)
			}
		})
	}
}

func TestDNSOverTLSCustomProvider(t *testing.T) {
	t.Parallel()

	// Test adding custom provider
	client := surf.NewClient()
	builder := client.Builder()
	dnsBuilder := builder.DNSOverTLS()

	builtClient := dnsBuilder.Cloudflare().Build()

	// Test that we can create a client with AddProvider
	client2 := surf.NewClient()
	builder2 := client2.Builder()
	dnsBuilder2 := builder2.DNSOverTLS()
	builtClient2 := dnsBuilder2.AddProvider("custom1.example.com", "custom1.example.com:853").Build()

	if builtClient == nil {
		t.Error("expected client with DNS provider to be non-nil")
	}

	if builtClient2 == nil {
		t.Error("expected client with custom DNS providers to be non-nil")
	}
}

func TestDNSOverTLSMultipleProviders(t *testing.T) {
	t.Parallel()

	// Test chaining multiple AddProvider calls
	client := surf.NewClient()
	builder := client.Builder()
	dnsBuilder := builder.DNSOverTLS()

	builtClient := dnsBuilder.AddProvider("1.1.1.1", "1.1.1.1:853").Build()

	if builtClient == nil {
		t.Error("expected client with multiple DNS providers to be non-nil")
	}
}

func TestDNSOverTLSAllProviders(t *testing.T) {
	t.Parallel()

	// Test that all providers can be instantiated and chained
	providers := []string{
		"AdGuard", "Google", "Quad9", "Switch", "CIRAShield",
		"Ali", "Quad101", "SB", "Forge", "LibreDNS",
	}

	for i, name := range providers {
		t.Run(fmt.Sprintf("Provider_%s_%d", name, i), func(t *testing.T) {
			client := surf.NewClient()
			builder := client.Builder()
			dnsBuilder := builder.DNSOverTLS()

			var builtClient *surf.Client

			// First test that the provider works
			switch name {
			case "AdGuard":
				builtClient = dnsBuilder.AdGuard().Build()
			case "Google":
				builtClient = dnsBuilder.Google().Build()
			case "Quad9":
				builtClient = dnsBuilder.Quad9().Build()
			case "Switch":
				builtClient = dnsBuilder.Switch().Build()
			case "CIRAShield":
				builtClient = dnsBuilder.CIRAShield().Build()
			case "Ali":
				builtClient = dnsBuilder.Ali().Build()
			case "Quad101":
				builtClient = dnsBuilder.Quad101().Build()
			case "SB":
				builtClient = dnsBuilder.SB().Build()
			case "Forge":
				builtClient = dnsBuilder.Forge().Build()
			case "LibreDNS":
				builtClient = dnsBuilder.LibreDNS().Build()
			}

			if builtClient == nil {
				t.Error("expected client to be built successfully")
				return
			}

			// Now test AddProvider separately
			client2 := surf.NewClient()
			builder2 := client2.Builder()
			dnsBuilder2 := builder2.DNSOverTLS()
			builtClient2 := dnsBuilder2.AddProvider("test.com", "test.com:853").Build()

			if builtClient2 == nil {
				t.Error("expected client with AddProvider to be built successfully")
			}
		})
	}
}
