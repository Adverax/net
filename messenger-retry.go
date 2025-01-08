package net

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type RetryMessengerErrorChecker interface {
	IsRetryableError(err error) bool
}

type RetryMessengerMetrics interface {
	IncSuccess()
	IncFailure()
	IncAttempts()
}

type RetryMessengerOptions struct {
	InitialInterval       time.Duration
	BackoffCoefficient    float64
	MaximumInterval       time.Duration
	MaximumAttempts       int
	RetryableErrorChecker RetryMessengerErrorChecker
	Metrics               RetryMessengerMetrics
}

type retryState struct {
	interval time.Duration
	attempts int
}

type RetryMessenger struct {
	options RetryMessengerOptions
	Messenger
}

// NewRetryMessenger is constructor for creating messenger with retry policy
func NewRetryMessenger(messenger Messenger, options RetryMessengerOptions) *RetryMessenger {
	return &RetryMessenger{
		options:   options,
		Messenger: messenger,
	}
}

func (that *RetryMessenger) Request(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := that.Messenger.Request(ctx, req)
	if err == nil {
		that.success()
		return resp, nil
	}

	if that.options.MaximumAttempts < 0 {
		that.failure()
		return nil, err
	}

	resp, err = that.retry(ctx, req, err)
	if err == nil {
		that.success()
		return resp, nil
	}

	that.failure()
	return nil, err
}

func (that *RetryMessenger) retry(ctx context.Context, req *http.Request, err error) (*http.Response, error) {
	state := retryState{
		interval: that.options.InitialInterval,
	}
	for that.canAttempt(err, &state) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(state.interval):
		}

		state.attempts++
		that.attempt()

		var resp *http.Response
		resp, err = that.Messenger.Request(ctx, req)
		if err == nil {
			return resp, nil
		}
	}
	return nil, err
}

func (that *RetryMessenger) canAttempt(err error, state *retryState) bool {
	if !that.IsRetryableError(err) {
		return false
	}

	if state.attempts >= that.options.MaximumAttempts && that.options.MaximumAttempts != 0 {
		return false
	}

	interval := time.Duration(that.options.BackoffCoefficient * float64(state.interval))
	if interval > that.options.MaximumInterval {
		state.interval = that.options.MaximumInterval
	} else {
		state.interval = interval
	}

	return true
}

func (that *RetryMessenger) success() {
	if that.options.Metrics != nil {
		that.options.Metrics.IncSuccess()
	}
}

func (that *RetryMessenger) failure() {
	if that.options.Metrics != nil {
		that.options.Metrics.IncFailure()
	}
}

func (that *RetryMessenger) attempt() {
	if that.options.Metrics != nil {
		that.options.Metrics.IncAttempts()
	}
}

func (that *RetryMessenger) IsRetryableError(err error) bool {
	return that.options.RetryableErrorChecker == nil || that.options.RetryableErrorChecker.IsRetryableError(err)
}

type retryErrorChecker struct {
	nonRetryableErrors []error
}

func NewRetryErrorChecker(nonRetryableErrors []error) RetryMessengerErrorChecker {
	return &retryErrorChecker{
		nonRetryableErrors: nonRetryableErrors,
	}
}

func (that *retryErrorChecker) IsRetryableError(err error) bool {
	for _, nonRetryableError := range that.nonRetryableErrors {
		if errors.Is(err, nonRetryableError) {
			return false
		}
	}
	return true
}
