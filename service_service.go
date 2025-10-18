// Copyright (c) 2024-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/service.go --fake-name Service . Service

// Service wraps application execution with Sentry error reporting and structured logging.
// It provides a layer between the main entry point and the application logic, handling
// error capture and reporting to Sentry automatically.
type Service interface {
	Run(ctx context.Context) error
}

// NewService creates a new Service instance that wraps the given application with Sentry integration.
// The sentryClient will be used for error reporting, and the app will receive it for use in business logic.
func NewService(
	sentryClient libsentry.Client,
	app Application,
) Service {
	return &service{
		app:          app,
		sentryClient: sentryClient,
	}
}

type service struct {
	sentryClient libsentry.Client
	app          Application
}

func (s *service) Run(ctx context.Context) error {
	if err := s.app.Run(ctx, s.sentryClient); err != nil {
		s.sentryClient.CaptureException(
			err,
			&sentry.EventHint{
				Context:           ctx,
				OriginalException: err,
			},
			sentry.NewScope(),
		)
		return errors.Wrap(ctx, err, "application failed")
	}
	glog.V(4).Infof("run finished without error")
	return nil
}
