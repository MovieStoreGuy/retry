package client

import (
	"errors"
	"net/http"

	"github.com/MovieStoreGuy/retry/http/transport"
)

// New will add transport decorator to in use transport of the client
// If the transport is not already set, http.DefaultTransport is used
func New(c *http.Client, attempts int, opts ...transport.Option) (*http.Client, error) {
	if c == nil {
		return nil, errors.New(`http client is nil`)
	}
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}
	tr, err := transport.New(c.Transport, attempts, opts...)
	if err != nil {
		return nil, err
	}
	c.Transport = tr
	return c, nil
}

// Default uses an unconfigured http.Client with the http.DefaultTransport set.
func Default(attempts int, opts ...transport.Option) (*http.Client, error) {
	return New(&http.Client{}, attempts, opts...)
}
