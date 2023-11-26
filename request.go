package rq

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	RequestSeparator = "###"
)

var ErrInvalidRequest = errors.New("invalid request")

var (
	headerRegexp        = regexp.MustCompile(`^([^:]+):\s*(.*)`)
	scriptStartRegexp   = regexp.MustCompile(`^<\s*\{%(.*)`)
	scriptFileRegexp    = regexp.MustCompile(`^<\s*(.*\.js)`)
	scriptEndRegexp     = regexp.MustCompile(`(.*)%\}`)
	scriptOneLineRegexp = regexp.MustCompile(`^<\s*\{%(.*)%\}`)
)

// Request is a struct that holds the HTTP request data.
type Request struct {
	// The name of the request
	// Example: Get User
	Name string

	// PreRequestScript is a piece of JavaScript code that executes before the request is sent.
	// Example: var token = req.environment.get("token");
	PreRequestScript string

	// PostRequestScript is a piece of JavaScript code that executes after the request is sent.
	PostRequestScript string

	// The HTTP method used (GET, POST, PUT, DELETE, etc.)
	Method string

	// The URL of the request
	URL string
	// The HTTP body
	// Note: The body is not parsed.
	// Example: {"foo":"bar"}
	// Example: <xml><foo>bar</foo></xml>
	// Example: foo=bar&baz=qux
	Body string

	// The http Headers
	Headers Headers
}

type Headers []Header

type Header struct {
	Key   string
	Value string
}

func (r Request) DisplayName() string {
	if r.Name != "" {
		return r.Name
	}
	return fmt.Sprintf("%s %s", r.Method, r.URL)
}

func (r Request) String() string {
	var buffer bytes.Buffer
	if r.Name != "" {
		buffer.WriteString(fmt.Sprintf("%s %s\n", RequestSeparator, r.Name))
	}
	buffer.WriteString(r.HttpText())
	return buffer.String()
}

func (r Request) HttpText() string {
	var buffer strings.Builder
	fmt.Fprintf(&buffer, "%s %s\n", r.Method, r.URL)
	for _, header := range r.Headers {
		fmt.Fprintf(&buffer, "%s: %s\n", header.Key, header.Value)
	}
	buffer.WriteString(r.Body)
	return buffer.String()
}

func resolvePath(dir, file string) string {
	if path.IsAbs(file) {
		return file
	}
	return path.Join(dir, file)
}

func (r Request) applyEnv(ctx context.Context) Request {
	env := getEnvironment(ctx)
	r.Method = replaceVariables(r.Method, env)
	r.URL = replaceVariables(r.URL, env)
	r.Body = replaceVariables(r.Body, env)
	r.Headers = replaceVariablesHeaders(r.Headers, env)
	return r
}

func (r Request) Do(ctx context.Context) (*Response, error) {
	rt := getRuntime(ctx)
	rt.setRequest(r)
	ctx = WithRuntime(ctx, rt)
	var preRequestAssertions []Assertion
	if r.PreRequestScript != "" {
		if _, err := rt.vm.RunString(r.PreRequestScript); err != nil {
			return nil, err
		}
		preRequestAssertions = rt.extractAssertions()
		rt.resetAssertions()
	}
	ctx = WithEnvironment(ctx, rt.environment)
	req, err := r.applyEnv(ctx).toHttpRequest(ctx)
	if err != nil {
		return nil, err
	}
	rawResp, err := getRequestRunner(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	resp := newResponse(rawResp)
	if r.PostRequestScript != "" {
		rt.setResponse(resp)
		if _, err := rt.vm.RunString(r.PostRequestScript); err != nil {
			return nil, err
		}
		resp.PostRequestAssertions = rt.extractAssertions()
		resp.PreRequestAssertions = preRequestAssertions
	}
	rt.reset()
	return resp, nil
}

func (r Request) toHttpRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method, r.URL, bytes.NewBufferString(r.Body))
	if err != nil {
		return nil, err
	}
	for _, header := range r.Headers {
		req.Header.Set(header.Key, header.Value)
	}
	return req, nil
}

