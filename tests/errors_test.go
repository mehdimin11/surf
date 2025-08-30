package surf_test

import (
	"testing"

	"github.com/enetx/surf"
)

func TestErrWebSocketUpgrade(t *testing.T) {
	t.Parallel()

	err := &surf.ErrWebSocketUpgrade{Msg: "client"}
	expected := "client received an unexpected response, switching protocols to WebSocket"

	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestErrUserAgentType(t *testing.T) {
	t.Parallel()

	err := &surf.ErrUserAgentType{Msg: "invalid-type"}
	expected := "unsupported user agent type: invalid-type"

	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestErr101ResponseCode(t *testing.T) {
	t.Parallel()

	err := &surf.Err101ResponseCode{Msg: "client"}
	expected := "client received a 101 response status code"

	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestErrorTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
		want string
	}{
		{
			"ErrWebSocketUpgrade",
			&surf.ErrWebSocketUpgrade{Msg: "client"},
			"client received an unexpected response, switching protocols to WebSocket",
		},
		{
			"ErrUserAgentType",
			&surf.ErrUserAgentType{Msg: "invalid"},
			"unsupported user agent type: invalid",
		},
		{
			"Err101ResponseCode",
			&surf.Err101ResponseCode{Msg: "client"},
			"client received a 101 response status code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.want {
				t.Errorf("expected %q, got %q", tc.want, tc.err.Error())
			}
		})
	}
}

func TestErrorTypesEmpty(t *testing.T) {
	t.Parallel()

	// Test with empty messages
	testCases := []struct {
		name string
		err  error
		want string
	}{
		{
			"ErrWebSocketUpgrade empty",
			&surf.ErrWebSocketUpgrade{Msg: ""},
			" received an unexpected response, switching protocols to WebSocket",
		},
		{
			"ErrUserAgentType empty",
			&surf.ErrUserAgentType{Msg: ""},
			"unsupported user agent type: ",
		},
		{
			"Err101ResponseCode empty",
			&surf.Err101ResponseCode{Msg: ""},
			" received a 101 response status code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.want {
				t.Errorf("expected %q, got %q", tc.want, tc.err.Error())
			}
		})
	}
}
