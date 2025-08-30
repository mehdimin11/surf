package surf

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/surf/header"
	"github.com/enetx/surf/internal/drainbody"
)

// Request is a struct that holds information about an HTTP request.
type Request struct {
	request     *http.Request   // The underlying http.Request.
	cli         *Client         // The associated client for the request.
	werr        *error          // An error encountered during writing.
	err         error           // A general error associated with the request.
	remoteAddr  net.Addr        // Remote network address.
	body        io.ReadCloser   // Request body.
	headersKeys g.Slice[string] // Order headers.
}

// GetRequest returns the underlying http.Request of the custom request.
func (req *Request) GetRequest() *http.Request { return req.request }

// Do performs the HTTP request and returns a Response object or an error if the request failed.
func (req *Request) Do() g.Result[*Response] {
	if req.err != nil {
		return g.Err[*Response](req.err)
	}

	if err := req.cli.applyReqMW(req); err != nil {
		return g.Err[*Response](err)
	}

	if req.request.Method != http.MethodHead {
		req.body, req.request.Body, req.err = drainbody.DrainBody(req.request.Body)
		if req.err != nil {
			return g.Err[*Response](req.err)
		}
	}

	var (
		resp     *http.Response
		attempts int
		err      error
	)

	start := time.Now()
	cli := req.cli.cli

	builder := req.cli.builder

retry:
	resp, err = cli.Do(req.request)
	if err != nil {
		return g.Err[*Response](err)
	}

	if builder != nil && builder.retryMax != 0 && attempts < builder.retryMax && builder.retryCodes.NotEmpty() &&
		builder.retryCodes.Contains(resp.StatusCode) {
		attempts++

		time.Sleep(builder.retryWait)
		goto retry
	}

	if req.werr != nil && *req.werr != nil {
		return g.Err[*Response](*req.werr)
	}

	response := &Response{
		Attempts:      attempts,
		Time:          time.Since(start),
		Client:        req.cli,
		ContentLength: resp.ContentLength,
		Cookies:       resp.Cookies(),
		Headers:       Headers(resp.Header),
		Proto:         g.String(resp.Proto),
		StatusCode:    StatusCode(resp.StatusCode),
		URL:           resp.Request.URL,
		UserAgent:     g.String(req.request.UserAgent()),
		remoteAddr:    req.remoteAddr,
		request:       req,
		response:      resp,
	}

	if req.request.Method != http.MethodHead {
		response.Body = new(Body)
		response.Body.Reader = resp.Body
		response.Body.cache = builder != nil && builder.cacheBody
		response.Body.contentType = resp.Header.Get(header.CONTENT_TYPE)
		response.Body.limit = -1
	}

	if err := req.cli.applyRespMW(response); err != nil {
		return g.Err[*Response](err)
	}

	return g.Ok(response)
}

// WithContext associates the provided context with the request.
func (req *Request) WithContext(ctx context.Context) *Request {
	if ctx != nil {
		req.request = req.request.WithContext(ctx)
	}

	return req
}

// AddCookies adds cookies to the request.
func (req *Request) AddCookies(cookies ...*http.Cookie) *Request {
	for _, cookie := range cookies {
		req.request.AddCookie(cookie)
	}

	return req
}

// SetHeaders sets headers for the request, replacing existing ones with the same name.
func (req *Request) SetHeaders(headers ...any) *Request {
	if req.request == nil || headers == nil {
		return req
	}

	applyHeaders(req.request, headers, req, func(h http.Header, k, v string) { h.Set(k, v) })

	return req
}

// AddHeaders adds headers to the request, appending to any existing headers with the same name.
func (req *Request) AddHeaders(headers ...any) *Request {
	if req.request == nil || headers == nil {
		return req
	}

	applyHeaders(req.request, headers, req, func(h http.Header, k, v string) { h.Add(k, v) })

	return req
}

func applyHeaders(r *http.Request, rawHeaders []any, req *Request, setOrAdd func(h http.Header, key, value string)) {
	if len(rawHeaders) >= 2 {
		var key, value string

		switch k := rawHeaders[0].(type) {
		case string:
			key = k
		case g.String:
			key = k.Std()
		default:
			panic(fmt.Sprintf("unsupported key type: expected 'string' or 'String', got %T", rawHeaders[0]))
		}

		switch v := rawHeaders[1].(type) {
		case string:
			value = v
		case g.String:
			value = v.Std()
		default:
			panic(fmt.Sprintf("unsupported value type: expected 'string' or 'String', got %T", rawHeaders[1]))
		}

		setOrAdd(r.Header, key, value)
		return
	}

	switch h := rawHeaders[0].(type) {
	case http.Header:
		for key, values := range h {
			for _, value := range values {
				setOrAdd(r.Header, key, value)
			}
		}
	case Headers:
		for key, values := range h {
			for _, value := range values {
				setOrAdd(r.Header, key, value)
			}
		}
	case map[string]string:
		for key, value := range h {
			setOrAdd(r.Header, key, value)
		}
	case map[g.String]g.String:
		for key, value := range h {
			setOrAdd(r.Header, key.Std(), value.Std())
		}
	case g.Map[string, string]:
		for key, value := range h {
			setOrAdd(r.Header, key, value)
		}
	case g.Map[g.String, g.String]:
		for key, value := range h {
			setOrAdd(r.Header, key.Std(), value.Std())
		}
	case g.MapOrd[string, string]:
		updated := updateRequestHeaderOrder(req, h)
		updated.Iter().ForEach(func(key, value string) { setOrAdd(r.Header, key, value) })
	case g.MapOrd[g.String, g.String]:
		updated := updateRequestHeaderOrder(req, h)
		updated.Iter().ForEach(func(key, value g.String) { setOrAdd(r.Header, key.Std(), value.Std()) })
	default:
		panic(fmt.Sprintf("unsupported headers type: expected 'http.Header', 'surf.Headers', 'map[~string]~string', 'Map[~string, ~string]', or 'MapOrd[~string, ~string]', got %T", rawHeaders[0]))
	}
}

func updateRequestHeaderOrder[T ~string](r *Request, h g.MapOrd[T, T]) g.MapOrd[T, T] {
	r.headersKeys.Push(h.Iter().
		Keys().
		Map(func(s T) T { return T(g.String(s).Lower()) }).
		Collect().
		ToStringSlice()...)

	headers, pheaders := r.headersKeys.Iter().Partition(func(v string) bool { return v[0] != ':' })

	if headers.NotEmpty() {
		r.request.Header[http.HeaderOrderKey] = headers
	}

	if pheaders.NotEmpty() {
		r.request.Header[http.PHeaderOrderKey] = pheaders
	}

	return h.Iter().
		Filter(func(header, data T) bool { return header[0] != ':' && len(data) != 0 }).
		Collect()
}
