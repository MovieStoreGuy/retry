package client_test

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"github.com/MovieStoreGuy/retry"
	"github.com/MovieStoreGuy/retry/http/client"
	"github.com/MovieStoreGuy/retry/http/transport"
)

func ExampleDefault_WithOptions() {
	c, err := client.Default(6, // Sets the static limit of allowed attempts
		transport.WithNoRetryOnResponseCodes(http.StatusBadGateway, http.StatusServiceUnavailable), // Abort if the response matches one of these
		transport.WithRetryUntilResponseCodes(http.StatusOK, http.StatusAccepted),                  // Keep retry the request until of these response codes has been returned
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

func ExampleNew_WithCustomTransport() {
	t := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Disable HTTP/2 support
	}
	c, err := client.New(&http.Client{Transport: t}, 8,
		transport.WithRetryUntilResponseCodes(http.StatusAccepted, http.StatusOK),
		transport.WithRateLimitCheck(func(h http.Header) (time.Duration, bool) {
			remaining, err := strconv.Atoi(h.Get(`X-Limit-Remaining`))
			if err != nil {
				// No limit was sent with this request, no need to delay
				return 0, false
			}
			until, err := time.Parse(time.RFC1123, h.Get(`X-Limit-Reset`))
			if err != nil {
				return 0, false
			}
			return time.Until(until), remaining < 0
		}),
	)

	if err != nil {
		panic(err)
	}

	if _, err := c.Get(`https://site-with-rate-limits.com`); err != nil {
		panic(err)
	}
}
