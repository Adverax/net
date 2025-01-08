package net

import (
	"context"
	"net/http"
	"time"
)

// MessengerWithTimeout is messenger with limited time to life
type MessengerWithTimeout struct {
	Messenger
	timeout time.Duration
}

func (that *MessengerWithTimeout) Request(ctx context.Context, request *http.Request) (resp *http.Response, err error) {
	ctx2, cancel := context.WithTimeout(ctx, that.timeout)
	defer cancel()

	request2 := request.WithContext(ctx2)
	resp, err = that.Messenger.Request(ctx2, request2)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// NewMessengerWithTimeout is constructor for create new instance of http timeout messenger
func NewMessengerWithTimeout(messenger Messenger, timeout time.Duration) *MessengerWithTimeout {
	return &MessengerWithTimeout{
		Messenger: messenger,
		timeout:   timeout,
	}
}
