package client

import (
	"context"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
)

func DefaultRetryPolicy(r *resty.Response, err error) bool {
	// Retry on all connection errors, except for when the context is canceled or deadline exceeded
	// Also skip retries on DNS resolution errors.
	if err != nil {
		return !errors.Is(err, context.Canceled) &&
			!errors.Is(err, context.DeadlineExceeded) &&
			!strings.Contains(err.Error(), "no such host")
	}

	// Retry on 429 and 5xx errors
	return r.StatusCode() == 429 || r.StatusCode() >= 500
}
