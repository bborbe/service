// Copyright (c) 2024-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"flag"
	"net/http"
	"runtime"
	"time"

	"github.com/bborbe/argument/v2"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/service-application.go --fake-name ServiceApplication . Application

// Application defines the contract for services that can be executed with Sentry integration.
// Implementations receive a configured Sentry client for error reporting and should implement
// the Run method to contain the application's business logic.
type Application interface {
	Run(ctx context.Context, sentryClient libsentry.Client) error
}

// Main initializes and runs the service application with Sentry integration.
// It handles argument parsing, timezone configuration (UTC), Sentry setup, signal handling,
// and graceful shutdown. Returns an exit code: 0 for success, 1 for runtime error,
// 2 for Sentry setup failure, 3 for missing Sentry DSN, 4 for argument parsing failure.
func Main(
	ctx context.Context,
	app Application,
	sentryDSN *string,
	sentryProxy *string,
	fns ...OptionsFn,
) int {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	_ = flag.Set("logtostderr", "true")

	time.Local = time.UTC
	glog.V(2).Infof("set global timezone to UTC")

	if err := argument.ParseAndPrint(ctx, app); err != nil {
		glog.Errorf("parse app failed: %v", err)
		return 4
	}

	options := NewOptions(fns...)

	if sentryDSN == nil {
		glog.Errorf("sentryDSN args missing")
		return 3
	}
	httpTransport := http.DefaultTransport
	if sentryProxy != nil {
		httpTransport = libsentry.NewProxyRoundTripper(
			httpTransport,
			*sentryProxy,
		)
		glog.V(2).Infof("use sentryProxy %s", *sentryProxy)
	}
	sentryClient, err := libsentry.NewClient(
		ctx,
		sentry.ClientOptions{
			Dsn:              *sentryDSN,
			TracesSampleRate: 1.0,
			HTTPTransport:    httpTransport,
		},
		options.ExcludeErrors...,
	)
	if err != nil {
		glog.Errorf("setting up Sentry failed: %+v", err)
		return 2
	}
	defer func() {
		_ = sentryClient.Flush(2 * time.Second)
		_ = sentryClient.Close()
	}()

	service := NewService(
		sentryClient,
		app,
	)

	glog.V(0).Infof("application started")
	if err := service.Run(run.ContextWithSig(ctx)); err != nil {
		glog.Error(err)
		return 1
	}
	glog.V(0).Infof("application finished")
	return 0
}
