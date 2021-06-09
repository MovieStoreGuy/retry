package retry_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MovieStoreGuy/retry"
)

func TestWrappedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err error
		is  func(error) bool
		msg string
	}{
		{err: retry.AbortedRetries(errors.New(`unable to continue`)), is: retry.HasAborted, msg: `Checks to see if an abort error correctly validates`},
		{err: retry.ExceededRetries(errors.New(`too many attempts`)), is: retry.HasExceeded, msg: `Checks to see if an exceeded error correctly validates`},
		{err: retry.AbortedRetries(errors.New(`doom`)), is: func(err error) bool {
			return !retry.HasExceeded(err)
		}, msg: `Ensures an abort error can not validate as exceeded error`},
		{err: retry.ExceededRetries(errors.New(`too many attempts`)), is: func(err error) bool {
			return !retry.HasAborted(err)
		}, msg: `Ensures an exceeded error can not validate as an abort error`},
		{err: nil, is: func(err error) bool { return !retry.HasAborted(err) }, msg: `Ensure that nil does not resolve as an abort`},
		{err: nil, is: func(err error) bool { return !retry.HasExceeded(err) }, msg: `Ensure that nil does not resolve as an exceeded`},
	}

	for _, test := range tests {
		assert.True(t, test.is(test.err), test.msg)
	}
}

func TestErrorPrinting(t *testing.T) {
	t.Parallel()

	assert.Contains(t, retry.AbortedRetries(errors.New("")).Error(), `aborted retries:`)
	assert.Contains(t, retry.ExceededRetries(errors.New("")).Error(), `exceeded attempts:`)
}
