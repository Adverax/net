package net

import (
	"context"
	"net/http"
)

type MessengerWithHeaders struct {
	messenger Messenger
	headers   map[string]string
}

func (that *MessengerWithHeaders) Request(ctx context.Context, request *http.Request) (resp *http.Response, err error) {
	for header, value := range that.headers {
		request.Header.Set(header, value)
	}

	return that.messenger.Request(ctx, request)
}

// NewMessengerHeaders is constructor for create MessengerWithHeaders
func NewMessengerHeaders(messenger Messenger, headers map[string]string) *MessengerWithHeaders {
	return &MessengerWithHeaders{
		messenger: messenger,
		headers:   headers,
	}
}
