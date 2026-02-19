package client

import (
	"context"
	"errors"
	"net"

	"github.com/go-resty/resty/v2"
)

// DefaultRetryPolicy is the default retry condition used by [Client]. It
// retries on HTTP 429 (rate limit) and 5xx server errors, and on transient
// connection errors. It does not retry on context cancellation, deadline
// exceeded, or DNS resolution failures.
//
// Supply a custom function via [WithRetryPolicy] to override this behaviour.
func DefaultRetryPolicy(r *resty.Response, err error) bool {
	if err != nil {
		// Don't retry on context cancellation or deadline exceeded
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}

		// Don't retry on DNS resolution errors
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return false
		}

		// Retry on other connection errors
		return true
	}

	// Retry on 429 (rate limit) and 5xx (server errors)
	return r.StatusCode() == 429 || r.StatusCode() >= 500
}
