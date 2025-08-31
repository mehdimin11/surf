package surf

import (
	"context"
	"crypto/tls"
	"math/rand"
	"net"
	_http "net/http"
	"net/url"
	"strings"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http3"
	"github.com/quic-go/quic-go"
	uquic "github.com/refraction-networking/uquic"
	"github.com/wzshiming/socks5"
)

// HTTP3Settings represents HTTP/3 settings with uQUIC fingerprinting support.
// https://github.com/refraction-networking/uquic
type HTTP3Settings struct {
	builder  *Builder
	quicID   *uquic.QUICID
	quicSpec *uquic.QUICSpec
}

// Chrome configures HTTP/3 settings to mimic Chrome browser.
func (h *HTTP3Settings) Chrome() *HTTP3Settings {
	h.quicID = &uquic.QUICChrome_115
	return h
}

// Firefox configures HTTP/3 settings to mimic Firefox browser.
func (h *HTTP3Settings) Firefox() *HTTP3Settings {
	h.quicID = &uquic.QUICFirefox_116
	return h
}

// SetQUICID sets a custom QUIC ID for fingerprinting.
func (h *HTTP3Settings) SetQUICID(quicID uquic.QUICID) *HTTP3Settings {
	h.quicID = &quicID
	return h
}

// SetQUICSpec sets a custom QUIC spec for advanced fingerprinting.
func (h *HTTP3Settings) SetQUICSpec(quicSpec uquic.QUICSpec) *HTTP3Settings {
	h.quicSpec = &quicSpec
	return h
}

// getQUICSpec returns the QUIC spec either from custom spec or by converting QUICID.
func (h *HTTP3Settings) getQUICSpec() g.Option[uquic.QUICSpec] {
	if h.quicSpec != nil {
		return g.Some(*h.quicSpec)
	}

	if h.quicID != nil {
		spec, err := uquic.QUICID2Spec(*h.quicID)
		if err == nil {
			return g.Some(spec)
		}
	}

	return g.None[uquic.QUICSpec]()
}

// Set applies the accumulated HTTP/3 settings.
// It configures the uQUIC transport for the surf client.
func (h *HTTP3Settings) Set() *Builder {
	if h.builder.forseHTTP1 {
		return h.builder
	}

	return h.builder.addCliMW(func(c *Client) {
		quicSpec := h.getQUICSpec()
		if quicSpec.IsNone() {
			return
		}

		var proxyURL string
		var useSOCKS5 bool

		// Check if proxy is SOCKS5 (supports UDP for QUIC)
		if h.builder.proxy != nil {
			proxyURL, useSOCKS5 = isSOCKS5Proxy(h.builder.proxy)
		}

		// HTTP/3 is incompatible with non-SOCKS5 proxies - fallback to HTTP/2
		if h.builder.proxy != nil && !useSOCKS5 {
			return
		}

		transport := &uquicTransport{
			quicSpec:         quicSpec.Some(),
			tlsConfig:        c.tlsConfig.Clone(),
			dialer:           c.GetDialer(),
			proxyURL:         proxyURL,
			cachedTransports: g.NewMapSafe[string, *http3.Transport](),
		}

		c.GetClient().Transport = transport
		c.transport = transport
	}, 0)
}

// uquicTransport implements http.RoundTripper using uQUIC fingerprinting with quic-go HTTP/3
type uquicTransport struct {
	quicSpec         uquic.QUICSpec
	tlsConfig        *tls.Config
	dialer           *net.Dialer
	proxyURL         string // SOCKS5 proxy URL if available
	cachedTransports *g.MapSafe[string, *http3.Transport]
}

func (ut *uquicTransport) CloseIdleConnections() {
	if ut.cachedTransports == nil {
		return
	}

	for k, h3 := range ut.cachedTransports.Iter() {
		h3.CloseIdleConnections()
		ut.cachedTransports.Delete(k)
	}
}

// address builds host:port (defaults 443 if port missing)
func (ut *uquicTransport) address(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}

	return net.JoinHostPort(req.URL.Host, "443")
}

// createH3 returns per-address cached http3.Transport with proper Dial & SNI
func (ut *uquicTransport) createH3(req *http.Request, addr string) *http3.Transport {
	if ut.cachedTransports == nil {
		ut.cachedTransports = g.NewMapSafe[string, *http3.Transport]()
	}

	if tr := ut.cachedTransports.Get(addr); tr.IsSome() {
		return tr.Some()
	}

	h3 := &http3.Transport{TLSClientConfig: ut.tlsConfig}

	if (ut.dialer != nil && ut.dialer.Resolver != nil) || ut.proxyURL != "" {
		hostname := req.URL.Hostname()
		h3.Dial = func(ctx context.Context, quicAddr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			if tlsCfg == nil {
				tlsCfg = new(tls.Config)
			}

			if tlsCfg.ServerName == "" {
				if hn := hostname; hn != "" {
					clone := tlsCfg.Clone()
					clone.ServerName = hn
					tlsCfg = clone
				}
			}

			if ut.proxyURL != "" {
				return ut.dialSOCKS5(ctx, quicAddr, tlsCfg, cfg)
			}

			return ut.dialDNS(ctx, quicAddr, tlsCfg, cfg)
		}
	}

	ut.cachedTransports.Set(addr, h3)
	return h3
}

