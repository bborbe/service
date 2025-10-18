# Service

A Go service framework library that provides a standardized way to build HTTP services with Sentry integration, structured logging, and graceful shutdown capabilities.

## Features

- **Standardized Service Architecture**: Interface → Constructor → Struct → Method pattern
- **Sentry Integration**: Built-in error reporting and exception handling
- **Graceful Shutdown**: Context-aware service lifecycle management
- **HTTP Server Utilities**: Health checks, metrics, and routing with Gorilla Mux
- **Concurrent Execution**: Run multiple service components with proper error handling
- **Testing Support**: Counterfeiter mock generation and Ginkgo/Gomega test framework

## Quick Start

Example for creating a service can be found in `example/main.go`

```go
func main() {
    app := &application{}
    os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}
```

See the [example directory](example/) for a complete implementation.

## License

Copyright (c) 2024-2025 Benjamin Borbe

This project is licensed under the BSD-2-Clause License - see the [LICENSE](LICENSE) file for details.
