// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"
	"errors"

	libsentry "github.com/bborbe/sentry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/service"
)

var _ = Describe("Options", func() {
	Context("NewOptions", func() {
		It("creates options with default values", func() {
			opts := service.NewOptions()

			Expect(opts).NotTo(BeNil())
			Expect(opts.ExcludeErrors).NotTo(BeNil())
		})

		It("applies functional options", func() {
			customExclude := func(err error) bool {
				return err.Error() == "custom error"
			}

			opts := service.NewOptions(
				func(o *service.Options) {
					o.ExcludeErrors = libsentry.ExcludeErrors{customExclude}
				},
			)

			Expect(opts.ExcludeErrors).To(HaveLen(1))
		})

		It("applies multiple functional options in order", func() {
			exclude1 := func(err error) bool {
				return err.Error() == "error 1"
			}
			exclude2 := func(err error) bool {
				return err.Error() == "error 2"
			}

			opts := service.NewOptions(
				func(o *service.Options) {
					o.ExcludeErrors = libsentry.ExcludeErrors{exclude1}
				},
				func(o *service.Options) {
					o.ExcludeErrors = append(o.ExcludeErrors, exclude2)
				},
			)

			Expect(opts.ExcludeErrors).To(HaveLen(2))
		})

		It("works with no options", func() {
			opts := service.NewOptions()

			Expect(opts).NotTo(BeNil())
			// Should have default excluded errors (context.Canceled, context.DeadlineExceeded)
			Expect(opts.ExcludeErrors).To(HaveLen(2))
		})
	})

	Context("Default ExcludeErrors", func() {
		It("excludes context.Canceled by default", func() {
			opts := service.NewOptions()

			// Test that context.Canceled is excluded
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(context.Canceled) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeTrue(), "Should exclude context.Canceled")
		})

		It("excludes context.DeadlineExceeded by default", func() {
			opts := service.NewOptions()

			// Test that context.DeadlineExceeded is excluded
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(context.DeadlineExceeded) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeTrue(), "Should exclude context.DeadlineExceeded")
		})

		It("does not exclude other errors by default", func() {
			opts := service.NewOptions()
			testErr := errors.New("test error")

			// Test that random errors are not excluded
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(testErr) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeFalse(), "Should not exclude random errors")
		})
	})

	Context("Customizing Options", func() {
		It("allows complete replacement of ExcludeErrors", func() {
			customExclude := func(err error) bool {
				return err.Error() == "custom"
			}

			opts := service.NewOptions(
				func(o *service.Options) {
					o.ExcludeErrors = libsentry.ExcludeErrors{customExclude}
				},
			)

			Expect(opts.ExcludeErrors).To(HaveLen(1))
		})

		It("allows appending to default ExcludeErrors", func() {
			customExclude := func(err error) bool {
				return err.Error() == "custom"
			}

			opts := service.NewOptions(
				func(o *service.Options) {
					o.ExcludeErrors = append(o.ExcludeErrors, customExclude)
				},
			)

			// Should have 2 default + 1 custom = 3 total
			Expect(opts.ExcludeErrors).To(HaveLen(3))

			// Verify context.Canceled is still excluded
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(context.Canceled) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeTrue())
		})

		It("can exclude custom errors", func() {
			myErr := errors.New("my custom error")
			customExclude := func(err error) bool {
				return errors.Is(err, myErr)
			}

			opts := service.NewOptions(
				func(o *service.Options) {
					o.ExcludeErrors = append(o.ExcludeErrors, customExclude)
				},
			)

			// Test the custom exclusion works
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(myErr) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeTrue())
		})
	})

	Context("OptionsFn type", func() {
		It("can be used to create reusable option functions", func() {
			addCustomExclude := func(errMsg string) service.OptionsFn {
				return func(o *service.Options) {
					excludeFn := func(err error) bool {
						return err.Error() == errMsg
					}
					o.ExcludeErrors = append(o.ExcludeErrors, excludeFn)
				}
			}

			opts := service.NewOptions(
				addCustomExclude("error 1"),
				addCustomExclude("error 2"),
			)

			// Should have 2 default + 2 custom = 4 total
			Expect(opts.ExcludeErrors).To(HaveLen(4))

			// Test that custom errors are excluded
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")

			excluded1 := false
			excluded2 := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(err1) {
					excluded1 = true
				}
				if excludeFn(err2) {
					excluded2 = true
				}
			}
			Expect(excluded1).To(BeTrue())
			Expect(excluded2).To(BeTrue())
		})
	})

	Context("Integration scenarios", func() {
		It("handles wrapped context errors correctly", func() {
			opts := service.NewOptions()

			// Create a wrapped context.Canceled error
			wrappedErr := errors.Join(context.Canceled, errors.New("during shutdown"))

			// Should still be excluded because it wraps context.Canceled
			excluded := false
			for _, excludeFn := range opts.ExcludeErrors {
				if excludeFn(wrappedErr) {
					excluded = true
					break
				}
			}
			Expect(excluded).To(BeTrue(), "Should exclude wrapped context.Canceled")
		})

		It("can combine multiple exclusion strategies", func() {
			opts := service.NewOptions(
				func(o *service.Options) {
					// Add exclusion for specific error messages
					o.ExcludeErrors = append(o.ExcludeErrors, func(err error) bool {
						return err.Error() == "expected error"
					})
				},
				func(o *service.Options) {
					// Add exclusion for error type
					o.ExcludeErrors = append(o.ExcludeErrors, func(err error) bool {
						var targetErr *customError
						return errors.As(err, &targetErr)
					})
				},
			)

			// Should have 2 default + 2 custom = 4 total
			Expect(opts.ExcludeErrors).To(HaveLen(4))
		})
	})
})

// customError is a test error type
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
