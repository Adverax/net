package net

import (
	"context"
	"github.com/adverax/policy"
)

type WithHeadersExecution struct {
	policy  policy.Policy
	headers map[string]string
}

func NewWithHeadersExecution(policy policy.Policy, headers map[string]string) *WithHeadersExecution {
	return &WithHeadersExecution{
		policy:  policy,
		headers: headers,
	}
}

func (that *WithHeadersExecution) Execute(ctx context.Context, action policy.Action) error {
	delivery := HttpDeliveryFromContext(ctx)
	if delivery != nil {
		for key, val := range that.headers {
			delivery.Request.Header.Set(key, val)
		}
	}

	return that.policy.Execute(ctx, action)
}
