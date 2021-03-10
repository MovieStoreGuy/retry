package transport_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MovieStoreGuy/retry"
	"github.com/MovieStoreGuy/retry/http/client"
	"github.com/MovieStoreGuy/retry/http/status"
	"github.com/MovieStoreGuy/retry/http/transport"
)

func TestCreatingTransport(t *testing.T) {
	t.Parallel()
	invalid := [][]transport.Option{
		{transport.WithRetryOptions(retry.WithLogger(nil))},
		{transport.WithRetryOptions(nil)},
	}

	for _, opts := range invalid {
		tr, err := transport.Default(1, opts...)
		assert.Error(t, err)
		assert.Panics(t, func() {
			transport.Must(tr, err)
		})
	}
	assert.Panics(t, func() {
		transport.Must(transport.Default(0))
	})
	assert.Panics(t, func() {
		transport.Must(transport.New(nil, 0))
	})
	assert.NotPanics(t, func() {
		transport.Must(transport.Default(1))
	})
}

func TestRetryClient(t *testing.T) {
	t.Parallel()

	var called int64 = 0

	r := chi.NewRouter()
	r.HandleFunc(`/`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&called, 1)
		code, err := strconv.Atoi(r.URL.Query().Get(`status`))
		if err != nil {
			code = 200
		}
		http.Error(w, http.StatusText(code), code)
	}))
	s := httptest.NewServer(r)
	defer s.Close()

	tests := []struct {
		status   int
		attempts int
		expect   int
		endpoint string
		msg      string
		opts     []transport.Option
	}{
		{status: 400, attempts: 6, expect: 6, endpoint: "/", msg: `Wrapped transport with retry on code`, opts: []transport.Option{
			transport.WithRetryOnStatusCode(400, 500),
		}},
		{status: 418, attempts: 10, expect: 10, endpoint: "/", msg: `Wrapped transport until status code appears`, opts: []transport.Option{
			transport.WithRetryOnStatusGroup(status.Group4xx),
		}},
		{status: 400, attempts: 6, expect: 6, endpoint: "/", msg: `Wrapped transport with retry on code and retry options`, opts: []transport.Option{
			transport.WithRetryOnStatusGroup(status.Group4xx, status.Group5xx),
			transport.WithRetryOptions(
				retry.WithExponentialBackoff(400*time.Millisecond, 1.2),
				retry.WithJitter(100*time.Millisecond),
			),
		}},
	}

	for _, test := range tests {
		atomic.StoreInt64(&called, 0)
		u, err := url.Parse(s.URL)
		require.NoError(t, err, `Failed to parse test server URL`)

		u.Path = path.Join(u.Path, test.endpoint)
		u.RawQuery = fmt.Sprintf(`status=%d`, test.status)

		c, err := client.Default(test.attempts, test.opts...)
		require.NoError(t, err, `Issue with creating default client`)

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		require.NoError(t, err, `Issue passing the request`)

		resp, err := c.Do(req)
		assert.Error(t, err, `Should of failed performing request`, test.msg)
		assert.Nil(t, resp, test.msg)
		assert.Equal(t, test.expect, int(atomic.LoadInt64(&called)))
	}

	c, err := client.Default(6)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	assert.NoError(t, err, `Should not return an error with the default wrapper`)
	assert.NotNil(t, resp)
}
