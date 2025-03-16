package net

import (
	"context"
	"net/http"
)

type HttpMethod string

const (
	MethodGet    HttpMethod = "GET"
	MethodPost   HttpMethod = "POST"
	MethodPatch  HttpMethod = "PATCH"
	MethodPut    HttpMethod = "PUT"
	MethodDelete HttpMethod = "DELETE"
)

type httpDeliveryKey int // context key for http delivery

const httpDeliveryKeyKey httpDeliveryKey = 0

type HttpError struct {
	Code int
	Text string
}

func (that *HttpError) Error() string {
	return that.Text
}

type HttpDelivery struct {
	Request  *http.Request
	Response *http.Response
}

func HttpDeliveryFromContext(ctx context.Context) *HttpDelivery {
	v := ctx.Value(httpDeliveryKeyKey)
	if v == nil {
		return nil
	}
	return v.(*HttpDelivery)
}
