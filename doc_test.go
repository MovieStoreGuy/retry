package retry_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/MovieStoreGuy/retry"
)

func ExampleNew() {
	r, err := retry.New()
	if err != nil {
		panic(err)
	}
	err = r.Attempt(4, func() error {
		fmt.Println(`Woohoo`)
		return nil
	})
	if err != nil {
		panic(err)
	}
	// Output: Woohoo
}

func ExampleMust() {
	r := retry.Must(
		retry.WithFixedDelay(100*time.Millisecond), // Ensures each failed attempt waits 100ms
		retry.WithJitter(10*time.Millisecond),      // Ensures each failed attempt waits at most 10ms
	)

	_ = r.Attempt(3, func() error {
		fmt.Print(`tick...`)
		return errors.New(`boom`)
	})

	// Output: tick...tick...tick...
}