// resolveAddress resolves the host using custom DNS resolver if available.
// Returns the original address if no custom DNS is configured.
func (ut *uquicTransport) resolveAddress(ctx context.Context, address string) (string, error) {
	if ut.dialer == nil || ut.dialer.Resolver == nil {
		return address, nil
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	// Resolve using custom DNS
	ips, err := ut.dialer.Resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", &net.DNSError{Err: "no such host", Name: host}
	}

	// Use the first resolved IP
	return net.JoinHostPort(ips[0].IP.String(), port), nil
}

// dialSOCKS5 establishes a QUIC connection through a SOCKS5 proxy.
// Uses custom DNS resolver if available before connecting through proxy.
func (ut *uquicTransport) dialSOCKS5(
	ctx context.Context,
	address string,
	tlsConfig *tls.Config,
	cfg *quic.Config,
) (*quic.Conn, error) {
	// Resolve address using custom DNS if available
	resolvedAddress, err := ut.resolveAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	// Parse proxy URL
	proxyURL, err := url.Parse(ut.proxyURL)
	if err != nil {
		return nil, err
	}

	// Create SOCKS5 dialer with UDP support
	dialer, err := socks5.NewDialer(proxyURL.String())
	if err != nil {
		return nil, err
	}

	// Dial through SOCKS5 proxy using UDP
	conn, err := dialer.DialContext(ctx, "udp", resolvedAddress)
	if err != nil {
		return nil, err
	}

	// Create remote address for QUIC
	remoteAddr, err := net.ResolveUDPAddr("udp", resolvedAddress)
	if err != nil {
		return nil, err
	}

	// Wrap connection for QUIC compatibility
	packetConn := newQUICPacketConn(conn, remoteAddr)

	// Establish QUIC connection through the proxy
	return quic.Dial(ctx, packetConn, remoteAddr, tlsConfig, cfg)
}

// dialDNS establishes a QUIC connection using custom DNS resolver.
func (ut *uquicTransport) dialDNS(
	ctx context.Context,
	address string,
	tlsConfig *tls.Config,
	cfg *quic.Config,
) (*quic.Conn, error) {
	// Resolve address using custom DNS
	resolvedAddress, err := ut.resolveAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(resolvedAddress)
	if err != nil {
		return nil, err
	}

	// Create UDP connection
	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	// Parse port
	portNum, _ := net.LookupPort("udp", port)

	// Create target address
	targetAddr := &net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: portNum,
	}

	return quic.Dial(ctx, udpConn, targetAddr, tlsConfig, cfg)
}

// RoundTrip implements the http.RoundTripper interface with HTTP/3 support
func (ut *uquicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "" {
		clone := *req.URL
		clone.Scheme = "https"
		req.URL = &clone
	}

	_req := &_http.Request{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           _http.Header(req.Header),
		Body:             req.Body,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Close:            req.Close,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          _http.Header(req.Trailer),
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       "",
		TLS:              req.TLS,
		Response:         nil,
		GetBody:          req.GetBody,
		Pattern:          req.Pattern,
		Cancel:           req.Cancel, // deprecated but kept for compatibility
	}

	_req = _req.WithContext(req.Context())

	addr := ut.address(req)
	h3 := ut.createH3(req, addr)

	_resp, err := h3.RoundTrip(_req)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Status:           _resp.Status,
		StatusCode:       _resp.StatusCode,
		Proto:            _resp.Proto,
		ProtoMajor:       _resp.ProtoMajor,
		ProtoMinor:       _resp.ProtoMinor,
		Header:           http.Header(_resp.Header),
		Body:             _resp.Body,
		ContentLength:    _resp.ContentLength,
		Close:            _resp.Close,
		Uncompressed:     _resp.Uncompressed,
		Trailer:          http.Header(_resp.Trailer),
		Request:          req,
		TLS:              _resp.TLS,
		TransferEncoding: _resp.TransferEncoding,
	}, nil
}

// isSOCKS5Proxy checks if the given proxy configuration is a SOCKS5 proxy supporting UDP.
// Returns the proxy URL and true if it's a SOCKS5 proxy, empty string and false otherwise.
func isSOCKS5Proxy(proxy any) (string, bool) {
	var p string
	switch v := proxy.(type) {
	case string:
		p = v
	case g.String:
		p = v.Std()
	case []string:
		p = v[rand.Intn(len(v))]
	case g.Slice[string]:
		p = v.Random()
	case g.Slice[g.String]:
		p = v.Random().Std()
	}

	if p == "" {
		return "", false
	}

	parsedURL, err := url.Parse(p)
	if err != nil {
		return "", false
	}

	scheme := strings.ToLower(parsedURL.Scheme)

	return p, scheme == "socks5" || scheme == "socks5h"
}
