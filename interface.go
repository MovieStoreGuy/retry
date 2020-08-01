package retry

import (
	"context"
	"errors"
)

var (
	// ErrAttemptsExceeded is return when number of attemps exceeds the allowed limit given
	ErrAttemptsExceeded = errors.New(`exceeded allowed attempts`)
)

// Retryer abstracts the retry functionality of executing a function
type Retryer interface {

	// Attempt will execute the function until it has reached the permissable limit
	// that is passed. If the function was to return nil, the retry will exit early
	// otherwise, the error is passed to an error handler and the next attempt is
	// started after the post execution function have run.
	// If the attempt limit has been reached, ErrAttemptsExceeded is returned.
	Attempt(limit int, f func() error) error

	// AttemptWithContext extendes the Attempt method by ensuring that any attempts are
	// aborted if the passed context is done.
	AttemptWithContext(ctx context.Context, limit int, f func() error) error
}
