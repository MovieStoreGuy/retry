package transport_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/MovieStoreGuy/retry"
	"github.com/MovieStoreGuy/retry/transport"
	"github.com/stretchr/testify/assert"
)

func StatusHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, _ := strconv.Atoi(r.URL.Query().Get(`status`))
		switch code {
		case 0, http.StatusOK:
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, http.StatusText(code), code)
	})
}

func TestCreatingTransport(t *testing.T) {
	t.Parallel()
	invalid := [][]transport.Option{
		{transport.WithRetryOptions(retry.WithLogger(nil))},
		{transport.WithRetryOptions(nil)},
		{transport.WithRateLimitCheck(nil)},
	}

	for _, opts := range invalid {
		tr, err := transport.DefaultTransport(1, opts...)
		assert.Error(t, err)
		assert.Panics(t, func() {
			transport.Must(tr, err)
		})
	}
	assert.Panics(t, func() {
		transport.Must(transport.DefaultTransport(0))
	})
	assert.Panics(t, func() {
		transport.Must(transport.New(nil, 0))
	})
	assert.NotPanics(t, func() {
		transport.Must(transport.DefaultTransport(1))
	})
}
