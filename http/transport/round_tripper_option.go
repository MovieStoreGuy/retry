package transport

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/MovieStoreGuy/retry"
)

// Option allow for advanced configuration of the transport being created.
type Option func(*config) error

// WithRetryOptions applies the options to the underlying retry process.
func WithRetryOptions(opts ...retry.Option) Option {
	return func(cf *config) error {
		for _, opt := range opts {
			if opt == nil {
				return errors.New(`nil retry option provided`)
			}
			cf.rtOpts = append(cf.rtOpts, opt)
		}
		return nil
	}
}

// WithNoRetryOnResponseCodes will compare the given response code(s) to
// what has been returned and if it matches, the retry function will stop retrying before
// max attempts have been reached.
// To ensure that this check will stop retries, it is prepended to the start of retry checks.
func WithNoRetryOnResponseCodes(codes ...int) Option {
	table := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		table[code] = struct{}{}
	}
	return func(cf *config) error {
		for _, code := range codes {
			if _, exist := cf.table[code]; exist {
				return fmt.Errorf(`clash on response code [%d:%s]`, code, http.StatusText(code))
			}
			cf.table[code] = struct{}{}
		}
		prepend := []func(*http.Response) error{
			func(resp *http.Response) error {
				if resp == nil {
					return errors.New(`response is nil`)
				}
				if _, exist := table[resp.StatusCode]; exist {
					return retry.AbortedRetries(fmt.Sprintf(`unable to recover response status code: %d`, resp.StatusCode))
				}
				return nil
			},
		}
		cf.checks = append(prepend, cf.checks...)
		return nil
	}
}

// WithRetryOnResponseCodes will look at the returned response codes
// cause the retry function to process the request again
func WithRetryOnResponseCodes(codes ...int) Option {
	table := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		table[code] = struct{}{}
	}
	return func(cf *config) error {
		for _, code := range codes {
			if _, exist := cf.table[code]; exist {
				return fmt.Errorf(`clash on response code [%d:%s]`, code, http.StatusText(code))
			}
			cf.table[code] = struct{}{}
		}
		cf.checks = append(cf.checks, func(resp *http.Response) error {
			if resp == nil {
				return errors.New(`response is nil`)
			}
			if _, exist := table[resp.StatusCode]; exist {
				return fmt.Errorf(`returned status code allows retries: %d`, resp.StatusCode)
			}
			return nil
		})
		return nil
	}
}

// WithRetryUntilResponseCodes will force the system to retry until one of the
// provided codes is meet.
func WithRetryUntilResponseCodes(codes ...int) Option {
	table := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		table[code] = struct{}{}
	}
	return func(cf *config) error {
		if len(codes) == 0 {
			return errors.New(`requires at least one response code to compare against`)
		}
		for _, code := range codes {
			if _, exist := cf.table[code]; exist {
				return fmt.Errorf(`clash on response code [%d:%s]`, code, http.StatusText(code))
			}
			cf.table[code] = struct{}{}
		}
		cf.checks = append(cf.checks, func(resp *http.Response) error {
			if _, ok := table[resp.StatusCode]; ok {
				return nil
			}
			return fmt.Errorf(`unmatched status code %d, continue retrying`, resp.StatusCode)
		})
		return nil
	}
}

// WithRateLimitCheck allows for dynamic delay based on returned headers from
// the response. The function must return a positive Duration and true in order to apply
// the delay once exceeded the allowed rate limit
func WithRateLimitCheck(limiter func(http.Header) (time.Duration, bool)) Option {
	delay := int64(0)
	return func(cf *config) error {
		if limiter == nil {
			return errors.New(`limiter function is nil`)
		}
		cf.rtOpts = append(cf.rtOpts, retry.WithDynamicDelay(&delay))
		cf.checks = append(cf.checks, func(resp *http.Response) error {
			if resp == nil {
				return errors.New(`response is nil`)
			}

			// Enforce that dynamic rate limiting must have a status code of 429
			// to comply with http standards
			if resp.StatusCode != http.StatusTooManyRequests {
				return nil
			}

			t, err := time.Duration(0), error(nil)

			// If we have exceeded the rate limit but the duration to wait has passed
			// then we have technically not exceeded the current rate limit but a previous one

			if d, exceeded := limiter(resp.Header); exceeded && d > 0 {
				t, err = d, errors.New(`exceeded rate limit`)
			}
			atomic.StoreInt64(&delay, int64(t))
			return err
		})
		return nil
	}
}
