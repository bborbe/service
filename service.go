// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	stderrors "errors"

	"github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

type ExcludeErrors []ExcludeError

func (e ExcludeErrors) IsExcluded(err error) bool {
	for _, ee := range e {
		if ee(err) {
			return true
		}
	}
	return false
}

type ExcludeError func(err error) bool

type Option struct {
	ExcludeErrors ExcludeErrors
}

type OptionFn func(option *Option)

//counterfeiter:generate -o mocks/service.go --fake-name Service . Service
type Service interface {
	Run(ctx context.Context) error
}

func NewService(
	sentryClient libsentry.Client,
	app Application,
	fns ...OptionFn,
) Service {
	serviceOption := Option{
		ExcludeErrors: ExcludeErrors{
			func(err error) bool {
				return stderrors.Is(err, context.Canceled)
			},
			func(err error) bool {
				return stderrors.Is(err, context.DeadlineExceeded)
			},
		},
	}
	for _, fn := range fns {
		fn(&serviceOption)
	}
	return &service{
		app:          app,
		sentryClient: sentryClient,
		options:      serviceOption,
	}
}

type service struct {
	sentryClient libsentry.Client
	app          Application
	options      Option
}

func (s *service) Run(ctx context.Context) error {
	err := s.app.Run(ctx, s.sentryClient)
	if err == nil {
		glog.V(4).Infof("run finished without error")
		return nil
	}
	if s.options.ExcludeErrors.IsExcluded(err) {
		glog.V(4).Infof("run finished with error, but is excluded")
		return nil
	}
	s.sentryClient.CaptureException(err, &sentry.EventHint{
		Context: ctx,
	}, nil)
	return errors.Wrapf(ctx, err, "application failed")
}
