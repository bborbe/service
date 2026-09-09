---
status: completed
summary: 'Guarded Sentry proxy selection on value not pointer: empty/nil SENTRY_PROXY now sends events direct via http.DefaultTransport with the chosen path logged at glog.V(0), plus a white-box DescribeTable and proxy-rewrite boundary test'
execution_id: service-sentry-proxy-guard-exec-005-sentry-proxy-guard-value-not-pointer
dark-factory-version: dev
created: "2026-09-09T15:59:40Z"
queued: "2026-09-09T15:59:40Z"
started: "2026-09-09T16:00:04Z"
completed: "2026-09-09T16:02:25Z"
---

# Guard Sentry proxy on the value, not the pointer

<summary>
- Services deployed without `SENTRY_PROXY` no longer silently lose all error reporting
- The transport selection now checks whether the proxy value is non-empty, not merely whether a pointer exists, so an empty config falls back to the standard HTTP transport
- The chosen transport path is logged at default verbosity (visible without `-v=2`) in both branches — proxy used, or direct
- The transport-selection logic is extracted into a small helper so it can be unit-tested in isolation
- A table test covers all three cases: nil pointer, empty string, populated string
- No behavior change for services that do configure a proxy
</summary>

<objective>
Stop the silent, total loss of Sentry error reporting for services deployed without `SENTRY_PROXY`: the guard in `Main` currently tests the pointer (never nil), so an empty value builds a proxy round tripper whose `url.Parse("")` succeeds and then blanks `req.URL.Host`/`Scheme` on every event. End state: empty or nil proxy config sends events direct via `http.DefaultTransport`, with the chosen path visible in logs at default verbosity.
</objective>

<context>
Read `service_main.go` — the `Main` function builds `httpTransport` before calling `libsentry.NewClient` (see the `if sentryProxy != nil {` block around `httpTransport := http.DefaultTransport`).

The proxy round tripper lives in `github.com/bborbe/sentry` (`NewProxyRoundTripper(roundtripper http.RoundTripper, url string) http.RoundTripper`); an empty URL produces a round tripper that always fails to deliver. The project logs with `github.com/golang/glog` — `glog.Infof` is default verbosity (visible without `-v`), `glog.V(2).Infof` is not.

Test conventions: Ginkgo/Gomega with the external test package `service_test` (see `service_options_test.go`, `service_run_test.go`, `service_suite_test.go`). `Describe`/`Context`/`It` blocks, `Expect(...).To(...)`.
</context>

<requirements>
1. In `service_main.go`, extract the transport selection into a helper function `sentryHTTPTransport(sentryProxy *string) http.RoundTripper` (package-level, unexported):
   - Default is `http.DefaultTransport`.
   - If `sentryProxy != nil && *sentryProxy != ""`: return `libsentry.NewProxyRoundTripper(http.DefaultTransport, *sentryProxy)` and log the proxy path at default verbosity (e.g. `glog.V(0).Infof("use sentryProxy %s", *sentryProxy)`, matching the file's existing default-verbosity style at lines 96/101).
   - Otherwise: return `http.DefaultTransport` and log the direct path at default verbosity (e.g. `glog.V(0).Infof("send sentry events direct via http.DefaultTransport (no sentry proxy configured)")`).
   - The log call must be `glog.V(0).Infof` (default verbosity) — NOT `glog.V(2).Infof` — so the choice is visible without `-v=2`.
2. Replace the inline transport block in `Main` (the current `if sentryProxy != nil { ... glog.V(2).Infof("use sentryProxy %s", *sentryProxy) }` block) with a single call: `httpTransport := sentryHTTPTransport(sentryProxy)`. Remove the now-duplicated inline log line. Do not change anything else in `Main` (no change to `sentryDSN` handling, `NewClient` call, flush/close, or exit codes).
3. Add a test in a new file `service_main_test.go` using a Ginkgo `DescribeTable` covering the three cases:
   - nil pointer → returned transport `BeIdenticalTo(http.DefaultTransport)`
   - non-nil pointer to empty string `""` → returned transport `BeIdenticalTo(http.DefaultTransport)`
   - non-nil pointer to populated string e.g. `"http://localhost:8080"` → returned transport `NotTo(BeIdenticalTo(http.DefaultTransport))`
   - The helper `sentryHTTPTransport` is unexported, so this file MUST be white-box: declare `package service` (not `package service_test`). `Main`'s transport is not observable from outside the package, so do NOT attempt to assert it through `service.Main`. Do NOT add any exported test seam.
   - Because the existing `TestSuite`/`RunSpecs` in `service_suite_test.go` lives in `package service_test`, a white-box file's Ginkgo specs will NOT run without their own entry point: add a `TestSentryHTTPTransport(t *testing.T)` function in `service_main_test.go` that calls `RegisterFailHandler(Fail)` and `RunSpecs(t, "Sentry HTTP Transport Suite")`. Otherwise the three-case table silently never executes while `make precommit` passes.
   - Add a fourth boundary case that crosses the `url.Parse` boundary the bug is about: construct the populated transport (`sentryHTTPTransport` with `"http://localhost:8080"`), call `transport.RoundTrip(request)` on a real `http.Request` (e.g. `http.NewRequest("POST", "https://o1.ingest.sentry.io/api/123/envelope/", strings.NewReader("{}"))`) wrapped with a short context deadline (`context.WithTimeout`, e.g. 5s), and assert `request.URL.Host == "localhost:8080"` and `request.URL.Scheme == "http"` after the call (the request is mutated before the inner round tripper runs, so the assertion holds regardless of the network outcome; the deadline prevents a stall if localhost:8080 ever accepts without responding). This locks the proxy-rewrite behavior the fix protects.
4. Add a `## Unreleased` section at the top of `CHANGELOG.md` (above `## v1.10.13`) with one entry describing the fix, e.g. `- fix: guard sentry proxy on value not pointer — services without SENTRY_PROXY now send events direct instead of silently failing`.
5. Do not touch any other file. Do not change `go.mod`/`go.sum`.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Keep the change minimal: one helper + `Main` call-site swap + one new test file + CHANGELOG entry
- Do not change the behavior for populated proxy configs
- Do not rename or restructure anything outside the transport-selection block
- Repo-relative paths only
</constraints>

<verification>
Run `make precommit` — must pass.
Confirm the diff is limited to: `service_main.go` (helper + call-site swap), `service_main_test.go` (new, white-box table test + boundary case + `TestSentryHTTPTransport` entry point), `CHANGELOG.md` (Unreleased entry). Confirm `glog.V(0).Infof` (default verbosity) is used for the path logs, no `glog.V(2)` remains for the sentry proxy path, and generated mocks are unchanged.
</verification>
