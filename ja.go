package surf

import (
	"context"
	"math/rand"
	"net"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/surf/pkg/connectproxy"

	utls "github.com/refraction-networking/utls"
)

// https://lwthiker.com/networks/2022/06/17/tls-fingerprinting.html
type JA struct {
	spec    utls.ClientHelloSpec
	id      utls.ClientHelloID
	builder *Builder
}

// SetHelloID sets a ClientHelloID for the TLS connection.
//
// The provided ClientHelloID is used to customize the TLS handshake. This
// should be a valid identifier that can be mapped to a specific ClientHelloSpec.
//
// It returns a pointer to the Options struct for method chaining. This allows
// additional configuration methods to be called on the result.
//
// Example usage:
//
//	JA().SetHelloID(utls.HelloChrome_Auto)
func (j *JA) SetHelloID(id utls.ClientHelloID) *Builder {
	j.id = id
	return j.build()
}

// SetHelloSpec sets a custom ClientHelloSpec for the TLS connection.
//
// This method allows you to set a custom ClientHelloSpec to be used during the TLS handshake.
// The provided spec should be a valid ClientHelloSpec.
//
// It returns a pointer to the Options struct for method chaining. This allows
// additional configuration methods to be called on the result.
//
// Example usage:
//
//	JA().SetHelloSpec(spec)
func (j *JA) SetHelloSpec(spec utls.ClientHelloSpec) *Builder {
	j.spec = spec
	return j.build()
}

func (j *JA) build() *Builder {
	return j.builder.addCliMW(func(c *Client) {
		if !j.builder.singleton {
			j.builder.addRespMW(closeIdleConnectionsMW, 0)
		}

		if j.builder.proxy != nil {
			var p string
			switch v := j.builder.proxy.(type) {
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

			if p != "" {
				if dialer, err := connectproxy.NewDialer(p); err != nil {
					c.GetTransport().(*http.Transport).DialContext = func(context.Context, string, string) (net.Conn, error) { return nil, err }
				} else {
					c.GetTransport().(*http.Transport).DialContext = dialer.DialContext
				}
			}
		}

		c.GetClient().Transport = newRoundTripper(j, c.GetTransport())
	}, 0)
}

// getSpec determines the ClientHelloSpec to be used for the TLS connection.
//
// The ClientHelloSpec is selected based on the following order of precedence:
// 1. If a custom ClientHelloID is set (via SetHelloID), it attempts to convert this ID to a ClientHelloSpec.
// 2. If none of the above conditions are met, it returns the currently set ClientHelloSpec.
//
// This method returns the selected ClientHelloSpec along with an error value. If an error occurs
// during conversion, it returns the error.
func (j *JA) getSpec() g.Result[utls.ClientHelloSpec] {
	if !j.id.IsSet() {
		return g.ResultOf(utls.UTLSIdToSpec(j.id))
	}

	return g.Ok(j.spec)
}

func (j *JA) Android() *Builder          { return j.SetHelloID(utls.HelloAndroid_11_OkHttp) }
func (j *JA) Chrome() *Builder           { return j.SetHelloID(utls.HelloChrome_Auto) }
func (j *JA) Chrome58() *Builder         { return j.SetHelloID(utls.HelloChrome_58) }
func (j *JA) Chrome62() *Builder         { return j.SetHelloID(utls.HelloChrome_62) }
func (j *JA) Chrome70() *Builder         { return j.SetHelloID(utls.HelloChrome_70) }
func (j *JA) Chrome72() *Builder         { return j.SetHelloID(utls.HelloChrome_72) }
func (j *JA) Chrome83() *Builder         { return j.SetHelloID(utls.HelloChrome_83) }
func (j *JA) Chrome87() *Builder         { return j.SetHelloID(utls.HelloChrome_87) }
func (j *JA) Chrome96() *Builder         { return j.SetHelloID(utls.HelloChrome_96) }
func (j *JA) Chrome100() *Builder        { return j.SetHelloID(utls.HelloChrome_100) }
func (j *JA) Chrome102() *Builder        { return j.SetHelloID(utls.HelloChrome_102) }
func (j *JA) Chrome106() *Builder        { return j.SetHelloID(utls.HelloChrome_106_Shuffle) }
func (j *JA) Chrome120() *Builder        { return j.SetHelloID(utls.HelloChrome_120) }
func (j *JA) Chrome120PQ() *Builder      { return j.SetHelloID(utls.HelloChrome_120_PQ) }
func (j *JA) Chrome131() *Builder        { return j.SetHelloID(utls.HelloChrome_131) }
func (j *JA) Edge() *Builder             { return j.SetHelloID(utls.HelloEdge_85) }
func (j *JA) Edge85() *Builder           { return j.SetHelloID(utls.HelloEdge_85) }
func (j *JA) Edge106() *Builder          { return j.SetHelloID(utls.HelloEdge_106) }
func (j *JA) Firefox() *Builder          { return j.SetHelloID(utls.HelloFirefox_Auto) }
func (j *JA) Firefox55() *Builder        { return j.SetHelloID(utls.HelloFirefox_55) }
func (j *JA) Firefox56() *Builder        { return j.SetHelloID(utls.HelloFirefox_56) }
func (j *JA) Firefox63() *Builder        { return j.SetHelloID(utls.HelloFirefox_63) }
func (j *JA) Firefox65() *Builder        { return j.SetHelloID(utls.HelloFirefox_65) }
func (j *JA) Firefox99() *Builder        { return j.SetHelloID(utls.HelloFirefox_99) }
func (j *JA) Firefox102() *Builder       { return j.SetHelloID(utls.HelloFirefox_102) }
func (j *JA) Firefox105() *Builder       { return j.SetHelloID(utls.HelloFirefox_105) }
func (j *JA) Firefox120() *Builder       { return j.SetHelloID(utls.HelloFirefox_120) }
func (j *JA) Firefox131() *Builder       { return j.SetHelloID(utls.HelloFirefox_120) }
func (j *JA) IOS() *Builder              { return j.SetHelloID(utls.HelloIOS_Auto) }
func (j *JA) IOS11() *Builder            { return j.SetHelloID(utls.HelloIOS_11_1) }
func (j *JA) IOS12() *Builder            { return j.SetHelloID(utls.HelloIOS_12_1) }
func (j *JA) IOS13() *Builder            { return j.SetHelloID(utls.HelloIOS_13) }
func (j *JA) IOS14() *Builder            { return j.SetHelloID(utls.HelloIOS_14) }
func (j *JA) Randomized() *Builder       { return j.SetHelloID(utls.HelloRandomized) }
func (j *JA) RandomizedALPN() *Builder   { return j.SetHelloID(utls.HelloRandomizedALPN) }
func (j *JA) RandomizedNoALPN() *Builder { return j.SetHelloID(utls.HelloRandomizedNoALPN) }
func (j *JA) Safari() *Builder           { return j.SetHelloID(utls.HelloSafari_Auto) }
