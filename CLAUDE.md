# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go HTTP client library for the Slack Manager API. Wraps [resty](https://github.com/go-resty/resty) with domain-specific functionality for sending alerts. Single package (`client`) with functional options pattern for configuration.

## Build Commands

```bash
make init              # Initialize modules (go mod tidy)
make test              # Full test suite: gosec, fmt, test with race detection, vet
make lint              # Run golangci-lint
make lint-fix          # Auto-fix linting issues
make bump-common-lib   # Update slack-manager-common to latest
```

Run a single test:
```bash
go test -run TestName ./...
```

**IMPORTANT:** Both `make test` and `make lint` MUST pass with zero errors before committing any changes. This applies regardless of whether the errors were introduced by your changes or existed previously - all issues must be resolved before committing. Always run both commands to verify code quality.

## Architecture

**Core Components:**
- `Client` - Main client wrapping resty.Client with Connect/Send methods
- `Options` - Functional options pattern for configuration (retries, auth, logging)
- `DefaultRetryPolicy` - Retry logic that handles 429/5xx, skips DNS and context errors
- `RequestLogger` - Interface for pluggable logging (NoopLogger default)

**Workflow:**
```go
c := client.New(baseURL, client.WithRetryCount(5), client.WithAuthToken("token"))
c, err := c.Connect(ctx)  // Validates via ping
err = c.Send(ctx, alerts...)
```

**Dependencies:**
- `github.com/go-resty/resty/v2` - HTTP client with retry support
- `github.com/peteraglen/slack-manager-common` - Shared Alert type

## Code Style

- Uses golangci-lint with strict config (see `.golangci.yaml`)
- All operations require context for cancellation support
- Errors wrapped with `fmt.Errorf("%w")` for chain inspection
- Protected headers (Content-Type, Accept) cannot be overridden
