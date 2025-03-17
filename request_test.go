package net

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

type MessengerMock struct {
	mock.Mock
	req *http.Request
}

func (that *MessengerMock) Request(ctx context.Context, request *http.Request) (*http.Response, error) {
	args := that.Called(ctx, request)
	that.req = request
	return args.Get(0).(*http.Response), args.Error(1)
}

func (that *MessengerMock) NewRequest() *Request {
	return NewRequest().
		WithMessenger(that).
		WithCodec(CodecJson)
}

func TestRequest(t *testing.T) {
	req := []string{"a", "b", "c"}
	var resp []string

	msngr := &MessengerMock{}
	msngr.On("Request", mock.Anything, mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`["a","b","c"]`))),
		}, nil)

	err := msngr.NewRequest().
		WithRequest(MethodPost, "http://example.com").
		WithBody(req).
		WithResponse(&resp).
		Send()
	require.NoError(t, err)
	require.NotNil(t, msngr.req)
	body, err := io.ReadAll(msngr.req.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), body)
	assert.Equal(t, req, resp)
}
