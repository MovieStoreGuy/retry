package client_test

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/MovieStoreGuy/retry"
	"github.com/MovieStoreGuy/retry/http/client"
	"github.com/MovieStoreGuy/retry/http/status"
	"github.com/MovieStoreGuy/retry/http/transport"
)

func ExampleDefault() {
	c, err := client.Default(6, // Sets the static limit of allowed attempts
		transport.WithRetryOnStatusCode(http.StatusConflict, http.StatusTooManyRequests),
		transport.WithRetryOnStatusGroup(status.Group5xx),
		transport.WithRetryOptions(
			retry.WithExponentialBackoff(200*time.Millisecond, 1.4),
			retry.WithJitter(80*time.Millisecond),
		),
	)
	if err != nil {
		panic(err)
	}
	resp, err := c.Get(`https://golang.org`)
	if err != nil {
		panic(err)
	}
	// handle response
	_ = resp
}

func ExampleNew() {
	t := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Disable HTTP/2 support
	}

	c, err := client.New(&http.Client{Transport: t}, 8,
		transport.WithRetryOnStatusCode(http.StatusTooManyRequests),
		transport.WithRetryOnStatusGroup(status.Group5xx),
	)

	if err != nil {
		panic(err)
	}

	if _, err := c.Get(`https://site-with-rate-limits.com`); err != nil {
		panic(err)
	}
}
