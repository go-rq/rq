package rq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func newResponse(raw *http.Response) *Response {
	return &Response{
		Response: raw,
	}
}

type Response struct {
	*http.Response

	PreRequestAssertions  []Assertion
	PostRequestAssertions []Assertion
}

func (resp *Response) Raw() *http.Response {
	return resp.Response
}

func (resp *Response) String() string {
	buf := bytes.NewBuffer(nil)
	resp.Response.Write(buf)
	payload := buf.String()
	resp.Body = io.NopCloser(bytes.NewBufferString(payload))
	return payload
}

func (resp *Response) PrettyString() (string, error) {
	// print out the http response in a pretty format, ex., pretty print JSON responses
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s %s\n", resp.Proto, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(builder, "%s: %s\n", key, value)
		}
	}
	if resp.ContentLength == 0 {
		return builder.String(), nil
	}
	builder.WriteString("\n")
	defer resp.Raw().Body.Close()
	body, err := io.ReadAll(resp.Raw().Body)
	if err != nil {
		return "", err
	}
	resp.Raw().Body = io.NopCloser(bytes.NewBuffer(body))
	switch mimetype := resp.Header.Get("Content-Type"); {
	case strings.Contains(mimetype, "application/json"):
		buf := bytes.NewBuffer(nil)
		if err := json.Indent(buf, body, "", "\t"); err != nil {
			return "", err
		}
		builder.WriteString(buf.String())
	default:
		builder.WriteString(string(body))
	}
	return builder.String(), nil
}
