package net

import (
	"github.com/adverax/policy"
	"net/http"
)

type MessengerBuilder struct {
	messenger *messenger
}

func NewMessengerBuilder() *MessengerBuilder {
	return &MessengerBuilder{
		messenger: &messenger{},
	}
}

func (that *MessengerBuilder) WithClient(client HttpClient) *MessengerBuilder {
	that.messenger.client = client
	return that
}

func (that *MessengerBuilder) WithPolicy(policy policy.Policy) *MessengerBuilder {
	that.messenger.policy = policy
	return that
}

func (that *MessengerBuilder) WithCodec(codec Codec) *MessengerBuilder {
	that.messenger.codec = codec
	return that
}

func (that *MessengerBuilder) Build() (Messenger, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}

	if err := that.updateDefaultFields(); err != nil {
		return nil, err
	}

	return that.messenger, nil
}

func (that *MessengerBuilder) checkRequiredFields() error {
	return nil
}

func (that *MessengerBuilder) updateDefaultFields() error {
	if that.messenger.client == nil {
		that.messenger.client = http.DefaultClient
	}
	if that.messenger.policy == nil {
		that.messenger.policy = policy.NewDefault()
	}
	if that.messenger.codec == nil {
		that.messenger.codec = CodecDefault
	}

	return nil
}
