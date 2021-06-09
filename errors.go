package retry

import (
	"errors"
	"fmt"
)

type exceeded struct {
	err error
}

func (e exceeded) Error() string {
	return fmt.Sprintf("exceeded attempts: %v", e.err)
}

func (e exceeded) Unwrap() error {
	return e.err
}

// HasExceeded checks error to validate
// if the exectued function has notified that it exceeded the limit.
func HasExceeded(err error) bool {
	if err == nil {
		return false
	}
	return errors.As(err, &exceeded{})
}

// ExceededRetries wraps the message passed and returns
// an error that be read by the error handler within the retry client.
func ExceededRetries(err error) error {
	return exceeded{err: err}
}

type abort struct {
	err error
}

func (a abort) Error() string {
	return fmt.Sprintf("aborted retries: %v", a.err)
}

func (a abort) Unwrap() error {
	return a.err
}

// AbortedRetries wraps the message passed and returns
// an error that be read by the error handler within the retry client.
func AbortedRetries(err error) error {
	return abort{err: err}
}

// HasAborted checks the error to validate
// if the executed function has notified that it should
// abort now.
func HasAborted(err error) bool {
	if err == nil {
		return false
	}
	return errors.As(err, &abort{})
}
