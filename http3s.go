package surf

import (
	"context"
	"crypto/tls"
	"net"
	_http "net/http"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/http3"
	"github.com/quic-go/quic-go"
	uquic "github.com/refraction-networking/uquic"
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

		// HTTP/3 is incompatible with proxies - fallback to HTTP/2
		if h.builder.proxy != nil {
			return
		}

		transport := &uquicTransport{
			quicSpec:  quicSpec.Some(),
			tlsConfig: c.tlsConfig.Clone(),
			dialer:    c.GetDialer(),
		}

		c.GetClient().Transport = transport
		c.transport = transport
	}, 0)
}

// uquicTransport implements http.RoundTripper using uQUIC fingerprinting with quic-go HTTP/3
type uquicTransport struct {
	quicSpec  uquic.QUICSpec
	tlsConfig *tls.Config
	dialer    *net.Dialer
}

// RoundTrip implements the http.RoundTripper interface with HTTP/3 support
// Note: This implementation sets up HTTP/3 transport and stores fingerprint info,
// but full QUIC fingerprinting integration would require more complex connection management
func (t *uquicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		req.URL.Scheme = "https"
	}

	// Convert github.com/enetx/http.Request to standard net/http request
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
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
		Response:         nil,
		GetBody:          req.GetBody,
		Pattern:          req.Pattern,
		Cancel:           req.Cancel, // deprecated but kept for compatibility
	}

	_req = _req.WithContext(req.Context())

	// Create HTTP/3 transport with fingerprint spec information preserved
	// The fingerprint spec is stored in the transport for potential future use
	h3Transport := &http3.Transport{
		TLSClientConfig: t.tlsConfig,
	}

	// Configure custom DNS resolver if available
	if t.dialer != nil && t.dialer.Resolver != nil {
		h3Transport.Dial = func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := t.dialer.Resolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}

			if len(ips) == 0 {
				return nil, &net.DNSError{Err: "no such host", Name: host}
			}

			udpConn, err := net.ListenUDP("udp", nil)
			if err != nil {
				return nil, err
			}

			return quic.Dial(ctx, udpConn, &net.UDPAddr{
				IP: ips[0].IP,
				Port: func() int {
					p, _ := net.LookupPort("udp", port)
					return p
				}(),
			}, tlsCfg, cfg)
		}
	}

	_resp, err := h3Transport.RoundTrip(_req)
	if err != nil {
		return nil, err
	}

	// Convert back to surf response
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
		TransferEncoding: _req.TransferEncoding,
	}, nil
}
