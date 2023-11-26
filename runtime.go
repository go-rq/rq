package rq

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	"github.com/dop251/goja"
)

type Runtime struct {
	vm          *goja.Runtime
	environment map[string]string
}

type Assertion struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (r *Runtime) extractEnvironment() {
	vm, err := r.vm.RunString(`environment`)
	if err != nil {
		panic(err)
	}
	r.environment = vm.Export().(map[string]string)
}

func (r *Runtime) extractAssertions() []Assertion {
	vm, err := r.vm.RunString(`assertions`)
	if err != nil {
		panic(err)
	}
	return vm.Export().([]Assertion)
}

func (r *Runtime) reset() {
	r.vm.Set("environment", r.environment)
	r.vm.Set("request", nil)
	r.vm.Set("response", nil)
	r.resetAssertions()
}

func (r *Runtime) resetAssertions() {
	r.vm.Set("assertions", []Assertion{})
}

func (r *Runtime) setRequest(req Request) {
	r.vm.Set("request", map[string]any{
		"name":    req.Name,
		"body":    req.Body,
		"headers": req.Headers,
		"method":  req.Method,
		"url":     req.URL,
	})
}

func (r *Runtime) setResponse(resp *Response) {
	// tee the body into a string buffer so we can read it multiple times
	// without draining the body
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body = ioutil.NopCloser(bytes.NewBufferString(string(b)))
	r.vm.Set("response", map[string]any{
		"body":       string(b),
		"headers":    resp.Header,
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
	})
}

// scripts are javascript scripts that are loaded into each runtime instance.
var scripts = []string{
	`function assert(condition, message) {
  assertions.push({ Message: message, Success: condition })
}`,

	`function setEnv(key, value) {
  environment[key] = value
}`,

	`function getEnv(key) {
return environment[key]
  }`,
}

type runtimeContextKey struct{}

func WithRuntime(ctx context.Context, rt *Runtime) context.Context {
	return WithEnvironment(context.WithValue(ctx, runtimeContextKey{}, rt), rt.environment)
}

func getRuntime(ctx context.Context) *Runtime {
	if rt, ok := ctx.Value(runtimeContextKey{}).(*Runtime); ok {
		return rt
	}

	rt := &Runtime{
		vm:          goja.New(),
		environment: getEnvironment(ctx),
	}
	for _, script := range scripts {
		_, err := rt.vm.RunString(script)
		if err != nil {
			panic(err)
		}
	}
	rt.reset()
	return rt
}
