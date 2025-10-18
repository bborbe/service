// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/argument/v2"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

// MainCmd initializes and runs a command-line application without Sentry integration.
// Unlike Main, this function is designed for CLI tools that do not require error reporting
// to Sentry and uses reduced logging verbosity (V(3) instead of V(0)). Returns an exit code:
// 0 for success, 1 for runtime error, 4 for argument parsing failure.
func MainCmd(
	ctx context.Context,
	app run.Runnable,
) int {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	_ = flag.Set("logtostderr", "true")

	time.Local = time.UTC
	glog.V(2).Infof("set global timezone to UTC")

	if err := argument.Parse(ctx, app); err != nil {
		glog.Errorf("parse app failed: %v", err)
		return 4
	}

	glog.V(3).Infof("application started")
	if err := app.Run(run.ContextWithSig(ctx)); err != nil {
		glog.Error(err)
		return 1
	}
	glog.V(3).Infof("application finished")
	return 0
}
