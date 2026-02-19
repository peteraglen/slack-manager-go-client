package client

// RequestLogger is the interface used by [Client] for logging HTTP requests
// and errors. Implement this interface to integrate with your logging library
// and supply the implementation via [WithRequestLogger].
type RequestLogger interface {
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
	Debugf(format string, v ...any)
}

// NoopLogger is a [RequestLogger] that silently discards all log messages.
// It is the default logger used when no logger is provided to [New].
type NoopLogger struct{}

func (l *NoopLogger) Errorf(_ string, _ ...any) {}
func (l *NoopLogger) Warnf(_ string, _ ...any)  {}
func (l *NoopLogger) Debugf(_ string, _ ...any) {}
