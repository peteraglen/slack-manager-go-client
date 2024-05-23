package client

import (
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Option func(*Options)

type Options struct {
	retryCount        int
	retryWaitTime     time.Duration
	retryMaxWaitTime  time.Duration
	requestLogger     RequestLogger
	retryPolicy       func(*resty.Response, error) bool
	requestHeaders    map[string]string
	basicAuthUsername string
	basicAuthPassword string
	authScheme        string
	authToken         string
}

func newClientOptions() *Options {
	return &Options{
		retryCount:       3,
		retryWaitTime:    500 * time.Millisecond,
		retryMaxWaitTime: 3 * time.Second,
		requestLogger:    &NoopLogger{},
		retryPolicy:      DefaultRetryPolicy,
		requestHeaders: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	}
}

func WithRetryCount(count int) Option {
	return func(o *Options) {
		if count >= 0 {
			o.retryCount = count
		}
	}
}

func WithRetryWaitTime(waitTime time.Duration) Option {
	return func(o *Options) {
		if waitTime >= 100*time.Millisecond {
			o.retryWaitTime = waitTime
		}
	}
}

func WithRetryMaxWaitTime(maxWaitTime time.Duration) Option {
	return func(o *Options) {
		if maxWaitTime >= 100*time.Millisecond {
			o.retryMaxWaitTime = maxWaitTime
		}
	}
}

func WithRequestLogger(logger RequestLogger) Option {
	return func(o *Options) {
		if logger != nil {
			o.requestLogger = logger
		}
	}
}

func WithRetryPolicy(policy func(*resty.Response, error) bool) Option {
	return func(o *Options) {
		if policy != nil {
			o.retryPolicy = policy
		}
	}
}

func WithRequestHeader(header, value string) Option {
	return func(o *Options) {
		header = strings.TrimSpace(header)

		if header == "" || strings.EqualFold(header, "Content-Type") || strings.EqualFold(header, "Accept") {
			return
		}

		o.requestHeaders[header] = value
	}
}

func WithBasicAuth(username, password string) Option {
	return func(o *Options) {
		o.basicAuthUsername = username
		o.basicAuthPassword = password
	}
}

func WithAuthScheme(scheme string) Option {
	return func(o *Options) {
		o.authScheme = scheme
	}
}

func WithAuthToken(token string) Option {
	return func(o *Options) {
		o.authToken = token
	}
}
