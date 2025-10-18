// Copyright (c) 2024-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	stderrors "errors"

	"github.com/bborbe/sentry"
)

// Options configures behavior for service execution, particularly Sentry error reporting.
// It allows customization of which errors should be excluded from Sentry reports.
type Options struct {
	ExcludeErrors sentry.ExcludeErrors
}

// OptionsFn is a functional option pattern function for configuring Options.
// It allows callers to customize service behavior by modifying the Options struct.
type OptionsFn func(option *Options)

// NewOptions creates a new Options instance with sensible defaults and applies the given
// functional options. By default, context.Canceled and context.DeadlineExceeded errors
// are excluded from Sentry reporting as they represent expected shutdown conditions.
func NewOptions(fns ...OptionsFn) Options {
	options := Options{
		ExcludeErrors: sentry.ExcludeErrors{
			func(err error) bool {
				return stderrors.Is(err, context.Canceled)
			},
			func(err error) bool {
				return stderrors.Is(err, context.DeadlineExceeded)
			},
		},
	}
	for _, fn := range fns {
		fn(&options)
	}
	return options
}
