# Service

[![Go Reference](https://pkg.go.dev/badge/github.com/bborbe/service.svg)](https://pkg.go.dev/github.com/bborbe/service)
[![CI](https://github.com/bborbe/service/actions/workflows/ci.yml/badge.svg)](https://github.com/bborbe/service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bborbe/service)](https://goreportcard.com/report/github.com/bborbe/service)

A Go service framework library that provides a standardized way to build HTTP services with Sentry integration, structured logging, and graceful shutdown capabilities.

## Features

- **Standardized Service Architecture**: Interface → Constructor → Struct → Method pattern
- **Sentry Integration**: Built-in error reporting and exception handling
- **Graceful Shutdown**: Context-aware service lifecycle management
- **HTTP Server Utilities**: Health checks, metrics, and routing with Gorilla Mux
- **Concurrent Execution**: Run multiple service components with proper error handling
- **Automatic Error Handling**: Panic recovery, error logging, and Sentry reporting
- **Argument Parsing**: Type-safe CLI argument parsing with struct tags
- **Testing Support**: Counterfeiter mock generation and Ginkgo/Gomega test framework

## Installation

```bash
go get github.com/bborbe/service
```

## Quick Start

Here's a minimal example to get started:

```go
package main

import (
    "context"
    "os"

    libsentry "github.com/bborbe/sentry"
    "github.com/bborbe/service"
)

type application struct {
    SentryDSN   string `required:"true" arg:"sentry-dsn" env:"SENTRY_DSN"`
    SentryProxy string `arg:"sentry-proxy" env:"SENTRY_PROXY"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
    // Your application logic here
    <-ctx.Done()
    return ctx.Err()
}

func main() {
    app := &application{}
    os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}
```

Run with:
```bash
./myservice --sentry-dsn="https://key@sentry.io/project"
```

## Core Concepts

### Application Interface

Your application must implement the `Application` interface:

```go
type Application interface {
    Run(ctx context.Context, sentryClient libsentry.Client) error
}
```

The framework handles:
- **Argument Parsing**: Automatically parses CLI arguments and environment variables from struct tags
- **Sentry Setup**: Configures error reporting with the provided DSN
- **Global Settings**: Sets timezone to UTC, configures logging, and GOMAXPROCS
- **Graceful Shutdown**: Manages context cancellation and signal handling
- **Exit Codes**: Returns appropriate exit codes (0=success, 1=runtime error, 2=Sentry failure, 3=missing DSN, 4=argument error)

### Running Multiple Components

Use `service.Run()` to execute multiple concurrent functions with automatic error handling:

```go
func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
    return service.Run(
        ctx,
        a.createHTTPServer(),
        a.createBackgroundWorker(),
    )
}
```

Each function automatically receives:
- **Panic Recovery**: Automatic panic catching and error conversion
- **Error Logging**: Structured logging with glog
- **Error Filtering**: Excludes expected errors (like `context.Canceled`) from Sentry
- **Context Cancellation**: Stops all functions when any one finishes or fails

### Argument Parsing

Use struct tags to define CLI arguments and environment variables:

```go
type application struct {
    SentryDSN string `required:"true" arg:"sentry-dsn" env:"SENTRY_DSN" usage:"Sentry DSN for error reporting"`
    Listen    string `required:"true" arg:"listen" env:"LISTEN" usage:"Address to listen on"`
}
```

Supported tags:
- `required`: Whether the argument is mandatory
- `arg`: CLI flag name (e.g., `--listen`)
- `env`: Environment variable name
- `usage`: Help text description
- `display`: How to display value in logs (`"length"` masks sensitive data)

## Full Example

Complete HTTP service with health checks and metrics:

```go
package main

import (
    "context"
    "os"

    libhttp "github.com/bborbe/http"
    "github.com/bborbe/run"
    libsentry "github.com/bborbe/sentry"
    "github.com/bborbe/service"
    "github.com/golang/glog"
    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    app := &application{}
    os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
    SentryDSN   string `required:"true"  arg:"sentry-dsn"   env:"SENTRY_DSN"   display:"length"`
    SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY"`
    Listen      string `required:"true"  arg:"listen"       env:"LISTEN"       usage:"address to listen to"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
    return service.Run(
        ctx,
        a.createBackgroundWorker(),
        a.createHTTPServer(),
    )
}

func (a *application) createBackgroundWorker() run.Func {
    return func(ctx context.Context) error {
        // Background task logic
        <-ctx.Done()
        return ctx.Err()
    }
}

func (a *application) createHTTPServer() run.Func {
    return func(ctx context.Context) error {
        ctx, cancel := context.WithCancel(ctx)
        defer cancel()

        router := mux.NewRouter()
        router.Path("/healthz").Handler(libhttp.NewPrintHandler("OK"))
        router.Path("/readiness").Handler(libhttp.NewPrintHandler("OK"))
        router.Path("/metrics").Handler(promhttp.Handler())

        glog.V(2).Infof("starting http server listen on %s", a.Listen)
        return libhttp.NewServer(a.Listen, router).Run(ctx)
    }
}
```

Run with:
```bash
./myservice --sentry-dsn="https://key@sentry.io/project" --listen=":8080"
```

The service provides:
- Health check endpoint: `http://localhost:8080/healthz`
- Readiness endpoint: `http://localhost:8080/readiness`
- Prometheus metrics: `http://localhost:8080/metrics`

## API Documentation

For complete API documentation, visit [pkg.go.dev/github.com/bborbe/service](https://pkg.go.dev/github.com/bborbe/service).

## Development

### Running Tests

```bash
make test
```

Tests run with race detection enabled and coverage reporting.

### Code Generation

Generate mocks for testing:

```bash
make generate
```

This creates Counterfeiter mocks for all interfaces in the `mocks/` directory.

### Full Development Workflow

Before committing changes:

```bash
make precommit
```

This runs:
- Dependency verification
- Code formatting (gofmt, goimports-reviser, golines)
- Code generation (mocks)
- Tests with race detection and coverage
- Static analysis (vet, errcheck, lint)
- Security scanning (gosec, trivy, osv-scanner, govulncheck)
- License header verification

## Testing Your Service

Example test using Counterfeiter mocks:

```go
package myapp_test

import (
    "context"
    "testing"

    libsentry "github.com/bborbe/sentry"
    "github.com/bborbe/service"
    "github.com/bborbe/service/mocks"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "MyApp Suite")
}

var _ = Describe("Application", func() {
    var (
        app          *application
        sentryClient *mocks.SentryClient
    )

    BeforeEach(func() {
        sentryClient = &mocks.SentryClient{}
        app = &application{
            Listen: ":8080",
        }
    })

    It("should run successfully", func() {
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // Test your application logic
        err := app.Run(ctx, sentryClient)
        Expect(err).To(BeNil())
    })
})
```

## Exit Codes

The framework returns standard exit codes:

- `0`: Success
- `1`: Runtime error
- `2`: Sentry setup failure
- `3`: Missing Sentry DSN
- `4`: Argument parsing failure

## Dependencies

Key runtime dependencies:

- [github.com/bborbe/argument/v2](https://github.com/bborbe/argument) - Type-safe argument parsing
- [github.com/bborbe/sentry](https://github.com/bborbe/sentry) - Sentry error reporting wrapper
- [github.com/bborbe/run](https://github.com/bborbe/run) - Concurrent function execution
- [github.com/bborbe/http](https://github.com/bborbe/http) - HTTP server utilities
- [github.com/gorilla/mux](https://github.com/gorilla/mux) - HTTP routing
- [github.com/onsi/ginkgo/v2](https://github.com/onsi/ginkgo) - BDD testing framework
- [github.com/onsi/gomega](https://github.com/onsi/gomega) - Matcher library

## License

Copyright (c) 2024-2025 Benjamin Borbe

This project is licensed under the BSD-2-Clause License - see the [LICENSE](LICENSE) file for details.
