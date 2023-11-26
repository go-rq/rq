package rq

import (
	"context"
	"net/http"
)

type (
	requestRunnerContextKey struct{}
	RequestRunner           interface {
		Do(req *http.Request) (*http.Response, error)
	}
)

func WithRequestRunner(ctx context.Context, runner RequestRunner) context.Context {
	return context.WithValue(ctx, requestRunnerContextKey{}, runner)
}

func getRequestRunner(ctx context.Context) RequestRunner {
	if runner, ok := ctx.Value(requestRunnerContextKey{}).(RequestRunner); ok {
		return runner
	}
	return http.DefaultClient
}
