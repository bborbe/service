---
status: approved
approved: "2026-10-18T10:25:36Z"
branch: dark-factory/transient-error-exclusion
---

## Summary

- Teach the service framework's **default** Sentry error-exclusion list to recognise an error a library has marked as an expected, self-resolving condition — so such an error is dropped before it becomes an event, for every service built on the framework, with no per-service change.
- The mechanism already exists and is already deployed: the framework builds a default exclusion list, applies it when the Sentry client is constructed, and the Sentry wrapper checks it *before* building an event — an excluded error costs no event id and no quota.
- The marker is a **structural Go interface**, so the framework gains **no dependency** on any domain library. A library marks its own errors; the framework matches the marker. Neither imports the other.
- The two existing defaults are preserved unchanged, an unmarked error is still reported exactly as today, and a caller can still replace or append to the list.
- This is the second half of a two-spec change. The first — the Kafka library marking its transient broker-unreachable errors — is `bborbe/kafka` *transient-consumer-error-marker*. Neither half changes behaviour alone; the runtime proof belongs to the owning vault task.

## Problem

A single Kafka broker outage makes every consuming pod report its own Sentry error for the same event. On 2026-10-12 a broker became unreachable and **158 distinct trading-service pods** each emitted an error for it, producing 251 distinct `event_id`s in one hour. The burst exhausted `sentry-proxy`'s global rate-limit window, so a large block of alerts was rejected before Sentry ever stored it — those alerts got no issue id and could not become verdicts.

The fan-out is per-pod by construction: the consumer collapses every partition error into one returned error, and the framework's reporter captures exactly once per run — so the event count equals the replica count. No per-process rate limiter can reduce it; a limiter caps a single emission, which is already one.

The reporter boundary already has the right shape for this. The default exclusion list exists precisely to drop expected, self-resolving conditions before they become events, and it already carries two such defaults. What it lacks is any way to recognise a *transient network* class — an error the raising library has marked as expected. Today the only way a consumer could suppress one is to pass its own predicate, which means editing every consumer's `main` function instead of one default list.

## Goal

The framework's default error-exclusion list contains a predicate that returns `true` for an error carrying a marker whose `Transient()` method returns `true`, and `false` for every other error. The framework therefore drops such errors from Sentry for every service built on it, with no per-service change and no new dependency. All existing behaviour is unchanged: the two existing defaults still apply, an unmarked error is still reported, and a caller can still replace or append to the list.

## Non-goals

- Do NOT import the Kafka library, the Sentry wrapper's error types, or any domain package. The predicate is structural — that is the whole point of the design.
- Do NOT match on error **text** (`connection refused`, `EOF`, …). A text match is broad and would suppress those strings for every service on the framework, marked or not.
- Do NOT add a second classification mechanism (sampling, dedup, rate limiting, fingerprinting). The exclusion list is the existing seam; this spec uses it.
- Do NOT change the two existing defaults or their semantics.
- Do NOT change the list's replace/append semantics — a caller passing its own list still fully replaces the defaults.
- Do NOT touch the Kafka library. The marker is implemented there, in the sibling spec.
- Do NOT change the framework's main entry point, the Sentry client, or the Sentry wrapper — the exclusion is already applied and already checked before event construction.
- Do NOT suppress the transient at the **consumer** boundary. Making the Kafka consumer's error handler return `nil` is a frozen anti-pattern in that library (it stalls the partition at an invalid offset). This spec suppresses only the *report*.

## Acceptance Criteria

