---
status: committing
summary: Added LOG_LEVEL environment variable handler to service_main.go and updated CHANGELOG.md with feat entry under Unreleased section.
container: service-003-add-log-level-env-support
dark-factory-version: v0.156.1-1-g04f3863-dirty
created: "2026-05-14T09:22:10Z"
queued: "2026-05-14T09:22:10Z"
started: "2026-05-14T09:27:30Z"
---

<summary>
- `service.Main` already handles process-wide bootstrap concerns (timezone, glog setup, GOMAXPROCS, sentry, signal handling)
- Add a `LOG_LEVEL` environment variable handler at startup that calls `flag.Set("v", value)` so operators can change glog verbosity at runtime without rebuilding
- Currently each service's verbosity is baked into its Docker `ENTRYPOINT ["/main", "-v=2"]` — changing requires image rebuild + redeploy
- After this change, any service using the framework can set `LOG_LEVEL: "4"` in its K8s Config (or `-e LOG_LEVEL=4` in docker) and get verbose logs immediately on next pod start, no image rebuild
- Single 4-line addition to `service_main.go`, plus `os` import, plus one CHANGELOG bullet
</summary>

<objective>
Add a `LOG_LEVEL` environment variable handler to `service.Main` so operators can adjust glog verbosity via env at runtime, without rebuilding the image or changing the binary's `-v=N` flag.
</objective>

<context>
Read `CLAUDE.md` at the repo root for project conventions (Never code directly, dark-factory pipeline, make precommit required, no AI attribution in commits).

Read these guides:
- `/Users/bborbe/Documents/workspaces/coding/docs/git-commit-guide.md` — mandatory commit process
- `/Users/bborbe/Documents/workspaces/coding/docs/go-service-implementation-patterns.md` — Interface → Constructor → Struct → Method pattern
- `/Users/bborbe/Documents/workspaces/coding/docs/go-glog-guide.md` — glog verbosity (`-v=N`) and `glog.V(N)` semantics

**Key file to read in full**: `service_main.go` — the `Main` function at line 34. The new logic goes immediately AFTER line 44 (`_ = flag.Set("logtostderr", "true")`) and BEFORE line 46 (`time.Local = time.UTC`).

**Inline reference — current `service_main.go:41-47`:**

```go
defer glog.Flush()
glog.CopyStandardLogTo("info")
runtime.GOMAXPROCS(runtime.NumCPU())
_ = flag.Set("logtostderr", "true")

time.Local = time.UTC
glog.V(2).Infof("set global timezone to UTC")
```

`flag.Set("v", value)` works at any point — glog registers its `-v` flag at package init via `init()`, so the flag is available before `argument.ParseAndPrint` (line 49) runs. Setting LOG_LEVEL via `flag.Set` is equivalent to passing `-v=N` on the command line, and takes effect for all subsequent `glog.V(N).Info*` calls in the same process.
</context>

<requirements>

## 1. Add `os` import to `service_main.go`

The current import block has no `os`. Add it in alphabetical order (after `net/http`, before `runtime`):

```go
import (
    "context"
    "flag"
    "net/http"
    "os"
    "runtime"
    "time"

    "github.com/bborbe/argument/v2"
    "github.com/bborbe/run"
    libsentry "github.com/bborbe/sentry"
    "github.com/getsentry/sentry-go"
    "github.com/golang/glog"
)
```

## 2. Add LOG_LEVEL handler immediately after `flag.Set("logtostderr", "true")`

The new block goes between the existing `flag.Set("logtostderr", ...)` line and the `time.Local = time.UTC` line:

```go
_ = flag.Set("logtostderr", "true")

// LOG_LEVEL env var sets glog verbosity at runtime so operators can adjust
// log volume via K8s Config / docker -e without rebuilding the image's
// ENTRYPOINT -v=N flag. Equivalent to passing -v=<value> on the command line.
// Empty / unset = leave the default (or whatever -v=N argv specifies).
if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
    _ = flag.Set("v", logLevel)
}

time.Local = time.UTC
```

`_ = flag.Set(...)` ignores the error per the existing convention (line 44 does the same for `logtostderr`).

Do NOT log inside the LOG_LEVEL block — the verbosity it sets affects glog calls only AFTER this point, so a `glog.V(0).Infof("LOG_LEVEL set to %s", logLevel)` would print at any level (V=0 always logs). Skipping the log keeps the code minimal and matches the silent style of the `logtostderr` line above it.

## 3. Add CHANGELOG entry

Append to `CHANGELOG.md` under `## Unreleased` (create the section if it doesn't exist, above the most recent `## v` heading):

```markdown
- feat: `LOG_LEVEL` environment variable sets glog verbosity at runtime — equivalent to passing `-v=N` on the command line, but settable via K8s Config / docker `-e LOG_LEVEL=4` without rebuilding the image's `ENTRYPOINT -v=N` flag
```

## 4. Run tests + precommit

```bash
make precommit
```

Must exit 0. No new tests required — the addition is a single `flag.Set` call that mirrors the existing `logtostderr` pattern; the broader framework's existing tests cover `service.Main`.

</requirements>

<constraints>
- Change is confined to `service_main.go` and `CHANGELOG.md`. No other files modified. No new files created.
- `service.Main` signature is frozen.
- The LOG_LEVEL handler must be positioned AFTER `_ = flag.Set("logtostderr", "true")` and BEFORE `time.Local = time.UTC`. Position matters: the LOG_LEVEL must be applied early enough that subsequent `glog.V(2).Infof("set global timezone to UTC")` (line 47) respects the new verbosity.
- Use `_ = flag.Set("v", logLevel)` — ignore the error, matching the existing `_ = flag.Set("logtostderr", "true")` style at line 44.
- Do NOT add a glog message confirming LOG_LEVEL was set — the verbosity it controls affects only subsequent glog calls, so any confirmation would have to be at V=0 which prints at any level (unnecessary noise for the unset case).
- Empty / unset LOG_LEVEL leaves the existing `-v=N` argv flag value untouched (no override).
- Error wrapping convention from the repo: this code path doesn't have an error to wrap (flag.Set returns an error that we ignore per established convention).
- No `time.Now()`, no `context.Background()` in pkg/ code — N/A here.
- `make precommit` must exit 0.
- Do NOT commit — dark-factory handles git.
</constraints>

<verification>

Verify the import was added:
```bash
grep -n '"os"' service_main.go
```
Expected: one match in the import block.

Verify the LOG_LEVEL handler is present and correctly positioned:
```bash
awk '/flag.Set\("logtostderr"/{p=1} p && /flag.Set\("v"/{print NR": "$0; exit}' service_main.go
```
Expected: prints the `flag.Set("v", ...)` line with a line number greater than the `logtostderr` line.

Verify the handler reads LOG_LEVEL:
```bash
grep -n 'os.Getenv("LOG_LEVEL")' service_main.go
```
Expected: one match.

Verify position before `time.Local = time.UTC`:
```bash
awk '/flag.Set\("v"/{v=NR} /time.Local = time.UTC/{t=NR; if (v && t > v) print "OK: LOG_LEVEL block precedes time.UTC"; else print "FAIL"}' service_main.go
```
Expected: `OK: LOG_LEVEL block precedes time.UTC`.

Verify CHANGELOG updated:
```bash
grep -in "LOG_LEVEL" CHANGELOG.md | head -3
```
Expected: at least one match under `## Unreleased`.

Run precommit:
```bash
make precommit
```
Expected: exit 0.

</verification>
