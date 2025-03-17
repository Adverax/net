package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	Headers() map[string]string
}

type value struct {
	data  interface{}
	codec Codec
}

func (that *value) encode() ([]byte, error) {
	return that.codec.Encode(that.data)
}

func (that *value) decode(data []byte) error {
	if that.data == nil {
		return nil
	}

	return that.codec.Decode(data, that.data)
}

func (that *value) Headers() map[string]string {
	return that.codec.Headers()
}

type ResponseHandler interface {
	Handle(ctx context.Context, resp Response) error
}

type ResponseHandlerFunc func(ctx context.Context, resp Response) error

func (fn ResponseHandlerFunc) Handle(ctx context.Context, resp Response) error {
	return fn(ctx, resp)
}

type Response interface {
	StatusCode() int
	ResponseBody() []byte
	Decode(data interface{}) error
}

type Request struct {
	url        string
	method     HttpMethod
	params     map[string]string
	headers    map[string]string
	messenger  Messenger
	codec      Codec
	body       interface{}
	response   interface{}
	handlers   map[int]ResponseHandler
	respBody   []byte
	statusCode int
}

func NewRequest() *Request {
	return &Request{
		method: http.MethodGet,
		codec:  CodecDefault,
	}
}

func (that *Request) StatusCode() int {
	return that.statusCode
}

func (that *Request) ResponseBody() []byte {
	return that.respBody
}

func (that *Request) Decode(data interface{}) error {
	return that.codec.Decode(that.respBody, data)
}

func (that *Request) WithMessenger(messenger Messenger) *Request {
	that.messenger = messenger
	return that
}

func (that *Request) WithCodec(codec Codec) *Request {
	that.codec = codec
	return that
}

func (that *Request) WithRequest(method HttpMethod, url string) *Request {
	that.method = method
	that.url = ensureProtocol(url)
	return that
}

func (that *Request) WithBody(body interface{}) *Request {
	that.body = body
	return that
}

func (that *Request) WithResponse(response interface{}) *Request {
	that.response = response
	return that
}

func (that *Request) WithHandler(status int, handler ResponseHandler) *Request {
	if that.handlers == nil {
		that.handlers = make(map[int]ResponseHandler)
	}
	that.handlers[status] = handler
	return that
}

func (that *Request) WithHandlers(handlers map[int]ResponseHandler) *Request {
	if that.handlers == nil {
		that.handlers = make(map[int]ResponseHandler)
	}
	for status, handler := range handlers {
		that.handlers[status] = handler
	}
	return that
}

func (that *Request) WithParam(key, value string) *Request {
	if that.params == nil {
		that.params = make(map[string]string)
	}
	that.params[key] = value
	return that
}

func (that *Request) WithParams(params map[string]string) *Request {
	if that.params == nil {
		that.params = make(map[string]string)
	}
	for key, val := range params {
		that.params[key] = val
	}
	return that
}

func (that *Request) WithHeader(key, val string) *Request {
	if that.headers == nil {
		that.headers = make(map[string]string)
	}
	that.headers[key] = val
	return that
}

func (that *Request) WithHeaders(hs map[string]string) *Request {
	if that.headers == nil {
		that.headers = make(map[string]string)
	}
	for key, val := range hs {
		that.headers[key] = val
	}
	return that
}

func (that *Request) Send() error {
	return that.SendContext(context.Background())
}

func (that *Request) SendContext(ctx context.Context) error {
	if err := that.checkRequiredFields(); err != nil {
		return fmt.Errorf("checkRequiredFields: %w", err)
	}

	request, err := that.newRequest(ctx)
	if err != nil {
		return err
	}

	return that.sendContext(ctx, request)
}

func (that *Request) sendContext(
	ctx context.Context,
	request *http.Request,
) error {
	resp, err := that.messenger.Request(ctx, request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	var body []byte
	if resp.Body != nil {
		var err error
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	that.respBody = body
	that.statusCode = resp.StatusCode

	err = that.handleResponse(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (that *Request) newRequest(ctx context.Context) (*http.Request, error) {
	body, err := that.codec.Encode(that.body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, string(that.method), that.getUrl(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if that.headers != nil {
		for key, val := range that.headers {
			request.Header.Set(key, val)
		}
	}

	headers := that.codec.Headers()
	if headers != nil {
		for key, val := range headers {
			request.Header.Set(key, val)
		}
	}

	request.Close = true

	return request, nil
}

func (that *Request) handleResponse(ctx context.Context) error {
	if that.handlers != nil {
		handler, ok := that.handlers[that.statusCode]
		if ok {
			return handler.Handle(ctx, that)
		}
	}

	if that.statusCode == http.StatusOK {
		if that.response == nil {
			return nil
		}

		return that.codec.Decode(that.respBody, that.response)
	}

	return nil
}

func (that *Request) checkRequiredFields() error {
	if that.url == "" {
		return ErrFieldUrlRequired
	}
	if that.messenger == nil {
		return ErrFieldMessengerRequired
	}

	switch that.method {
	case http.MethodPost:
		if that.body == nil {
			return ErrFieldBodyRequired
		}
	}

	return nil
}

func (that *Request) getUrl() string {
	if that.params == nil {
		return that.url
	}

	params := url.Values{}
	for key, val := range that.params {
		params.Set(key, val)
	}

	return fmt.Sprintf("%s?%s", that.url, params.Encode())
}

var (
	ErrFieldUrlRequired       = fmt.Errorf("url is required")
	ErrFieldMessengerRequired = fmt.Errorf("messenger is required")
	ErrFieldBodyRequired      = fmt.Errorf("body is required")
)

func ensureProtocol(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "http://" + url
}
