package rq

import "context"

type loggerKey struct{}

// Logger is the interface that wraps Log method. Log entries made in the
// pre-request and post-request scripts are passed to the Log method.
type Logger interface {
	Log(...any)
}

// WithLogger returns a new context with the given logger. When a request is executed, the logger
// is used to log out the pre-request and post-request script logs
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func getLogger(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey{}).(Logger); ok {
		return logger
	}
	return nil
}
