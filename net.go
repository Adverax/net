package net

import (
	"context"
	"net/http"
)

type HttpError struct {
	Code int
	Text string
}

func (that *HttpError) Error() string {
	return that.Text
}

// Messenger is abstract http manager
type Messenger interface {
	Request(ctx context.Context, request *http.Request) (resp *http.Response, err error)
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
