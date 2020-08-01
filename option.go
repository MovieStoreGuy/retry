package retry

import (
	"errors"
	"math/rand"
	"sync/atomic"
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
		r.post = append(r.post, func() {
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
		r.post = append(r.post, func() {
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
	// Technically not required but to help play safe since it is being shared
	// among to different functions, access to it will require atomic
	d := int64(delay)
	return func(r *retry) error {
		if delay <= 0 {
			return errors.New(`delay must be positive value`)
		}
		if multiplier < 1.0 {
			return errors.New(`multiplier must be greater than 1.0`)
		}
		r.pre = append(r.pre, func() {
			r.log.Info(`Setting the delay to initial value`, zap.Duration(`delay`, delay), zap.String(`step`, `expo-backoff`))
			atomic.StoreInt64(&d, int64(delay))
		})
		r.post = append(r.post, func() {
			t := atomic.LoadInt64(&d)
			r.log.Info(`Delaying execution`, zap.Duration(`delay`, time.Duration(t)))
			time.Sleep(time.Duration(t))
			atomic.StoreInt64(&d, int64(float64(t)*multiplier))
		})
		return nil
	}
}
