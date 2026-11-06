package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/paveltessman/jao/protocol/awg"
)

func TestErrorMessage(t *testing.T) {

	cases := []struct {
		desc string
		err  error
		want string
	}{
		{desc: "nil error", err: nil, want: ""},
		{desc: "known not internal error", err: ErrBadPathValue, want: "bad path value"},
		{desc: "known internal error", err: awg.ErrBadEnv, want: "internal server error"},
		{desc: "unknown internal error", err: errors.New(""), want: "internal server error"},
	}

	for _, c := range cases {
		msg := errorMessage(c.err)
		if msg != c.want {
			t.Errorf("%s: got=%q, want=%q", c.desc, msg, c.want)
		}
	}
}

func TestStatusCode(t *testing.T) {

	cases := []struct {
		desc string
		err  error
		want int
	}{
		{desc: "nil error", err: nil, want: http.StatusOK},
		{desc: "unknown internal error", err: errors.New(""), want: http.StatusInternalServerError},
	}

	for _, c := range cases {
		code := statusCode(c.err)
		if code != c.want {
			t.Errorf("%s: got=%d, want=%d", c.desc, code, c.want)
		}
	}
}

func TestErrorSpec(t *testing.T) {

	knownSentinelErr := errors.New("")
	knownSentinelSpec := ErrorSpec{internal: false, statusCode: 404}
	wrappedErr := fmt.Errorf("%w: something", knownSentinelErr)
	unknownErr := errors.New("")

	errorSpecs = map[error]ErrorSpec{
		knownSentinelErr: knownSentinelSpec,
	}

	cases := []struct {
		desc  string
		err   error
		spec  ErrorSpec
		found bool
	}{
		{"known sentinel", knownSentinelErr, knownSentinelSpec, true},
		{"wrapped in known sentinel", wrappedErr, knownSentinelSpec, true},
		{"unknown error", unknownErr, ErrorSpec{}, false},
	}

	for _, c := range cases {
		spec, found := errorSpec(c.err)

		if found != c.found {
			t.Fatalf("%s: found=%t, want=%t", c.desc, found, c.found)
		}

		if found && (spec != c.spec) {
			t.Errorf("%s: spec=%v, want=%v", c.desc, spec, c.spec)

		}
	}
}