- [ ] The default exclusion list excludes an error carrying a marker whose `Transient()` returns `true` — evidence: unit test declares a local fake type implementing the marker with `Transient() bool { return true }`, passes it to the default list, and asserts at least one default predicate returns `true`; `go test -run <TestName> ./...` exits 0.
- [ ] The marker is found through a wrapped chain, not only on the outermost error — evidence: the same unit test wraps the fake two levels deep and asserts exclusion; `go test` exits 0.
- [ ] An error whose marker returns `false` is NOT excluded — evidence: unit test uses a fake whose `Transient()` returns `false` and asserts no default predicate returns `true`; `go test` exits 0. (Negative AC — kills a hardcoded `return true`.)
- [ ] An unmarked error is still reported — evidence: the existing `does not exclude other errors by default` test still passes with a plain unmarked error; `go test` exits 0. (Negative AC — no regression.)
- [ ] The two existing defaults are unchanged — evidence: the existing `excludes context.Canceled by default` and `excludes context.DeadlineExceeded by default` tests pass unmodified, and `handles wrapped context errors correctly` still passes; `go test` exits 0.
- [ ] **All four** default-count assertions move by exactly one, and no other assertion changes — evidence: `service_options_test.go:66` `HaveLen(2)`→`HaveLen(3)`, `:142` `HaveLen(3)`→`HaveLen(4)`, `:196` `HaveLen(4)`→`HaveLen(5)`, `:253` `HaveLen(4)`→`HaveLen(5)`, with the `2 default + …` comments at `:141`, `:195`, `:252` corrected to `3 default + …`, and the default-enumeration comment at `:65` (`// Should have default excluded errors (context.Canceled, context.DeadlineExceeded)`) updated to name all three defaults — note `:65` carries no `2 default + …` text, so a literal search for that string there will not find it; `:38`, `:58` and `:127` (`HaveLen(1)`, `HaveLen(2)`, `HaveLen(1)`) are **replacement-based and must NOT change**; `go test` exits 0.
- [ ] Complete replacement still works — evidence: `service_options_test.go:116` `allows complete replacement of ExcludeErrors` still passes with `HaveLen(1)`; `go test` exits 0.
- [ ] The framework gains no new module dependency — evidence: `git diff go.mod` returns empty after the change.
- [ ] The predicate lives in the default list, not in the entry point — evidence: `grep -n 'Transient' service_options.go` returns ≥1 match AND `grep -n 'Transient' service_main.go` returns 0 lines.
- [ ] No error-text matching was introduced — evidence: `grep -nE 'connection refused|ErrUnexpectedEOF|io\.EOF' service_options.go` returns 0 lines.
- [ ] `make precommit` exits 0 in the service package — evidence: exit code 0.

## Verification

**Rung 1 — container-executable (runs inside the YOLO container at prompt time):**

```
make precommit
```

Expected: exit code 0. The new in-package unit tests pass — a marked error (direct, and wrapped two levels) is excluded by a default predicate; a fake whose marker returns `false` is not excluded; a plain unmarked error is not excluded. The four default-count assertions read `HaveLen(3)`, `HaveLen(4)`, `HaveLen(5)`, `HaveLen(5)`; the three replacement-based assertions still read `HaveLen(1)`, `HaveLen(2)`, `HaveLen(1)`. `git diff go.mod` is empty; the entry point contains no `Transient` reference; no error-text pattern appears in the options file.

**Rung 2 — operator-executable, deferred to a named owning task (NOT satisfiable in this spec):**

The runtime proof — that a forced broker loss against bumped consumers produces no per-pod Sentry fan-out, and that the `KafkaConsumerLag` alert fires in its place — runs on the cluster, after the Kafka sibling spec lands and consumers are bumped to both libraries. Owning task: **Suppress the Trading Fleet's Kafka Consumer-Error Storm on Broker Outage** (vault task `c6d987db-384a-4961-8e4e-affc60ef95a6`), subtasks *nuke dev* and *nuke prod*; expected result is ≤ 1 distinct `server_name` per incident against the 158-`server_name` baseline, **and** the replacement signal firing. **Completion semantics: this spec does NOT reach `completed` on Rung 1 alone** — it stays `verifying` until those readings pass. This rung is not satisfiable inside `verify-spec`: it needs a deployed consumer and a real broker loss.

## Desired Behavior

1. The framework's options constructor returns an exclusion list containing a predicate that returns `true` iff the error — or any error in its wrapped chain — carries a marker satisfying the structural interface `interface{ Transient() bool }` whose method returns `true`.
2. The predicate walks the chain with the standard library's `errors.As`, consistent with how the two existing defaults use `errors.Is`.
3. The two existing default predicates (`context.Canceled`, `context.DeadlineExceeded`) remain present and unchanged.
4. An error satisfying none of the defaults is not excluded, and is reported by the entry point exactly as today.
5. A caller supplying its own exclusion list still fully replaces the defaults; a caller appending still appends. The default list grows by exactly one entry.
6. The service module's `go.mod` is unchanged — the predicate references only the standard library and the marker interface it declares inline.
7. The entry point's wiring is unchanged: it still passes the exclusion list to the Sentry client, which still checks it before constructing an event.

## Assumptions

