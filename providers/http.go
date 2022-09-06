package providers

import (
	"io"
	"net/http"
)

type HttpClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type HttpClientImpl struct {
}

func (h HttpClientImpl) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return http.Post(url, contentType, body)
}
