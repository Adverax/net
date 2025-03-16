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
}

type messenger struct {
	client HttpClient
	policy policy.Policy
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

func NewMessenger(client HttpClient, p policy.Policy) Messenger {
	if client == nil {
		client = http.DefaultClient
	}
	if p == nil {
		p = policy.NewDefault()
	}

	return &messenger{
		client: client,
		policy: p,
	}
}
