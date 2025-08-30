package surf

import "fmt"

type (
	ErrWebSocketUpgrade struct{ Msg string }
	ErrUserAgentType    struct{ Msg string }
	Err101ResponseCode  struct{ Msg string }
)

func (e *ErrWebSocketUpgrade) Error() string {
	return fmt.Sprintf("%s received an unexpected response, switching protocols to WebSocket", e.Msg)
}

func (e *ErrUserAgentType) Error() string {
	return fmt.Sprintf("unsupported user agent type: %s", e.Msg)
}

func (e *Err101ResponseCode) Error() string {
	return fmt.Sprintf("%s received a 101 response status code", e.Msg)
}
