package surf

import (
	"net"
	"net/url"
	"time"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/surf/header"
)

// Response represents a custom response structure.
type Response struct {
	*Client                      // Client is the associated client for the response.
	remoteAddr    net.Addr       // Remote network address.
	URL           *url.URL       // URL of the response.
	response      *http.Response // Underlying http.Response.
	Body          *Body          // Response body.
	request       *Request       // Corresponding request.
	Headers       Headers        // Response headers.
	UserAgent     g.String       // User agent string.
	Proto         g.String       // HTTP protocol version.
	Cookies       Cookies        // Response cookies.
	Time          time.Duration  // Total time taken for the response.
	ContentLength int64          // Length of the response content.
	StatusCode    StatusCode     // HTTP status code.
	Attempts      int            // Number of attempts made.
}

// GetResponse returns the underlying http.Response of the custom response.
func (resp Response) GetResponse() *http.Response { return resp.response }

// Referer returns the referer of the response.
func (resp Response) Referer() g.String { return g.String(resp.response.Request.Referer()) }

// Location returns the location of the response.
func (resp Response) Location() g.String { return resp.Headers.Get(header.LOCATION) }

// GetCookies returns the cookies from the response for the given URL.
func (resp Response) GetCookies(rawURL g.String) []*http.Cookie { return resp.getCookies(rawURL) }

// RemoteAddress returns the remote address of the response.
func (resp Response) RemoteAddress() net.Addr { return resp.remoteAddr }

// SetCookies sets cookies for the given URL in the response.
func (resp *Response) SetCookies(rawURL g.String, cookies []*http.Cookie) error {
	return resp.setCookies(rawURL, cookies)
}

// TLSGrabber returns a tlsData struct containing information about the TLS connection if it
// exists.
func (resp Response) TLSGrabber() *TLSData {
	if resp.response.TLS != nil {
		return tlsGrabber(resp.response.TLS)
	}

	return nil
}
