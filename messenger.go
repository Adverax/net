package net

import (
	"context"
	"github.com/adverax/policy"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Messenger is abstract http manager
type Messenger interface {
	Request(ctx context.Context, request *http.Request) (resp *http.Response, err error)
	NewRequest() *Request
}

type messenger struct {
	client HttpClient
	policy policy.Policy
	codec  Codec
}

func (that *messenger) NewRequest() *Request {
	return NewRequest().
		WithMessenger(that).
		WithCodec(that.codec)
}

func (that *messenger) Request(ctx context.Context, request *http.Request) (resp *http.Response, err error) {
	delivery := &HttpDelivery{
		Request: request,
	}

	ctx = context.WithValue(ctx, httpDeliveryKeyKey, delivery)

	err = that.policy.Execute(
		ctx,
		policy.ActionFunc(func(ctx context.Context) error {
			delivery.Response, err = that.client.Do(request)
			return err
		}),
	)

	return delivery.Response, err
}
