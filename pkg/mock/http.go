package mock

import (
	"io"
	"net/http"
	"strings"
)

// Response holds the data for a single canned HTTP response.
type Response struct {
	Code    int
	Body    io.ReadCloser
	Headers http.Header
}

// HTTPClient is a mock HTTP client for testing.
// If Responses is non-empty, each call to Do consumes the next entry (the last
// entry is repeated once exhausted).  Otherwise RespCode/RespBody are used for
// every call, preserving backwards compatibility.
type HTTPClient struct {
	Requests  []*http.Request
	RespCode  int
	RespBody  io.ReadCloser
	Responses []Response
}

func (x *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	x.Requests = append(x.Requests, req)

	if len(x.Responses) > 0 {
		idx := len(x.Requests) - 1
		if idx >= len(x.Responses) {
			idx = len(x.Responses) - 1
		}
		r := x.Responses[idx]
		body := r.Body
		if body == nil {
			body = io.NopCloser(strings.NewReader("OK"))
		}
		headers := r.Headers
		if headers == nil {
			headers = http.Header{}
		}
		return &http.Response{
			StatusCode: r.Code,
			Body:       body,
			Header:     headers,
		}, nil
	}

	code := x.RespCode
	if code == 0 {
		code = 200
	}
	body := x.RespBody
	if body == nil {
		body = io.NopCloser(strings.NewReader("OK"))
	}

	return &http.Response{
		StatusCode: code,
		Body:       body,
	}, nil
}

func (x *HTTPClient) Body(n int) string {
	if n >= len(x.Requests) {
		panic("n is too large")
	}

	body, err := x.Requests[n].GetBody()
	if err != nil {
		panic(err)
	}
	defer body.Close() //nolint:errcheck
	raw, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (x *HTTPClient) RequestNum() int {
	return len(x.Requests)
}
