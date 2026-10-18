---
status: completed
spec: [001-transient-error-exclusion]
summary: Added a structural Transient() bool predicate to the default Sentry ExcludeErrors list, with tests, count updates, and a CHANGELOG entry.
execution_id: service-transient-exec-004-transient-error-exclusion
dark-factory-version: dev
created: "2026-10-18T13:20:00Z"
queued: "2026-10-18T13:33:58Z"
started: "2026-10-18T13:35:35Z"
completed: "2026-10-18T13:37:19Z"
---

# Exclude transient-marked errors from Sentry by default

<summary>
- The framework's default Sentry error-exclusion list learns to recognise an error a library has marked as an expected, self-resolving condition
- Such an error is dropped before it becomes an event, for every service built on the framework, with no per-service change
- The marker is a structural interface, so the framework gains no dependency on any domain library
- The two existing defaults are preserved, an unmarked error is still reported, and a caller can still replace or append to the list
- The default list grows by exactly one entry, so four existing test expectations each move by one
</summary>

<objective>
Teach the service framework's default Sentry error-exclusion list to recognise an error carrying a marker whose `Transient()` method returns true, so a reporter can drop expected transient conditions without the framework importing the library that raised them. Existing behaviour is otherwise unchanged.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read `service_options.go` — `Options`, `OptionsFn`, and `NewOptions`, which builds the default `ExcludeErrors` list (today two predicates: `context.Canceled`, `context.DeadlineExceeded`).
Read `service_options_test.go` — the Ginkgo/Gomega suite. Note which assertions encode the **default list length** and which are replacement-based (see requirement 5).
Read `service_main.go` — where `options.ExcludeErrors...` is passed to the Sentry client. Do not modify it.
</context>

<requirements>
1. Add one predicate to the default `ExcludeErrors` list built by `NewOptions` in `service_options.go`. It returns true iff `errors.As` finds, in the error's wrapped chain, a value satisfying the structural interface `interface{ Transient() bool }` whose method returns true (first match wins — `errors.As` stops at the first assignable value). Consistent with the two existing defaults, which use `errors.Is`.

2. The predicate must be **structural and text-free**. Do not match on error message content — no `strings.Contains` on `err.Error()`, no literal `"EOF"` or `"ECONNREFUSED"`. The authoritative check is in `<verification>`: `! grep -qE 'connection refused|EOF|ECONNREFUSED' service_options.go`. (Note `grep -n` also exits 1 on zero matches, so if you run that pattern directly, exit 1 means *clean*, not failed.)

3. Do not import any domain library (in particular not `github.com/bborbe/kafka`) and do not add a shared marker package. The marker is satisfied implicitly by Go's structural typing. `go.mod` and `go.sum` must be unchanged.

4. The two existing default predicates (`context.Canceled`, `context.DeadlineExceeded`) must remain present and unchanged.

5. The default list grows by exactly one entry, so **four** existing assertions must each move by exactly one, and the three replacement-based assertions must NOT change:
   - `It("works with no options")` — `HaveLen(2)` → `HaveLen(3)`, and its comment currently enumerating two defaults must name three.
   - `It("allows appending to default ExcludeErrors")` — `HaveLen(3)` → `HaveLen(4)`; comment `2 default + 1 custom = 3 total` → `3 default + 1 custom = 4 total`.
   - `It("can be used to create reusable option functions")` — `HaveLen(4)` → `HaveLen(5)`; comment `2 default + 2 custom = 4 total` → `3 default + 2 custom = 5 total`.
   - `It("can combine multiple exclusion strategies")` — `HaveLen(4)` → `HaveLen(5)`; comment `2 default + 2 custom = 4 total` → `3 default + 2 custom = 5 total`.
   ⚠️ Do NOT change `It("applies functional options")` (`HaveLen(1)`), `It("applies multiple functional options in order")` (`HaveLen(2)`), or `It("allows complete replacement of ExcludeErrors")` (`HaveLen(1)`) — those replace the list before appending and are independent of the default length.

6. Add tests in `service_options_test.go` using Ginkgo/Gomega, matching the existing suite:
   - A local fake type implementing `Transient() bool { return true }`, wrapped two levels deep with `github.com/bborbe/errors` (`errors.Wrapf(ctx, err, "level 1")` then `errors.Wrapf(ctx, err, "level 2")` — alias it `liberrors`, because the test file already imports stdlib `errors`), is excluded by the default list.
   - A local fake whose `Transient()` returns `false` is NOT excluded by any default predicate.
   - A plain unmarked error is NOT excluded — the existing `It("does not exclude other errors by default")` must still pass.
   - The existing `It("excludes context.Canceled by default")`, `It("excludes context.DeadlineExceeded by default")` and `It("handles wrapped context errors correctly")` must still pass unmodified.
   - For the new assertions, prefer `opts.ExcludeErrors.IsExcluded(err)` — the method the Sentry client itself calls — rather than a hand-rolled loop. Keep the existing loop style where you are not adding assertions.

7. Add a CHANGELOG entry. `CHANGELOG.md` has **no** `## Unreleased` section today (it jumps straight to `## v1.10.8`), so create one immediately above `## v1.10.8` with:
   `- feat: default Sentry exclusion list drops errors a library has marked transient via a structural `Transient() bool` interface`

8. Do not add a second classification mechanism (sampling, dedup, rate limiting, fingerprinting), and do not change the replace/append semantics of `ExcludeErrors`.

Before you finish: re-run `<verification>` and confirm it passes; walk each requirement above against the change.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Do NOT import any domain library or introduce a shared marker package (requirement 3).
- Do NOT match on error text (requirement 2).
- Do NOT change the two existing defaults or the replace/append semantics.
- Do NOT change the four default-count assertions by more than exactly one each, and do not touch the three replacement-based assertions.
- Do NOT modify `service_main.go` or the Sentry client — the exclusion is already applied and already checked before an event is built.
- Do NOT wrap the predicate in `recover()`. A panic in an implementer's `Transient()` is a defect in that library (spec failure-modes row 3), not something this framework masks.
- Existing tests must still pass.
</constraints>

<verification>
ROOTDIR=/workspace make precommit
grep -q 'Transient' service_options.go
! grep -q 'Transient' service_main.go
! grep -qE 'connection refused|EOF|ECONNREFUSED' service_options.go
grep -q 'Transient' CHANGELOG.md

All five must pass. `make precommit` must exit 0. The positive `grep -q` lines exit 0 on a match, asserting the predicate landed in the options file and that the entry point was left alone; the negative line asserts no error-text matching was introduced, and is written `! grep -q` because `grep -c`/`grep -n` exit 1 on a zero count and would fail the step they were meant to pass. (`ROOTDIR=/workspace` is the canonical container form; this repo's `Makefile` includes only `tools.env` and does not consume `ROOTDIR`, so the prefix is harmless rather than load-bearing. No git command appears in this verification.)
</verification>
