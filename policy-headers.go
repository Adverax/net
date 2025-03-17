package net

import (
	"context"
	"github.com/adverax/policy"
)

// PolicyWithHeaders is policy with headers
type PolicyWithHeaders struct {
	policy  policy.Policy
	headers map[string]string
}

func NewPolicyWithHeaders(policy policy.Policy, headers map[string]string) *PolicyWithHeaders {
	return &PolicyWithHeaders{
		policy:  policy,
		headers: headers,
	}
}

func (that *PolicyWithHeaders) Execute(ctx context.Context, action policy.Action) error {
	delivery := HttpDeliveryFromContext(ctx)
	if delivery != nil {
		for key, val := range that.headers {
			delivery.Request.Header.Set(key, val)
		}
	}

	return that.policy.Execute(ctx, action)
}
