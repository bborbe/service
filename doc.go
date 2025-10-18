// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package service provides a standardized framework for building Go services
// with Sentry integration, structured logging, graceful shutdown, and signal handling.
//
// The framework handles cross-cutting concerns including:
//   - Command-line argument parsing via github.com/bborbe/argument
//   - Sentry error reporting with configurable exclusions
//   - Concurrent function execution with proper error handling
//   - Signal-based context cancellation
//   - Panic recovery and logging
//
// Example usage:
//
//	type app struct {
//		SentryDSN string `required:"true" arg:"sentry-dsn" env:"SENTRY_DSN"`
//	}
//
//	func (a *app) Run(ctx context.Context, sentryClient libsentry.Client) error {
//		return service.Run(ctx, a.createServer(), a.createWorker())
//	}
//
//	func main() {
//		app := &app{}
//		os.Exit(service.Main(context.Background(), app, &app.SentryDSN, nil))
//	}
package service
