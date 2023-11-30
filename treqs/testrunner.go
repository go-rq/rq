package treqs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-rq/rq"
)

var httpFileFilter = regexp.MustCompile(`^.*\.http$`)

// Run runs all requests provided as argumentss. Each request is executed in a subtest with the
// name of the request and each assertion result is marked as a pass or fail in the test output.
// The requests are executed using the provided context. Environment variables or a shared runtime
// can be provided via the context.
func Run(t *testing.T, ctx context.Context, reqs []rq.Request, options ...Option) {
	settings := Options{}
	for _, option := range options {
		option(&settings)
	}
	for _, request := range reqs {
		t.Run(request.DisplayName(), func(t *testing.T) {
			if settings.Verbose {
				ctx = rq.WithRequestRunner(ctx, httpClient(t))
			}
			resp, err := request.Do(ctx)
			if err != nil {
				t.Error(err)
			}
			if len(request.PreRequestAssertions) > 0 {
				t.Run("Pre-Request Assertions", func(t *testing.T) {
					var failed bool
					for _, assertion := range request.PreRequestAssertions {
						if assertion.Success {
							t.Logf("passed: %s\n", assertion.Message)
						} else {
							failed = true
							t.Errorf("failed: %s\n", assertion.Message)
						}
					}
					if failed {
						t.FailNow()
					}
				})
			}
			if len(resp.PostRequestAssertions) > 0 {
				t.Run("Post-Request Assertions", func(t *testing.T) {
					var failed bool
					for _, assertion := range resp.PostRequestAssertions {
						if assertion.Success {
							t.Logf("passed: %s\n", assertion.Message)
						} else {
							failed = true
							t.Errorf("failed: %s\n", assertion.Message)
						}
					}
					if failed {
						t.FailNow()
					}
				})
			}
		})
	}
}

// RunFile runs all requests in a file. Each requests is run in a subtest
// with the name of the request and each assertion result is marked as a pass or fail in the test output.
func RunFile(t *testing.T, ctx context.Context, path string, options ...Option) {
	requests, err := rq.ParseFromFile(path)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	Run(t, ctx, requests, options...)
}

// RunDir runs all requests from all files with a .http extension found recursively in the directory designated in the path.
func RunDir(t *testing.T, ctx context.Context, path string, options ...Option) {
	var files []string
	filepath.Walk(path, func(filePath string, _ os.FileInfo, _ error) error {
		if httpFileFilter.MatchString(filePath) {
			files = append(files, filePath)
		}
		return nil
	})

	for _, file := range files {
		t.Run(filepath.Clean(file), func(t *testing.T) {
			RunFile(t, ctx, file, options...)
		})
	}
}

type roundtripper struct {
	proxied http.RoundTripper
	t       *testing.T
}

func httpClient(t *testing.T) *http.Client {
	return &http.Client{
		Transport:     &roundtripper{proxied: http.DefaultTransport, t: t},
		CheckRedirect: http.DefaultClient.CheckRedirect,
		Jar:           http.DefaultClient.Jar,
		Timeout:       http.DefaultClient.Timeout,
	}
}

func (r *roundtripper) RoundTrip(request *http.Request) (*http.Response, error) {
	// log out the request
	var builder strings.Builder
	builder.WriteString("\n----- Request\n")
	builder.WriteString(fmt.Sprintf("%s %s\n", request.Method, request.URL.String()))
	for k, v := range request.Header {
		builder.WriteString(fmt.Sprintf("%s: %s\n", k, v[0]))
	}
	if request.Body != nil {
		builder.WriteString("\n")
		body, _ := io.ReadAll(request.Body)
		request.Body.Close()
		request.Body = io.NopCloser(bytes.NewBuffer(body))
		builder.WriteString(string(body))
	}

	r.t.Log(builder.String())
	builder.Reset()
	// run the request
	start := time.Now() // start a timer
	resp, err := r.proxied.RoundTrip(request)

	// log out the response
	builder.WriteString(fmt.Sprintf("----- Response (duration: %s)\n", time.Since(start)))
	defer resp.Body.Close()
	if resp.ContentLength > 0 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// reset the response body to be read, then set it again to a new
		// reader afterwards so it's available by callers
                resp.Body = io.NopCloser(bytes.NewBuffer(b))
		buf := bytes.NewBuffer(nil)
		resp.Write(buf)
		builder.WriteString(buf.String())
		resp.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	r.t.Log(builder.String())
	return resp, err
}
