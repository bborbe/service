// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"
	"errors"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/service"
)

var _ = Describe("Run", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	Context("with successful functions", func() {
		It("runs all functions and returns nil when context is canceled", func() {
			var mu sync.Mutex
			executed := make(map[string]bool)

			fn1 := func(ctx context.Context) error {
				mu.Lock()
				executed["fn1"] = true
				mu.Unlock()
				<-ctx.Done()
				return ctx.Err()
			}

			fn2 := func(ctx context.Context) error {
				mu.Lock()
				executed["fn2"] = true
				mu.Unlock()
				<-ctx.Done()
				return ctx.Err()
			}

			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			err := service.Run(ctx, fn1, fn2)

			Expect(err).To(BeNil())
			mu.Lock()
			defer mu.Unlock()
			Expect(executed["fn1"]).To(BeTrue())
			Expect(executed["fn2"]).To(BeTrue())
		})
	})

	Context("with function returning error", func() {
		It("stops all functions and returns the error", func() {
			expectedErr := errors.New("test error")

			fn1 := func(ctx context.Context) error {
				return expectedErr
			}

			fn2 := func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			}

			err := service.Run(ctx, fn1, fn2)

			Expect(err).To(MatchError(expectedErr))
		})
	})

	Context("with function panicking", func() {
		It("recovers from panic", func() {
			fn1 := func(ctx context.Context) error {
				panic("test panic")
			}

			// Should not panic - the panic is recovered
			Expect(func() {
				_ = service.Run(ctx, fn1)
			}).NotTo(Panic())
		})
	})

	Context("with context.Canceled error", func() {
		It("filters out context.Canceled errors", func() {
			fn1 := func(ctx context.Context) error {
				<-ctx.Done()
				return context.Canceled
			}

			fn2 := func(ctx context.Context) error {
				<-ctx.Done()
				return context.Canceled
			}

			cancel()

			err := service.Run(ctx, fn1, fn2)

			Expect(err).To(BeNil())
		})
	})

	Context("with mixed errors", func() {
		It("returns non-filtered error even if context.Canceled also occurs", func() {
			expectedErr := errors.New("real error")

			fn1 := func(ctx context.Context) error {
				return expectedErr
			}

			fn2 := func(ctx context.Context) error {
				<-ctx.Done()
				return context.Canceled
			}

			err := service.Run(ctx, fn1, fn2)

			Expect(err).To(MatchError(expectedErr))
		})
	})

	Context("with no functions", func() {
		It("returns nil immediately", func() {
			err := service.Run(ctx)

			Expect(err).To(BeNil())
		})
	})

	Context("with single function", func() {
		It("runs the function correctly", func() {
			executed := false

			fn := func(ctx context.Context) error {
				executed = true
				<-ctx.Done()
				return ctx.Err()
			}

			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			err := service.Run(ctx, fn)

			Expect(err).To(BeNil())
			Expect(executed).To(BeTrue())
		})
	})
})

var _ = Describe("FilterErrors", func() {
	Context("with context.Canceled error", func() {
		It("returns a function that filters out context.Canceled", func() {
			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return context.Canceled
				},
				context.Canceled,
			)

			err := fn(context.Background())

			Expect(err).To(BeNil())
		})
	})

	Context("with context.DeadlineExceeded error", func() {
		It("filters out context.DeadlineExceeded when specified", func() {
			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return context.DeadlineExceeded
				},
				context.DeadlineExceeded,
			)

			err := fn(context.Background())

			Expect(err).To(BeNil())
		})
	})

	Context("with non-filtered error", func() {
		It("returns the error unchanged", func() {
			expectedErr := errors.New("test error")

			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return expectedErr
				},
				context.Canceled,
			)

			err := fn(context.Background())

			Expect(err).To(Equal(expectedErr))
		})
	})

	Context("with nil error", func() {
		It("returns nil", func() {
			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return nil
				},
				context.Canceled,
			)

			err := fn(context.Background())

			Expect(err).To(BeNil())
		})
	})

	Context("with multiple filters", func() {
		It("filters out all specified errors", func() {
			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return context.Canceled
				},
				context.Canceled,
				context.DeadlineExceeded,
			)

			err := fn(context.Background())

			Expect(err).To(BeNil())
		})

		It("returns error if not in filter list", func() {
			expectedErr := errors.New("test error")

			fn := service.FilterErrors(
				func(ctx context.Context) error {
					return expectedErr
				},
				context.Canceled,
				context.DeadlineExceeded,
			)

			err := fn(context.Background())

			Expect(err).To(Equal(expectedErr))
		})
	})
})
