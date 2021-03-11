package retry_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MovieStoreGuy/retry"
)

func BenchmarkRetryWithoutError_FiveAttemptsNoDelays(b *testing.B) {
	rt, err := retry.New()

	require.NoError(b, err, "Must have a valid retryer")
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err = rt.Do(5, func() error {
			return nil
		})
	}
	assert.NoError(b, err, "Must not error during successful operation")
}

func BenchmarkRetryWithError_FiveAttemptsNoDelays(b *testing.B) {
	rt, err := retry.New()

	require.NoError(b, err, "Must have a valid retryer")
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err = rt.Do(5, func() error {
			return errors.New("boom")
		})
	}
	assert.Error(b, err, "Must not error during successful operation")
}
