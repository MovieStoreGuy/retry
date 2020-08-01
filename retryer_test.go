package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/MovieStoreGuy/retry"
)

func TestFailedAttempts(t *testing.T) {
	t.Parallel()
	var called int = 0

	tests := []struct {
		f      func() error
		expect int
		limit  int
		msg    string
	}{
		{f: func() error { called++; return errors.New(`discard`) }, expect: 4, limit: 4, msg: `testing constant failure`},
		{f: func() error { called++; return retry.NoRecover("doom") }, expect: 1, limit: 2, msg: `testing unrecoverable error`},
		{f: nil, expect: 0, limit: 1, msg: `testing nil function`},
		{f: func() error { return nil }, expect: 0, limit: 0, msg: `testing with no attempts`},
	}

	for _, test := range tests {
		called = 0
		err := retry.Must().Attempt(test.limit, test.f)
		assert.Equal(t, test.expect, called, test.msg)
		assert.LessOrEqual(t, called, test.limit, test.msg)
		assert.Error(t, err, test.msg)
	}

}

func TestSuccessfulAttempts(t *testing.T) {
	t.Parallel()

	var called int = 0
	good := []struct {
		f      func() error
		expect int
		limit  int
		msg    string
	}{
		{f: func() error { called++; return nil }, expect: 1, limit: 2, msg: `successful function`},
		{f: func() error {
			called++
			if called == 2 {
				return nil // recovered
			}
			return errors.New(`discard`)
		}, expect: 2, limit: 4, msg: `testing recovered function`},
	}

	for _, test := range good {
		called = 0
		err := retry.Must().Attempt(test.limit, test.f)
		assert.Equal(t, test.expect, called, test.msg)
		assert.LessOrEqual(t, called, test.limit, test.msg)
		assert.NoError(t, err, test.msg)
	}
}

func TestAttemptsUsingContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		ctx context.Context
		msg string
	}{
		{ctx: nil, msg: `testing using nil context`},
		{ctx: ctx, msg: `testing using cancelled context`},
	}

	for _, test := range tests {
		called := 0
		err := retry.Must().AttemptWithContext(test.ctx, 1, func() error {
			called++
			return nil
		})
		assert.Error(t, err, test.msg)
		assert.False(t, called == 1, test.msg)
	}

	ctx, cancel = context.WithCancel(context.Background())
	err := retry.Must().AttemptWithContext(ctx, 2, func() error {
		cancel()
		return errors.New(`discard`)
	})
	assert.Equal(t, ctx.Err(), err, `testing cancelled context during attempts`)
}

func TestInvalidOptions(t *testing.T) {
	t.Parallel()

	invalid := []retry.Option{
		retry.WithLogger(nil),
		retry.WithFixedDelay(-time.Second),
		retry.WithFixedDelay(0),
		retry.WithJitter(-time.Second),
		retry.WithJitter(0),
		retry.WithExponentialBackoff(-time.Second, 1.0),
		retry.WithExponentialBackoff(time.Second, 0.0),
		retry.WithExponentialBackoff(0, 1.0),
	}

	for _, opt := range invalid {
		_, err := retry.New(opt)
		assert.Error(t, err)
		assert.Panics(t, func() {
			_ = retry.Must(opt)
		})
	}
}

func TestValidOption(t *testing.T) {
	t.Parallel()

	valid := []retry.Option{
		retry.WithLogger(zap.NewNop()),
		retry.WithFixedDelay(time.Second),
		retry.WithJitter(time.Second),
		retry.WithExponentialBackoff(time.Second, 2.8),
	}

	for _, opt := range valid {
		assert.NotPanics(t, func() {
			_ = retry.Must(opt)
		})
	}
}

func TestWithAppliedOptions(t *testing.T) {
	t.Parallel()
	log := zap.NewNop()

	opts := [][]retry.Option{
		{retry.WithLogger(log.Named(`no-other-options`))},
		{retry.WithLogger(log), retry.WithFixedDelay(10 * time.Millisecond)},
		{retry.WithLogger(log), retry.WithJitter(10 * time.Millisecond)},
		{retry.WithLogger(log), retry.WithFixedDelay(time.Millisecond), retry.WithExponentialBackoff(10*time.Millisecond, 2.4)},
	}

	for _, apply := range opts {
		r, err := retry.New(apply...)
		require.NoError(t, err, `All options configured are valid`)
		called := 0
		err = r.Attempt(6, func() error {
			called++
			return errors.New(`discard`)
		})
		assert.Error(t, err, `Function was to error out until attempts reached`)
		assert.Equal(t, retry.ErrAttemptsExceeded, err)
		assert.Equal(t, 6, called, `Must have used all allowed attempts`)
	}
}
