package retry

import (
	"context"
	"errors"
)

type retry struct {
	actions []func(remaining, limit int)
}

var _ Retryer = (*retry)(nil)

// New creates a new retry with the configured options provided.
// An error is returned if any of the options failed to apply
func New(opts ...Option) (Retryer, error) {
	r := &retry{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Must is a convenience function for New to avoid having to handle the error
// and allow for inline creation. If an error was to be returned from the wrapped New
// function, it would cause this function to panic instead.
func Must(opts ...Option) Retryer {
	r, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *retry) Do(limit int, f func() error) error {
	return r.do(context.Background(), limit, f)
}

func (r *retry) DoWithContext(ctx context.Context, limit int, f func() error) error {
	return r.do(ctx, limit, f)
}

// deprecated: Should use the Do method instead
func (r *retry) Attempt(limit int, f func() error) error {
	return r.Do(limit, f)
}

// deprecated: Should use the DoWithContext method instead
func (r *retry) AttemptWithContext(ctx context.Context, limit int, f func() error) error {
	return r.DoWithContext(ctx, limit, f)
}

func (r *retry) do(ctx context.Context, limit int, f func() error) error {
	if ctx == nil || ctx.Err() != nil {
		return errors.New(`invalid context provided`)
	}
	if f == nil {
		return errors.New(`invalid function provided`)
	}

	// Since limit is not being check if negative, the default assumes all
	// avaliable attempts have been exceeded
	err := errors.New(`exceeded allowed attempts`)
	// It is permissable to cache the channel returned here in order to avoid the locking call
	// within the Done method.
	done := ctx.Done()
	for rem := limit; rem > 0; rem-- {
		select {
		case <-done:
			// Context has be finalised, need to exit
			return ctx.Err()
		default:
			// Avoid indefinate waiting on context to finish
		}

		if err = f(); err == nil {
			return nil
		}

		// Check if err is marked as an abort error
		// an exit from there
		if HasAborted(err) {
			return err
		}

		for _, a := range r.actions {
			a(rem, limit)
		}
	}
	// Returns the last error recorded
	return ExceededRetries(err.Error())
}
