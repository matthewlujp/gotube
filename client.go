package gotube

import (
	"net/http"
)

type client interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}

type youtubeClient struct{}

// Get wraps http.Get method
func (c *youtubeClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

// Head wraps http.Head method
func (c *youtubeClient) Head(url string) (*http.Response, error) {
	return http.Head(url)
}
