package transport

import (
	"errors"
	"net/http"

	"github.com/MovieStoreGuy/retry"
)

type retryTransport struct {
	attempts int
	wrapped  http.RoundTripper

	opts   []retry.Option
	checks []func(*http.Response) error
}

type config struct {
	rtOpts []retry.Option
	checks []func(*http.Response) error
	table  map[int]struct{}
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

	cf := &config{table: make(map[int]struct{})}
	for _, opt := range opts {
		if err := opt(cf); err != nil {
			return nil, err
		}
	}
	// Validate that the options will work early on
	if _, err := retry.New(cf.rtOpts...); err != nil {
		return nil, err
	}

	return &retryTransport{
		attempts: attempts,
		wrapped:  rt,
		opts:     cf.rtOpts,
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
	// Creating a retry per RoundTrip to avoid issues with concurrent requests being made by the client
	err := retry.Must(rt.opts...).AttemptWithContext(req.Context(), rt.attempts, func() error {
		r, err := rt.wrapped.RoundTrip(req)
		if err != nil {
			return retry.AbortedRetries(err.Error())
		}

		// Checking the response returned and if we fail any of the checks
		// return the error without setting response to compile with the interface
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
