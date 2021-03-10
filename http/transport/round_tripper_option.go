package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/MovieStoreGuy/retry"
	"github.com/MovieStoreGuy/retry/http/status"
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

func WithRetryOnStatusCode(codes ...int) Option {
	return func(c *config) error {
		table := make(map[int]struct{})
		for _, code := range codes {
			if _, exist := c.table[code]; exist {
				return fmt.Errorf("status code %d already exist", code)
			}
			table[code] = struct{}{}
		}

		c.checks = append(c.checks, func(r *http.Response) error {
			if r == nil {
				return errors.New("nil response provided")
			}
			if _, exist := table[r.StatusCode]; exist {
				return fmt.Errorf("returned status code %d allows retries", r.StatusCode)
			}
			return nil
		})

		return nil
	}
}

func WithRetryOnStatusGroup(groups ...int) Option {
	return func(c *config) error {
		table := make(map[int]struct{})
		for _, g := range groups {
			if _, exist := c.table[g]; exist {
				return fmt.Errorf("status group %dxx already exists", g)
			}
			table[g] = struct{}{}
		}

		c.checks = append(c.checks, func(r *http.Response) error {
			if r == nil {
				return errors.New("nil response provided")
			}
			if _, exist := table[status.Group(r.StatusCode)]; exist {
				return fmt.Errorf("returned status group %dxx allows retries", status.Group(r.StatusCode))
			}
			return nil
		})

		return nil
	}
}