- Go's implicit interface satisfaction makes the contract usable without a shared package: a library implements the marker method and the predicate finds it via `errors.As`, with no import in either direction.
- The Sentry wrapper checks the exclusion list before building the event and returns `nil` on a match (`bborbe/sentry` `sentry_client.go:154-157`), so an excluded error costs no event id and no quota. Read from the current source, not assumed.
- The framework's main entry point is the reporter for every pod in the storm, so a default-list change reaches all of them on a version bump — no per-service wiring needed.
- **Four existing assertions encode the default list's length, not one** — `service_options_test.go:66`, `:142`, `:196`, `:253`, each moving by exactly one. Three further assertions (`:38`, `:58`, `:127`) are replacement-based and unaffected. This was verified against `origin/master`, not the local checkout.
- Until the Kafka sibling spec lands, the new predicate matches nothing in practice (no error implements the marker yet). That is safe: an unmatched predicate changes no behaviour.
- Line references in this spec are given against `origin/master`; the local working copy may be behind. Prefer the symbol names (`NewOptions`, the `ExcludeErrors` field, the entry point's client construction) over line numbers when reading.

## Constraints

- **The predicate must be structural and text-free.** Matching on `connection refused` / `EOF` strings would suppress those errors for every service on the framework, marked or not — a far wider blast radius than the marker contract the operator chose. `grep -nE 'connection refused|ErrUnexpectedEOF|io\.EOF' service_options.go` must return 0 lines.
- The service module must not import the Kafka library or any domain library. The framework's `CLAUDE.md` describes it as "a Go service framework library"; adding a domain dependency would invert the layering.
- The two existing defaults and the replace/append semantics are frozen.
- **The default list grows by exactly one entry, so the four default-count assertions move by exactly one each; no other assertion changes.** The three replacement-based assertions are unaffected and must stay as they are. Changing any other assertion's expectation is out of scope.
- Error handling in the new code follows `github.com/bborbe/errors` patterns and the repo's `docs/dod.md`.
- Tests use Ginkgo/Gomega, matching the existing options test file.
- Counterfeiter is the mock generator (`mocks/` dir); no mock is required by any Desired Behavior or Acceptance Criterion — agent decides at impl time if one proves useful.

## Failure Modes

| Trigger | Expected behavior | Recovery | Detection | Reversibility | Concurrency |
|---------|-------------------|----------|-----------|---------------|-------------|
| Marked transient error reaches the reporter | Excluded — no event, no quota draw | None needed | The Sentry wrapper's `capture error … is excluded => skip` line at glog V(4) | Reversible — remove the predicate | Stateless predicate; no shared state |
| Unmarked error reaches the reporter | Reported exactly as today | — | Existing Sentry flow | N/A | Unchanged |
| Marker method panics | The exclusion check panics, and the error is NOT captured — worse than a missed suppression | The implementing library must make the method total and side-effect-free; treat a panic as a defect in that library | Panic in the reporter path | Reversible — fix the implementer | The method must be total and side-effect-free |
| A library marks a class that is not transient | That class stops being reported — a real defect is silenced | Remove the cause from the implementing library's classification | Missing Sentry events for that class | Reversible | N/A |
| A prolonged broker outage is marked transient | No Sentry event for the outage | The replacement signal must carry it — the owning task's forced-loss readings must show the `KafkaConsumerLag` alert firing, or the suppression is unsafe | Owning task's dev/prod forced-loss readings | Reversible | N/A |
| Caller replaces the exclusion list wholesale | Defaults including the new predicate are dropped; the caller owns the behaviour | Caller appends instead | Caller's own tests | Reversible | N/A |
| Chain-walk target declared as a non-pointer | Predicate never matches; nothing is excluded | Fix the target type | Unit test asserting a marked error IS excluded | Reversible | N/A |

## Suggested Decomposition

Single-layer, single-behaviour change in one file, plus the test-expectation updates it forces — a single prompt covers every Desired Behavior and Acceptance Criterion.

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | Add the structural transient predicate to the default exclusion list; move the four default-count assertions and their comments; add marked / unmarked / wrapped-chain unit tests | 1-7 | 1-11 | — |

## Why a spec rather than a prompt

The change is small, but it alters an **observable reporting contract** — which errors reach Sentry — for every service built on this framework, and its *why* is genuinely cross-repo: the structural-marker decision (rather than a text match, a shared package, or a domain import) is only explicable alongside the sibling Kafka spec. `choosing-a-flow.md` lists "a behavior change that affects observable contract … downstream consumers" and "a multi-prompt change where the cross-prompt rationale would be lost in any single prompt" as spec cases. The rationale is the durable artifact here; a future reader asking "why does the framework match an interface instead of a sentinel?" will grep `specs/completed/`.

## Do-Nothing Option

Doing nothing leaves the framework unable to recognise an expected transient class, so every consumer that wants to suppress one must pass its own predicate — editing 158+ `main` functions instead of one default list, and doing it again for the next library with a transient class. The cost of acting is one predicate plus four test-expectation updates, in a mechanism already deployed to every pod and already checked before an event is built. Doing nothing is acceptable only if the org prefers per-service wiring to a framework default — which the 2026-10-17 scope decision (fix the shared library) already rejected.
