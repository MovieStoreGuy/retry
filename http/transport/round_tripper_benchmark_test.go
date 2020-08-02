package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MovieStoreGuy/retry/http/client"
	"github.com/MovieStoreGuy/retry/http/transport"
	"github.com/stretchr/testify/require"
)

type benchTransport struct {
	do func()
	rt http.RoundTripper
}

func (bt *benchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bt.do()
	return bt.rt.RoundTrip(req)
}

func BenchTransport(rt http.RoundTripper, do func()) http.RoundTripper {
	return &benchTransport{
		do: do,
		rt: rt,
	}
}

func BenchmarkFailingTransport(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(404), 404)
	}))
	defer s.Close()

	c, err := client.New(&http.Client{
		Transport: BenchTransport(http.DefaultTransport, func() {
			b.ReportAllocs()
		})}, 4,
		transport.WithRetryUntilResponseCodes(200),
	)
	require.NoError(b, err)
	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if resp, err := c.Do(req); err == nil && resp != nil {
			require.NoError(b, resp.Body.Close())
		}
		s.CloseClientConnections()
	}
}
