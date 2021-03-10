package retry

import (
	"errors"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// Option allows for additional functionality to be added to the Retryer
// on creation
type Option func(r *retry) error

// WithLogger sets the logger used for the retryer since by default
// it will use a noop logger
func WithLogger(log *zap.Logger) Option {
	return func(r *retry) error {
		if log == nil {
			return errors.New(`logger is nil`)
		}
		r.log = log
		return nil
	}
}

// WithFixedDelay will set the delay experienced after each failed attempted
func WithFixedDelay(delay time.Duration) Option {
	return func(r *retry) error {
		if delay <= 0 {
			return errors.New(`delay must be a positive value`)
		}
		r.actions = append(r.actions, func(_, _ int) {
			r.log.Info(`Delaying executon`, zap.Duration(`delay`, delay), zap.String(`step`, `fixed-delay`))
			time.Sleep(delay)
		})
		return nil
	}
}

// WithJitter generates a random sleep interval between [0, delay) to help spread retry
// attemps over different intervals
func WithJitter(delay time.Duration) Option {
	return func(r *retry) error {
		if delay <= 0 {
			return errors.New(`delay must be a positive value`)
		}
		r.actions = append(r.actions, func(_, _ int) {
			t := time.Duration(rand.Int63n(int64(delay)))
			r.log.Info(`Delaying executon`, zap.Duration(`delay`, t), zap.String(`step`, `jitter`))
			time.Sleep(t)
		})
		return nil
	}
}

// WithExponentialBackoff will start from a fixed delay and increase the delay amount
// by increasing it by the multiplier amount
func WithExponentialBackoff(delay time.Duration, multiplier float64) Option {
	return func(r *retry) error {
		if delay <= 0 {
			return errors.New(`delay must be positive value`)
		}
		if multiplier < 1.0 {
			return errors.New(`multiplier must be greater than 1.0`)
		}

		r.actions = append(r.actions, func(remaining, limit int) {
			t := delay * time.Duration(multiplier*(float64(remaining-limit)))
			r.log.Info(`Delaying execution`, zap.Duration(`delay`, time.Duration(t)))
			time.Sleep(t)
		})

		return nil
	}
}
