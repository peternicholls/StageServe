# Application, runtime and interaction contract

Target for spec 012. This is an in-process/CLI contract, not an HTTP API. Existing names should be reused where semantics match; additive changes need tests and documented compatibility.

## Entry points

| Entry | Input | Output and side effects |
|---|---|---|
| Bare `stage` | Current folder, terminal mode, shared config | Detect/read context; interactive guided shell or noninteractive plan; no automatic mutation on entry |
| `stage setup` / `stage doctor` | Resolved scope | Read-only readiness report; assistance is a separately selected action |
| `stage init` | Previewed settings / explicit force flags | TTY styled/text preview; explicit --non-interactive or non-TTY invocation uses validated defaults; existing files require --force outside confirmed editor |
| `stage up` | Current/explicit project, resolved defaults | Scoped start with health/route validation and persisted outcome |
| `stage status` | Current, `--project`, or `--all` | Read-only desired/observed state; unavailable runtime reported honestly |
| `stage logs [service]` | Explicit project/service; positional and flag conflict rejected | Cancellable bounded stream; useful missing-service error |
| `stage down` | Explicit project; all-project mode visibly broad | Stop selected resources; retain data unless separate explicit destructive request |
| `stage attach` | Explicit configured project | Re-register and restore route; reuse healthy owned services, otherwise start; never claim healthy before checks |
| `stage detach` | Explicit current project | Stop project, clear registration/route/generated env; preserve retained ledger/settings/source/volumes |
| Guided restart/browser/settings/remove | Typed action and exact scope | Existing domain actions where available; revalidation and target-specific confirmation as appropriate |

All direct help paths bypass the TUI. `--notui`, `--cli`, `STAGESERVE_NO_TUI=1` choose text. JSON request wins over human styling/assistance, with diagnostics on stderr and one structured result on stdout; streaming logs need an explicitly documented stream format if JSON is offered. Non-TTY must never block for input. `NO_COLOR` removes styling without removing meaning.

Exit semantics: success is 0; ordinary operational/validation failures are 1. Errors implementing the existing `ExitCoder` contract preserve their explicit command code and silent-output convention (including setup/doctor readiness codes). Cancellation is 130 when the returned error wraps `context.Canceled` or execution fails after the root context is canceled by SIGINT/SIGTERM, including subprocess errors reported as a killed process. Cancellation takes precedence over an operational error code; diagnostics go to stderr. A command that completes successfully remains 0 even if cancellation arrives at completion. Deadline expiry alone is a failure, not cancellation. These numeric conventions are frozen by the binary entrypoint tests in T004. JSON carries stable check/error IDs, schema version, project scope, observed state, next action and redacted details. Unknown/additive fields must not change existing field meaning.

## Runtime adapter

Input is a validated Apple manifest and resolved project identity, not arbitrary shell text. Invoke CLI with separate argv tokens and context deadlines, never shell interpolation. Pin supported version behaviour and build parsers from captured real CLI fixtures. Pull/build/run/mount/inspect/stop/delete/logs/exec/restart semantics are verified against the supported release.

Start returns observed service addresses/IDs. Orchestrator generates application/gateway endpoints from verified addresses or proven names. Readiness includes runtime service, required features and selected host topology. Unsupported capabilities fail before creating resources. No silent Docker fallback.

Every destructive call must prove exact StageServe ownership against current observation; mismatched labels/unknown resources are not touched. Use manifest volume names, not reconstructed guesses. Inspect health independently of process-local caches. Health checks have a total deadline and named failed-service result. Exec preserves output/exit status and cancels cleanly.

## Operation and presentation

One operation service owns ordering, locking, state commits and cleanup. CLI and TUI call it; neither a render function nor mockup defines truth. Emit structured phase/progress/result events; observers may detach without corrupting state. After every completed/failed action, collect fresh context and replan. TTY --cli/--notui uses the complete interactive line-oriented flow defined in ux-validation.md; redirected output remains noninteractive.

Running Enter = view logs. Ready/stopped Enter = run. Host-change and destructive sheets show exact target/effect and default to Cancel. Browser opening is explicit, uses the observed validated URL and is not part of the runtime health test. Display unavailable/unknown distinctly from stopped.

## Local networking and host changes

Default target is local HTTP `.test`; local TLS is opt-in and requires verified trust and displayed port. Project settings with other explicit suffixes remain preserved until validation. DNS provider remains behind the existing boundary; no unverified claim that Apple DNS replaces setup. Host DNS/trust previews identify exact files/domains/certificates and require operator authorization. List every published port and verify local exposure. Database/debug endpoints are not public defaults.

## Release and migration

Reject incompatible host/runtime and state-schema versions before changes. Installation verifies artifact checksums and binary/asset compatibility. Update stages a digest-qualified pair then switches its active pointer atomically, only for qualified state/image/database compatibility; rollback must preserve new writes or refuse with explicit recovery. Never overwrite project config or DB during update. Legacy Docker records are migration inputs only, not runtime resources. Migration instructions require backup, export/import verification and restoration steps without global prune/reset.

## Exact lifecycle and automation decisions

- `down`: stop only selected project, retain its registration/settings/volumes and mark stopped; retain the hostname reservation and serve an explicit 503 stopped response; never proxy a stopped service. `down --all` applies this to recorded projects without deleting records/data.
- `detach`: stop selected services, unregister and clear its route/generated env; preserve the independent identity/retained-data ledger, source, settings and database volumes. Guided "Remove this project from StageServe" maps here and previews both stopping and unregistering. Direct explicit `detach` is authorization for that documented non-data-destructive scope; no new hidden data deletion.
- `attach`: add back and route selected project; if owned services are healthy reuse them, otherwise start the required services. Restore routing/registration only after health/route validation. Reattachment finds retained volume by exact stable ownership.
- Data deletion remains an explicit `down --volumes` request, not ordinary stop/detach. `--confirm-project <slug>` is required for noninteractive use, matching the selected project exactly. Interactive use shows the validated project UUID and retained live volume names/IDs, with Cancel default; typing the exact slug authorizes the preview. `--all --volumes` is rejected so each destructive scope is explicit. Dry-run previews do not require consent or run the runtime. CLI input consent is implemented; complete ownership, lock revalidation and live destructive acceptance remain T040/T020 obligations.
- Interactive `init` previews values before writes. Direct `init --non-interactive` (or non-TTY/JSON invocation) is explicit creation consent with validated documented defaults; it reports the exact resulting values/path. `init --dry-run` previews without writing and works with JSON; `--force` is explicit overwrite consent. Existing files without force remain unchanged. No new preview token or mandatory confirmation round-trip is added for scripts. T004 locks this compatibility, including invalid web folder, cancel/refusal and no-write checks.
- Moved folders: v1 detects/refuses ambiguous relocation with a documented manual recovery procedure using preserved settings and backed-up data. A first-class relocation command is deferred. Do not auto-register a second identity or reuse an old project's volumes based on basename alone.

T004 freezes these action inputs/results and transition semantics before fixture-driven dashboard implementation. T013–T020 implement their ownership/locking guarantees before live UI acceptance.

## Detailed contracts and precedence within this package

[Project/runtime contract](project-runtime-contract.md) defines settings transactions,
durable ledger, application DB fallback/probes, connectivity and mixed TLS policy.
[Release qualification](release-qualification.md) defines candidate artifacts, version
compatibility, offline/resource and host lifecycle obligations. [UX protocol](../ux-validation.md)
defines the early decision gate and accessible interactive text semantics.
These elaborate this command contract; conflicts must be resolved together, not by
letting different interfaces choose different behaviour.
