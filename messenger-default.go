package net

import (
	"context"
	"net/http"
)

type DefaultMessenger struct {
	client HttpClient
}

func (that *DefaultMessenger) Request(ctx context.Context, request *http.Request) (resp *http.Response, err error) {
	return that.client.Do(request)
}

func NewDefaultMessenger(client HttpClient) *DefaultMessenger {
	if client == nil {
		client = http.DefaultClient
	}

	return &DefaultMessenger{
		client: client,
	}
}
