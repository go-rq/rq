package rq

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/dop251/goja"
)

type Runtime struct {
	vm          *goja.Runtime
	environment map[string]string
	request     *Request
}

type Assertion struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (r *Runtime) extractEnvironment() {
	value, err := r.vm.RunString(`environment`)
	if err != nil {
		panic(err)
	}
	r.environment = value.Export().(map[string]string)
}

func (r *Runtime) extractAssertions() []Assertion {
	value, err := r.vm.RunString(`assertions`)
	if err != nil {
		panic(err)
	}
	return value.Export().([]Assertion)
}

func (r *Runtime) extractLogs() []string {
	value, err := r.vm.RunString(`logs`)
	if err != nil {
		panic(err)
	}
	return value.Export().([]string)
}

func (r *Runtime) reset() {
	r.request = nil
	r.vm.Set("environment", r.environment)
	r.vm.Set("request", nil)
	r.vm.Set("response", nil)
	r.resetLogs()
	r.resetAssertions()
}

func (r *Runtime) resetAssertions() {
	r.vm.Set("assertions", []Assertion{})
}

func (r *Runtime) resetLogs() {
	r.vm.Set("logs", []string{})
}

func (r *Runtime) setRequest(req *Request) {
	r.request = req
	r.vm.Set("request", map[string]any{
		"name":    req.Name,
		"body":    req.Body,
		"headers": req.Headers,
		"method":  req.Method,
		"url":     req.URL,
	})
}

func (r *Runtime) executeScript(script string) error {
	_, err := r.vm.RunString(script)
	req := r.extractRequest()
	if value, ok := req["skip"].(bool); ok {
		r.request.Skip = value
	}
	return err
}

func (r *Runtime) extractRequest() map[string]any {
	value, err := r.vm.RunString("request")
	if err != nil {
		panic(err)
	}
	return value.Export().(map[string]any)
}

func (r *Runtime) setResponse(resp *Response) {
	// tee the body into a string buffer so we can read it multiple times
	// without draining the body
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body = io.NopCloser(bytes.NewBufferString(string(b)))
	respData := map[string]any{
		"body":       string(b),
		"headers":    resp.Header,
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var data any
		if err := json.Unmarshal(b, &data); err != nil {
			panic(err)
		}
		respData["json"] = data
	}
	r.vm.Set("response", respData)
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
	`function log(entry) {
  logs.push(entry)
}
`,
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
		environment: GetEnvironment(ctx),
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
