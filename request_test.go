package rq

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestRequest_String(t *testing.T) {
	t.Run("A request is converted to a string", func(t *testing.T) {
		request := Request{
			Name:   "Get User",
			Method: "GET",
			URL:    "http://localhost:3838/users/123?fizz=buzz",
			Headers: []Header{
				{"Accept", "application/json"},
				{"Authorization", "Bearer {{token}}"},
			},
		}
		expected := `### Get User
GET http://localhost:3838/users/123?fizz=buzz
Accept: application/json
Authorization: Bearer {{token}}
`
		if diff := cmp.Diff(expected, request.String()); diff != "" {
			t.Errorf("request mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestParseRequests(t *testing.T) {
	t.Run("A single request is parsed", func(t *testing.T) {
		input := `### Get User
< {% foo baz bar %}
GET http://localhost:3838/users/123?fizz=buzz
Accept: application/json
Authorization: Bearer {{token}}
`

		requests, err := ParseRequests(input)
		if err != nil {
			t.Error(err)
		}

		if diff := cmp.Diff([]Request{
			{
				Name:             "Get User",
				Method:           "GET",
				PreRequestScript: "foo baz bar",
				URL:              "http://localhost:3838/users/123?fizz=buzz",
				Headers: []Header{
					{"Accept", "application/json"},
					{"Authorization", "Bearer {{token}}"},
				},
			},
		}, requests); diff != "" {
			t.Errorf("requests mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("A single request with script files is parsed", func(t *testing.T) {
		script := `foo
baz
bar`
		scriptPath := path.Join(t.TempDir(), "testScript.js")
		if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
			t.Error(err)
			t.FailNow()
		}
		input := fmt.Sprintf(`### Create a User
< %s
POST http://localhost:3838/users/123?fizz=buzz
Accept: application/json
Authorization: Bearer {{token}}

{ "name": "John Doe" }
< %s
`, scriptPath, scriptPath)

		requests, err := ParseRequests(input)
		if err != nil {
			t.Error(err)
		}

		if diff := cmp.Diff([]Request{
			{
				Name:              "Create a User",
				Method:            "POST",
				PreRequestScript:  "foo\nbaz\nbar",
				PostRequestScript: "foo\nbaz\nbar",
				URL:               "http://localhost:3838/users/123?fizz=buzz",
				Headers: []Header{
					{"Accept", "application/json"},
					{"Authorization", "Bearer {{token}}"},
				},
				Body: `{ "name": "John Doe" }` + "\n",
			},
		}, requests); diff != "" {
			t.Errorf("requests mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("A single request with multi-line scripts is parsed", func(t *testing.T) {
		input := `### Create User
< {% foo 
    baz 
    bar %}
POST http://localhost:3838/users/123?fizz=buzz
Content-Type: application/json
Accept: application/json
Authorization: Bearer {{token}}

{ "name": "John Doe" }

< {% foo
baz
bar %}
`

		requests, err := ParseRequests(input)
		if err != nil {
			t.Error(err)
		}

		if diff := cmp.Diff([]Request{
			{
				Name:              "Create User",
				Method:            "POST",
				PreRequestScript:  "foo\nbaz\nbar",
				PostRequestScript: "foo\nbaz\nbar",
				URL:               "http://localhost:3838/users/123?fizz=buzz",
				Headers: []Header{
					{"Content-Type", "application/json"},
					{"Accept", "application/json"},
					{"Authorization", "Bearer {{token}}"},
				},
				Body: `{ "name": "John Doe" }` + "\n",
			},
		}, requests); diff != "" {
			t.Errorf("requests mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Multiple requests are parsed", func(t *testing.T) {
		input := `### Get User
GET http://localhost:3838/users/123?fizz=buzz
Accept: application/json
Authorization: Bearer {{token}}

### Create a User
POST http://localhost:3838/users
Content-Type: application/json

{
  "name": "John Doe"
}
`

		requests, err := ParseRequests(input)
		if err != nil {
			t.Error(err)
		}

		if diff := cmp.Diff([]Request{
			{
				Name:   "Get User",
				Method: "GET",
				URL:    "http://localhost:3838/users/123?fizz=buzz",
				Headers: []Header{
					{"Accept", "application/json"},
					{"Authorization", "Bearer {{token}}"},
				},
			},
			{
				Name:   "Create a User",
				Method: "POST",
				URL:    "http://localhost:3838/users",
				Headers: []Header{
					{"Content-Type", "application/json"},
				},
				Body: `{
  "name": "John Doe"
}
`,
			},
		}, requests); diff != "" {
			t.Errorf("requests mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestRequest_applyEnv(t *testing.T) {
	t.Run("Variables are replaced", func(t *testing.T) {
		request := Request{
			Method: "GET",
			URL:    "http://localhost:3838/users/{{id}}",
			Headers: []Header{
				{Key: "Authorization", Value: "Bearer {{token}}"},
			},
		}
		expected := `GET http://localhost:3838/users/1234
Authorization: Bearer abc123` + "\n"
		if diff := cmp.Diff(expected, request.applyEnv(WithEnvironment(context.Background(), map[string]string{
			"id":    "1234",
			"token": "abc123",
		})).String()); diff != "" {
			t.Errorf("request mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestRequest_Do(t *testing.T) {
	t.Run("A request is executed", func(t *testing.T) {
		request := Request{
			Method: "GET",
			URL:    "{{host}}/users/1234",
			Headers: []Header{
				{Key: "Authorization", Value: "Bearer abc123"},
			},
		}
		location, _ := time.LoadLocation("GMT")
		expected := strings.Join([]string{
			"HTTP/1.1 200 OK",
			"Content-Length: 29",
			"Content-Type: application/json",
			fmt.Sprintf("Date: %s", time.Now().In(location).Format(time.RFC1123)),
			"",
			`{"id":1234,"name":"John Doe"}`,
		}, "\r\n")
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":1234,"name":"John Doe"}`))
		}))
		defer srv.Close()
		ctx := WithEnvironment(context.Background(), map[string]string{
			"host": srv.URL,
		})
		resp, err := request.Do(ctx)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if diff := cmp.Diff(expected, resp.String()); diff != "" {
			t.Errorf("response mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("The environment is updated by pre-request scripts and successfully executed", func(t *testing.T) {
		request := Request{
			Method: "GET",
			URL:    "{{host}}/users/1234",
			Headers: []Header{
				{Key: "Authorization", Value: "Bearer abc123"},
			},
			PostRequestScript: `assert(response.statusCode === 200, 'The response statusCode is 200')`,
		}
		location, _ := time.LoadLocation("GMT")
		expected := strings.Join([]string{
			"HTTP/1.1 200 OK",
			"Content-Length: 29",
			"Content-Type: application/json",
			fmt.Sprintf("Date: %s", time.Now().In(location).Format(time.RFC1123)),
			"",
			`{"id":1234,"name":"John Doe"}`,
		}, "\r\n")
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":1234,"name":"John Doe"}`))
		}))
		defer srv.Close()

		request.PreRequestScript = fmt.Sprintf(`setEnv('host', "%s")
assert(request !== undefined, 'request is defined')`, srv.URL)
		ctx := WithEnvironment(context.Background(), map[string]string{})
		rt := getRuntime(ctx)
		resp, err := request.Do(WithRuntime(ctx, rt))
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if diff := cmp.Diff(expected, resp.String()); diff != "" {
			t.Errorf("response mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]Assertion{
			{
				Message: "request is defined",
				Success: true,
			},
		}, resp.PreRequestAssertions); diff != "" {
			t.Errorf("assertions mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]Assertion{
			{
				Message: "The response statusCode is 200",
				Success: true,
			},
		}, resp.PostRequestAssertions); diff != "" {
			t.Errorf("assertions mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("The environment is updated by post-request scripts and persisted to the context", func(t *testing.T) {
		request := Request{
			Method: "GET",
			URL:    "{{host}}/users/1234",
			Headers: []Header{
				{Key: "Authorization", Value: "Bearer abc123"},
			},
			PostRequestScript: `assert(response.statusCode === 200, 'The response statusCode is 200');
setEnv('someVar', 'someValue');`,
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":1234,"name":"John Doe"}`))
		}))
		defer srv.Close()

		request.PreRequestScript = fmt.Sprintf(`setEnv('host', "%s")
assert(request !== undefined, 'request is defined')`, srv.URL)
		ctx := WithEnvironment(context.Background(), map[string]string{})
		rt := getRuntime(ctx)
		resp, err := request.Do(WithRuntime(ctx, rt))
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if diff := cmp.Diff([]Assertion{
			{
				Message: "request is defined",
				Success: true,
			},
		}, resp.PreRequestAssertions); diff != "" {
			t.Errorf("assertions mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]Assertion{
			{
				Message: "The response statusCode is 200",
				Success: true,
			},
		}, resp.PostRequestAssertions); diff != "" {
			t.Errorf("assertions mismatch (-want +got):\n%s", diff)
		}
		env := getEnvironment(ctx)
		if diff := cmp.Diff(map[string]string{
			"host":    srv.URL,
			"someVar": "someValue",
		}, env); diff != "" {
			t.Errorf("environment mismatch (-want +got):\n%s", diff)
		}
	})
}
