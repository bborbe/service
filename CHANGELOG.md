# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.8.1

- fix data race in service_run_test.go by adding mutex synchronization
- update dependencies (argument v2.8.0, collection v1.14.0, validation v1.3.3, bbolt v1.4.3)
- remove unused actgardner/gogen-avro dependency

## v1.8.0

- add comprehensive test coverage (0% → 32.4% with 37 passing tests)
- add tests for Run(), FilterErrors(), NewOptions(), and Service interface
- expand README documentation (31 → 309 lines) with installation, examples, and API docs
- add badges to README (Go Reference, CI, Go Report Card)
- enable race detection in test target
- update glog dependency to fix security vulnerability
- add golang.org/x/tools v0.38.0 exclude for counterfeiter compatibility
- update golang.org/x/tools from v0.38.0 to v0.37.0

## v1.7.1

- fix counterfeiter directive placement to exclude from GoDoc output

## v1.7.0

- add complete GoDoc documentation for all exported items
- add package-level documentation (doc.go)
- add Ginkgo CLI tool to tools.go for consistent test execution
- remove deprecated golang.org/x/lint/golint from tools.go
- fix error wrapping (errors.Wrapf → errors.Wrap when no formatting arguments)
- update copyright years to match git history (2024-2025 year ranges)
- update LICENSE file to 2024-2025 year range
- update CI workflow Go version from 1.25.2 to 1.25.3
- add license section to README.md
- fix example code linter issues (createHttpServer → createHTTPServer, simplify select statement)
- remove legacy // +build comment from tools.go

## v1.6.4

- update README with comprehensive description and features
- improve documentation with quick start guide

## v1.6.3

- add example service main
- go mod update
- add LICENSE

## v1.6.2

- fix MainCmd

## v1.6.1

- MainBasic -> MainCmd

## v1.6.0

- add MainBasic

## v1.5.0

- remove vendor
- go mod update

## v1.4.0

- remove set v=2
- go mod update

## v1.3.1

- go mod update

## v1.3.0

- refactor exclude sentry errors

## v1.2.2

- add original error to sentry exeception

## v1.2.1

- go mod update

## v1.2.0

- flush sentry before exit

## v1.1.1

- fix sentryProxy

## v1.1.0

- add sentryProxy argument

## v1.0.0

- Initial Version
