// Copyright (c) 2024-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"errors"

	"github.com/bborbe/run"
)

// Run executes multiple functions concurrently with automatic error logging, panic recovery,
// and context.Canceled error filtering. Each function is wrapped with logging, panic recovery,
// and error filtering middleware. Returns on first function completion using the
// CancelOnFirstFinishWait strategy, which cancels all other functions when any one completes.
func Run(ctx context.Context, funcs ...run.Func) error {
	for i, fn := range funcs {
		funcs[i] = run.LogErrors(
			run.CatchPanic(
				FilterErrors(
					fn,
					context.Canceled,
				),
			),
		)
	}
	return run.CancelOnFirstFinishWait(ctx, funcs...)
}

// FilterErrors wraps a run.Func to suppress specified errors, returning nil instead of the error
// if it matches any of the provided filteredErrors using errors.Is. This is useful for filtering
// out expected errors like context.Canceled during graceful shutdown.
func FilterErrors(fn run.Func, filteredErrors ...error) run.Func {
	return func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			for _, filteredError := range filteredErrors {
				if errors.Is(err, filteredError) {
					return nil
				}
			}
			return err
		}
		return nil
	}
}
