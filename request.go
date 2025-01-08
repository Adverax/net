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
}

type value struct {
	value interface{}
	codec Codec
}

func (that *value) encode() ([]byte, error) {
	return that.codec.Encode(that.value)
}

func (that *value) decode(data []byte) error {
	if that.value == nil {
		return nil
	}

	return that.codec.Decode(data, that.value)
}

type Validator interface {
	Validate(status int, body []byte) error
}

type ValidatorFunc func(status int, body []byte) error

func (f ValidatorFunc) Validate(status int, body []byte) error {
	return f(status, body)
}

type Request struct {
	url       string
	method    string
	params    map[string]string
	headers   map[string]string
	messenger Messenger
	body      value
	response  value
	validator Validator
	raw       *[]byte
}

func NewRequest() *Request {
	return &Request{
		method:    http.MethodGet,
		validator: defValidator,
		body: value{
			codec: CodecDefault,
		},
		response: value{
			codec: CodecDefault,
		},
	}
}

func (that *Request) Request(messenger Messenger, method string, url string) *Request {
	that.messenger = messenger
	that.method = method
	that.url = ensureProtocol(url)
	return that
}

func (that *Request) Body(body interface{}, codec Codec) *Request {
	that.body.value = body
	that.body.codec = codec
	return that
}

func (that *Request) Response(response interface{}, codec Codec) *Request {
	that.response.value = response
	that.response.codec = codec
	return that
}

func (that *Request) Param(key, value string) *Request {
	if that.params == nil {
		that.params = make(map[string]string)
	}
	that.params[key] = value
	return that
}

func (that *Request) Params(params map[string]string) *Request {
	if that.params == nil {
		that.params = make(map[string]string)
	}
	for key, val := range params {
		that.params[key] = val
	}
	return that
}

func (that *Request) Header(key, val string) *Request {
	if that.headers == nil {
		that.headers = make(map[string]string)
	}
	that.headers[key] = val
	return that
}

func (that *Request) Headers(hs map[string]string) *Request {
	if that.headers == nil {
		that.headers = make(map[string]string)
	}
	for key, val := range hs {
		that.headers[key] = val
	}
	return that
}

func (that *Request) Raw(raw *[]byte) *Request {
	that.raw = raw
	return that
}

func (that *Request) Validator(validator Validator) *Request {
	that.validator = validator
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
		defer resp.Body.Close()
	}

	err = that.validator.Validate(resp.StatusCode, body)
	if err != nil {
		return err
	}

	if that.raw != nil {
		*that.raw = body
	}

	err = that.response.decode(body)
	if err != nil {
		return err
	}

	return nil
}

func (that *Request) newRequest(ctx context.Context) (*http.Request, error) {
	body, err := that.body.encode()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, that.method, that.getUrl(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if that.headers != nil {
		for key, val := range that.headers {
			request.Header.Set(key, val)
		}
	}

	request.Close = true

	return request, nil
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
		if that.body.value == nil {
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

type defaultValidator struct {
}

func (that *defaultValidator) Validate(status int, body []byte) error {
	if status == http.StatusOK {
		return nil
	}

	return &HttpError{
		Code: status,
		Text: fmt.Sprintf("response invalid status (%d) >> %s", status, string(body)),
	}
}

var defValidator = &defaultValidator{}
