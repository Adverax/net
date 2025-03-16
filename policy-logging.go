package net

import (
	"bytes"
	"context"
	"fmt"
	"github.com/adverax/log"
	"github.com/adverax/policy"
	"net/http"
	"time"
)

// WithLogging is a policy that logs the action before executing it.
type WithLogging struct {
	policy.Policy
	logger  log.Logger
	entity  string
	headers bool
}

func NewWithLogging(
	policy policy.Policy,
	logger log.Logger,
	entity string,
	headers bool,
) *WithLogging {
	return &WithLogging{
		Policy:  policy,
		logger:  logger,
		entity:  entity,
		headers: headers,
	}
}

func (that *WithLogging) Execute(ctx context.Context, action policy.Action) error {
	delivery := HttpDeliveryFromContext(ctx)
	if delivery == nil {
		return that.Policy.Execute(ctx, action)
	}

	started := time.Now()
	that.logRequest(ctx, delivery.Request)
	err := that.Policy.Execute(ctx, action)
	if delivery.Response != nil {
		that.logResponse(ctx, delivery.Response, time.Since(started))
	}

	return err
}

func (that *WithLogging) logRequest(
	ctx context.Context,
	request *http.Request,
) {
	var err error
	var body []byte
	body, request.Body, err = DrainHttpBody(request.Body)
	if err != nil {
		return
	}

	var logger log.LoggerEntry = that.logger
	if that.headers {
		logger = logger.WithField("headers", that.getHeaders(request))
	}

	logger.
		WithFields(log.Fields{
			log.FieldKeyEntity:  that.entity,
			log.FieldKeyAction:  log.EntityActionOutcomeRequest,
			log.FieldKeyMethod:  request.Method,
			log.FieldKeySubject: request.URL.String(),
			log.FieldKeyData:    string(body),
		}).
		Info(ctx, "Request")
}

func (that *WithLogging) logResponse(
	ctx context.Context,
	resp *http.Response,
	elapsed time.Duration,
) {
	var err error
	var body []byte
	if resp.Request.Method == http.MethodGet || resp.Request.Method == http.MethodPost {
		body, resp.Body, err = DrainHttpBody(resp.Body)
		if err != nil {
			that.logger.Debug(ctx, fmt.Sprintf("logResponse get body error: %v", err))
			return
		}
	}

	that.logger.
		WithFields(log.Fields{
			log.FieldKeyEntity:   that.entity,
			log.FieldKeyAction:   log.EntityActionIncomeResponse,
			log.FieldKeyData:     string(body),
			log.FieldKeyDuration: elapsed,
		}).
		Info(ctx, "Response")
}

func (that *WithLogging) getHeaders(
	request *http.Request,
) string {
	var buffer bytes.Buffer
	_ = request.Header.Write(&buffer)
	return buffer.String()
}
