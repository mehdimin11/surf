package surf

import (
	"errors"
	"fmt"
	"math/rand"
	"net/textproto"

	"github.com/enetx/g"
	"github.com/enetx/http/httptrace"
	"github.com/enetx/surf/header"
)

// default user-agent for surf.
func defaultUserAgentMW(req *Request) error {
	if headers := req.GetRequest().Header; headers.Get(header.USER_AGENT) == "" {
		// Set the default user-agent header.
		headers.Set(header.USER_AGENT, _userAgent)
	}

	return nil
}

// userAgentMW sets the "User-Agent" header for the given Request. The userAgent parameter
// can be a string or a slice of strings. If it is a slice, a random user agent is selected
// from the slice. If the userAgent is not a string or a slice of strings, an error is returned.
// The function updates the request headers with the selected or given user agent.
func userAgentMW(req *Request, userAgent any) error {
	var ua string

	switch v := userAgent.(type) {
	case string:
		ua = v
	case g.String:
		ua = v.Std()
	case []string:
		if len(v) == 0 {
			return &ErrUserAgentType{"cannot select a random user agent from an empty slice"}
		}
		ua = v[rand.Intn(len(v))]
	case g.Slice[string]:
		if v.Empty() {
			return &ErrUserAgentType{"cannot select a random user agent from an empty slice"}
		}
		ua = v.Random()
	case g.Slice[g.String]:
		if v.Empty() {
			return &ErrUserAgentType{"cannot select a random user agent from an empty slice"}
		}
		ua = v.Random().Std()
	default:
		return &ErrUserAgentType{fmt.Sprintf("'%T' %v", v, v)}
	}

	req.GetRequest().Header.Set(header.USER_AGENT, ua)

	return nil
}

// got101ResponseMW configures the request's context to handle 1xx responses.
// It sets up a client trace for capturing 1xx responses and returns any error encountered.
func got101ResponseMW(req *Request) error {
	req.WithContext(httptrace.WithClientTrace(req.GetRequest().Context(),
		&httptrace.ClientTrace{
			Got1xxResponse: func(code int, _ textproto.MIMEHeader) error {
				if code != 101 {
					return nil
				}

				return &Err101ResponseCode{
					fmt.Sprintf(`%s "%s" error:`, req.request.Method, req.request.URL.String()),
				}
			},
		},
	))

	return nil
}

// remoteAddrMW configures the request's context to get the remote address
// of the server if the 'remoteAddrMW' option is enabled.
func remoteAddrMW(req *Request) error {
	req.WithContext(httptrace.WithClientTrace(req.GetRequest().Context(),
		&httptrace.ClientTrace{
			GotConn: func(info httptrace.GotConnInfo) { req.remoteAddr = info.Conn.RemoteAddr() },
		},
	))

	return nil
}

// bearerAuthMW adds a Bearer token to the Authorization header of the given request.
func bearerAuthMW(req *Request, token g.String) error {
	if token.NotEmpty() {
		req.AddHeaders(g.Map[g.String, g.String]{header.AUTHORIZATION: "Bearer " + token})
	}

	return nil
}

// basicAuthMW sets basic authentication for the request based on the client's options.
func basicAuthMW(req *Request, authentication g.String) error {
	if req.GetRequest().Header.Get(header.AUTHORIZATION) != "" {
		return nil
	}

	var username, password g.String

	authentication.Split(":").Collect().Unpack(&username, &password)

	if username == "" || password == "" {
		return errors.New("basic authorization fields cannot be empty")
	}

	req.GetRequest().SetBasicAuth(username.Std(), password.Std())

	return nil
}

// contentTypeMW sets the Content-Type header for the given HTTP request.
func contentTypeMW(req *Request, contentType g.String) error {
	if contentType.Empty() {
		return fmt.Errorf("Content-Type is empty")
	}

	req.SetHeaders(g.Map[g.String, g.String]{header.CONTENT_TYPE: contentType})

	return nil
}
