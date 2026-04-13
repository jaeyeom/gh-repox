package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DefaultRetryAttempts is the number of attempts for post-creation retries.
const DefaultRetryAttempts = 5

// DefaultRetryDelay is the delay between retry attempts.
const DefaultRetryDelay = 2 * time.Second

// IsNotFoundError reports whether err looks like an HTTP 404 response from the
// GitHub API (e.g. "HTTP 404: Not Found" or "Not Found").
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "not found") || strings.Contains(lower, "404")
}

// RetryOnNotFound calls fn up to maxAttempts times, retrying when the returned
// error looks like a 404. Non-404 errors and nil are returned immediately.
// This is used after repo creation to handle GitHub's eventual consistency.
func RetryOnNotFound(ctx context.Context, maxAttempts int, delay time.Duration, fn func() error) error {
	var err error
	for attempt := range maxAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("%w (after %d attempt(s), last error: %v)", ctx.Err(), attempt, err)
			case <-time.After(delay):
			}
		}
		err = fn()
		if err == nil || !IsNotFoundError(err) {
			return err
		}
	}
	return err
}
