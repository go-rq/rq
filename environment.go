package req

import (
	"context"
	"regexp"
)

var variableRegexp = regexp.MustCompile(`{{(.*?)}}`)

type environmentContextKey struct{}

func replaceVariables(input string, variables map[string]string) string {
	result := variableRegexp.ReplaceAllStringFunc(input, func(match string) string {
		varName := match[2 : len(match)-2]
		if val, ok := variables[varName]; ok {
			return val
		}
		return match
	})
	return result
}

func WithEnvironment(ctx context.Context, env map[string]string) context.Context {
	return context.WithValue(ctx, environmentContextKey{}, env)
}

func ResetEnvironment(ctx context.Context) {
	WithEnvironment(ctx, map[string]string{})
}

func getEnvironment(ctx context.Context) map[string]string {
	if env, ok := ctx.Value(environmentContextKey{}).(map[string]string); ok {
		return env
	}
	return map[string]string{}
}
