package mock

import (
	"io"
	"net/http"
	"strings"
)

type HTTPClient struct {
	Requests []*http.Request
	RespCode int
	RespBody io.ReadCloser
}

func (x *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	x.Requests = append(x.Requests, req)

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
	defer body.Close()
	raw, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (x *HTTPClient) RequestNum() int {
	return len(x.Requests)
}
