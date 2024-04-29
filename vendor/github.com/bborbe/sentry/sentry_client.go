// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentry

import (
	"context"
	"io"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/sentry-client.go --fake-name SentryClient . Client
type Client interface {
	CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	CaptureException(exception error, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	io.Closer
}

func NewClient(ctx context.Context, clientOptions sentry.ClientOptions) (Client, error) {
	newClient, err := sentry.NewClient(clientOptions)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create sentry client failed")
	}
	return &client{
		client: newClient,
	}, nil
}

type client struct {
	client *sentry.Client
}

func (c *client) CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID {
	eventID := c.client.CaptureMessage(message, hint, scope)
	glog.V(2).Infof("capture sentry message with id %s: %s", *eventID, message)
	return eventID
}

func (c *client) CaptureException(err error, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		glog.V(2).Infof("skip error: %v", err)
		return nil
	}
	eventID := c.client.CaptureException(err, hint, scope)
	glog.V(2).Infof("capture sentry execption with id %s: %v", *eventID, err)
	return eventID
}

func (c *client) Close() error {
	c.client.Flush(2 * stdtime.Second)
	return nil
}