func replaceVariablesHeaders(headers Headers, variables map[string]string) Headers {
	var result Headers
	for _, header := range headers {
		result = append(result, Header{
			Key:   replaceVariables(header.Key, variables),
			Value: replaceVariables(header.Value, variables),
		})
	}
	return result
}

func ParseRequests(input string) ([]Request, error) {
	return parseRequests("", input)
}

func ParseFromFile(path string) ([]Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	return parseRequests(dir, string(data))
}

func parseRequests(dir, input string) ([]Request, error) {
	scanner := bufio.NewScanner(bytes.NewBufferString(input))
	var requests []Request
	var currentRequest *Request
	headerParsed := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, RequestSeparator) {
			headerParsed = false
			if currentRequest != nil {
				requests = append(requests, *currentRequest)
			}
			currentRequest = &Request{
				Name: strings.TrimSpace(line[3:]),
			}
		} else {
			if currentRequest == nil {
				currentRequest = &Request{}
			}
			if currentRequest.Method == "" {
				if strings.HasPrefix(strings.TrimSpace(line), "<") {
					currentRequest.PreRequestScript, _ = parseRequestScript(dir, scanner)
					continue
				}
				if err := parseMethodAndURL(currentRequest, line); err != nil {
					return nil, err
				}
			} else {
				if !headerParsed {
					currentRequest.Headers = parseHeaders(scanner)
					headerParsed = true
					continue
				}
				if line == "" {
					continue
				}
				if script, ok := parseRequestScript(dir, scanner); ok {
					currentRequest.PostRequestScript = script
					continue
				}
				if currentRequest.PostRequestScript == "" {
					currentRequest.Body += line + "\n"
				}
			}
		}
	}

	if currentRequest != nil {
		requests = append(requests, *currentRequest)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func parseRequestScript(dir string, scanner *bufio.Scanner) (string, bool) {
	var script strings.Builder
	line := strings.TrimSpace(scanner.Text())
	if match := scriptFileRegexp.FindStringSubmatch(line); match != nil {
		// the script is in a file at the path defined after the '<', read the script from the file
		// read the file
		data := strings.TrimSpace(match[1])
		file, err := os.Open(resolvePath(dir, data))
		if err != nil {
			panic(err)
		}
		defer file.Close()

		b, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		return string(b), true
	}
	if match := scriptOneLineRegexp.FindStringSubmatch(line); match != nil {
		return strings.TrimSpace(match[1]), true
	}
	if match := scriptStartRegexp.FindStringSubmatch(line); match != nil {
		script.WriteString(strings.TrimSpace(match[1]) + "\n")
	} else {
		return "", false
	}
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if match := scriptEndRegexp.FindStringSubmatch(line); match != nil {
			script.WriteString(strings.TrimSpace(match[1]))
			return script.String(), true
		}
		script.WriteString(strings.TrimSpace(line) + "\n")
	}
	return script.String(), true
}

func parseHeaders(scanner *bufio.Scanner) Headers {
	var headers Headers
	line := scanner.Text()
	k, v, ok := parseHeader(line)
	if !ok {
		return headers
	}
	headers = append(headers, Header{Key: k, Value: v})
	for scanner.Scan() {
		line = scanner.Text()
		k, v, ok := parseHeader(line)
		if !ok {
			return headers
		}
		headers = append(headers, Header{Key: k, Value: v})
	}
	return headers
}

func parseHeader(line string) (string, string, bool) {
	match := headerRegexp.FindStringSubmatch(line)
	if match == nil {
		return "", "", false
	}
	return match[1], match[2], true
}

func parseMethodAndURL(req *Request, line string) error {
	if req.Method != "" || req.URL != "" {
		return fmt.Errorf("Request method and target has already been parsed %w", ErrInvalidRequest)
	}
	if line == "" {
		return nil
	}
	if match := regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+(.+)$`).FindStringSubmatch(line); match != nil {
		req.Method = match[1]
		req.URL = match[2]
	} else {
		return fmt.Errorf("Request does not include method or URL %w", ErrInvalidRequest)
	}
	return nil
}
