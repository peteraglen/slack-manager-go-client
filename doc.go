// Package client provides an HTTP client for the Slack Manager API.
//
// The client wraps [github.com/go-resty/resty/v2] with automatic retries,
// configurable connection pooling, and pluggable logging.
//
// # Basic Usage
//
//	c := client.New("https://api.example.com",
//	    client.WithAuthToken("my-token"),
//	    client.WithRetryCount(5),
//	)
//
//	if err := c.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	if err := c.Send(ctx, alert); err != nil {
//	    log.Fatal(err)
//	}
//
// # Configuration
//
// All configuration is supplied as [Option] functions passed to [New].
// Invalid values are silently ignored and the default is retained;
// all configuration is validated when [Client.Connect] is called.
//
// # Retry Behaviour
//
// [DefaultRetryPolicy] retries on HTTP 429 (rate limit) and 5xx server
// errors, and on transient connection errors. It respects the Retry-After
// response header for rate-limit backoff. Context cancellation, deadline
// exceeded, and DNS resolution errors are never retried. Supply a custom
// function via [WithRetryPolicy] to override this behaviour.
//
// # Authentication
//
// Token-based authentication is configured with [WithAuthToken] (and
// optionally [WithAuthScheme]). HTTP Basic authentication is configured
// with [WithBasicAuth]. The two methods are mutually exclusive.
//
// # Logging
//
// Implement [RequestLogger] and supply it via [WithRequestLogger] to
// integrate with your logging library. The default [NoopLogger] discards
// all log output. Ensure your implementation redacts credentials and tokens
// from request and response bodies before persisting logs.
package client
