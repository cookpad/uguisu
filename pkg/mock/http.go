package mock

import (
	"io"
	"io/ioutil"
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
		body = ioutil.NopCloser(strings.NewReader("OK"))
	}

	return &http.Response{
		StatusCode: code,
		Body:       body,
	}, nil
}

func (x *HTTPClient) Body(n int) string {
	if len(x.Requests) < n {
		panic("n is too large")
	}

	body, err := x.Requests[n].GetBody()
	if err != nil {
		panic(err)
	}

	raw, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (x *HTTPClient) RequestNum() int {
	return len(x.Requests)
}
