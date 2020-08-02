package retry_test

import (
	"errors"
	"testing"

	"github.com/MovieStoreGuy/retry"
)

func BenchmarkDefaultRetry(b *testing.B) {
	_ = retry.Must().Attempt(b.N, func() error {
		b.ReportAllocs()
		return errors.New(`discard`)
	})
}
