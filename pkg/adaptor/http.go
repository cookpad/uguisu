package adaptor

import "net/http"

// HTTPClient is http.Client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
