// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"
	"errors"

	liberrors "github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	sentrymocks "github.com/bborbe/sentry/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/service"
	"github.com/bborbe/service/mocks"
)

var _ = Describe("Service", func() {
	var (
		sentryClient *sentrymocks.SentryClient
		application  *mocks.ServiceApplication
		svc          service.Service
		ctx          context.Context
	)

	BeforeEach(func() {
		sentryClient = &sentrymocks.SentryClient{}
		application = &mocks.ServiceApplication{}
		svc = service.NewService(sentryClient, application)
		ctx = context.Background()
	})

	Context("NewService", func() {
		It("creates a service instance", func() {
			Expect(svc).NotTo(BeNil())
		})

		It("returns Service interface type", func() {
			var _ service.Service = svc
		})
	})

	Context("Run", func() {
		Context("when application runs successfully", func() {
			BeforeEach(func() {
				application.RunReturns(nil)
			})

			It("calls application.Run with context and sentry client", func() {
				err := svc.Run(ctx)

				Expect(err).To(BeNil())
				Expect(application.RunCallCount()).To(Equal(1))
				passedCtx, passedSentry := application.RunArgsForCall(0)
				Expect(passedCtx).To(Equal(ctx))
				Expect(passedSentry).To(Equal(sentryClient))
			})

			It("does not capture error to Sentry", func() {
				err := svc.Run(ctx)

				Expect(err).To(BeNil())
				Expect(sentryClient.CaptureExceptionCallCount()).To(Equal(0))
			})
		})

		Context("when application returns error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("application error")
				application.RunReturns(expectedErr)
			})

			It("returns wrapped error", func() {
				err := svc.Run(ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("application failed"))
			})

			It("captures error to Sentry", func() {
				err := svc.Run(ctx)

				Expect(err).NotTo(BeNil())
				Expect(sentryClient.CaptureExceptionCallCount()).To(Equal(1))
				capturedErr, _, _ := sentryClient.CaptureExceptionArgsForCall(0)
				Expect(capturedErr).To(MatchError(expectedErr))
			})
		})

		Context("when application returns wrapped error", func() {
			BeforeEach(func() {
				baseErr := errors.New("base error")
				wrappedErr := liberrors.Wrap(ctx, baseErr, "wrapped")
				application.RunReturns(wrappedErr)
			})

			It("captures and returns the wrapped error", func() {
				err := svc.Run(ctx)

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("wrapped"))
				Expect(sentryClient.CaptureExceptionCallCount()).To(Equal(1))
			})
		})

		Context("when application returns context.Canceled", func() {
			BeforeEach(func() {
				application.RunReturns(context.Canceled)
			})

			It("still captures to Sentry (filtering happens at Run level)", func() {
				err := svc.Run(ctx)

				Expect(err).To(HaveOccurred())
				Expect(sentryClient.CaptureExceptionCallCount()).To(Equal(1))
			})
		})
	})

	Context("Integration with real application", func() {
		It("executes application logic correctly", func() {
			executed := false

			realApp := &testApplication{
				runFunc: func(ctx context.Context, client libsentry.Client) error {
					executed = true
					return nil
				},
			}

			svc := service.NewService(sentryClient, realApp)
			err := svc.Run(ctx)

			Expect(err).To(BeNil())
			Expect(executed).To(BeTrue())
		})

		It("propagates errors from application", func() {
			expectedErr := errors.New("test error")

			realApp := &testApplication{
				runFunc: func(ctx context.Context, client libsentry.Client) error {
					return expectedErr
				},
			}

			svc := service.NewService(sentryClient, realApp)
			err := svc.Run(ctx)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("application failed"))
		})

		It("receives sentry client in application", func() {
			var receivedClient libsentry.Client

			realApp := &testApplication{
				runFunc: func(ctx context.Context, client libsentry.Client) error {
					receivedClient = client
					return nil
				},
			}

			svc := service.NewService(sentryClient, realApp)
			_ = svc.Run(ctx)

			Expect(receivedClient).To(Equal(sentryClient))
		})
	})
})

// testApplication is a simple implementation of service.Application for testing
type testApplication struct {
	runFunc func(context.Context, libsentry.Client) error
}

func (t *testApplication) Run(ctx context.Context, client libsentry.Client) error {
	if t.runFunc != nil {
		return t.runFunc(ctx, client)
	}
	return nil
}
