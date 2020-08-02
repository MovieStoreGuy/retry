package transport

import (
	"errors"
	"net/http"

	"github.com/MovieStoreGuy/retry"
)

type retryTransport struct {
	attempts int
	try      retry.Retryer
	wrapped  http.RoundTripper

	checks []func(*http.Response) error
}

type config struct {
	rtOpts []retry.Option
	checks []func(*http.Response) error
}

var _ http.RoundTripper = &retryTransport{}

// New will wrap the passed RoundTripper with a retry abstraction that will retry on any recoverable failures
// by default. Any options are evaluated in the order they are passed into this method, changing the order can
// also lead to differences in runtime behaviour.
func New(rt http.RoundTripper, attempts int, opts ...Option) (http.RoundTripper, error) {
	if rt == nil {
		return nil, errors.New(`roundtripper is nil`)
	}
	if attempts < 1 {
		return nil, errors.New(`attempts must be positive`)
	}

	cf := &config{}
	for _, opt := range opts {
		if err := opt(cf); err != nil {
			return nil, err
		}
	}

	try, err := retry.New(cf.rtOpts...)
	if err != nil {
		return nil, err
	}

	return &retryTransport{
		attempts: attempts,
		try:      try,
		wrapped:  rt,
		checks:   cf.checks,
	}, nil
}

// Default uses the default http.DefaultTransport as the RoundTripper to be used
// with the retry ability applied to it.
func Default(attempts int, opts ...Option) (http.RoundTripper, error) {
	return New(http.DefaultTransport, attempts, opts...)
}

// Must wraps an evaluated New method and will panic if an error has been returned.
func Must(rt http.RoundTripper, err error) http.RoundTripper {
	if err != nil {
		panic(err)
	}
	return rt
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New(`request is nil`)
	}
	var resp *http.Response
	err := rt.try.AttemptWithContext(req.Context(), rt.attempts, func() error {
		r, err := rt.wrapped.RoundTrip(req)
		switch {
		case r == nil && err != nil:
			return retry.NoRecover(err.Error())
		case err != nil:
			return err
		}
		for _, check := range rt.checks {
			if err = check(r); err != nil {
				return err
			}
		}
		resp = r
		return nil
	})
	return resp, err
}
