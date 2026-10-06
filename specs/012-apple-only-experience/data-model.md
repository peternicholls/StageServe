# Data model and state transitions

Revised 2026-09-13. Target types mapped onto existing packages; no passing runtime evidence implied.
Detailed rules: [project/runtime](contracts/project-runtime-contract.md) and [release qualification](contracts/release-qualification.md).

| Entity | Owner | Durable fields and invariant |
|---|---|---|
| Effective settings | core/config | Value plus explicit origin, selected project UUID, app DB fallback below explicit StageServe values; secrets redacted |
| Identity / retained-data ledger | core/state, identities/<uuid>.json | Installation UUID, immutable project UUID, canonical path, display slug, registration status, exact resources, engine digest/last writer and schemas; survives detach |
| Registry projection | core/state | Registered identities, route and display status; deleting registration never deletes retained ledger |
| Apply journal | core/lifecycle + core/state | Operation ID, base revision, desired/applied hashes, prior compatible config, pending resources, phases and leftovers; durable intent before mutation |
| Runtime observation | core/runtime | Exact IDs/labels, per-interface address, health, timestamp; unknown is distinct from stopped |
| Next action plan | core/guidance | Situation, explicit selected UUID, verdict/facts, visible defaults, base revision and enabled actions; no renderer-owned truth |
| Route | infra/gateway | Project UUID, hostname, scheme, redirect, leaf certificate, observed upstream interface; TLS policy per route |
| Listener/host-change record | platform + core/state | Installation ports, exact resolver/cert ownership, original bytes and expected current hash; never undo external edits |
| Release compatibility | installer + retained ledger | Binary/assets/installer digests, image platform digests, state read/write range, volume engine provenance, qualified rollback pairs and evidence digests |

## State machine

Unconfigured -> preview -> configured/stopped -> starting -> infrastructure-running.
Application health is a separate optional observation: not_checked/healthy/unhealthy.
Infrastructure-running -> stopping -> stopped; stop preserves registration and ledger.
Detach stops selected services, clears route/registration/generated env, retains ledger,
settings and volume. Reattach checks ledger and exact ownership before reuse.
Delete-data is explicitly confirmed against immutable identity; volume tombstones remain.

Settings: preview(base revision) -> locked revalidation -> journal -> prepare -> apply
-> probe -> gateway reload -> commit. Failed apply restores prior compatible config;
incomplete rollback is degraded with named leftovers, never a false running state.
Stopped edits change desired settings only; next start reconciles applied fingerprint.
No readiness probe rewrites app files or application database rows.

## Migration and concurrency

Missing legacy runtime identity remains legacy; unknown/newer schema is rejected before
mutation. Ordinary N/N+1 updates require backwards-readable state and pinned DB engine;
unsafe rollback is refused, not implemented by discarding new writes. Copies get new
identity/data; moved or missing-ledger paths require explicit recovery.

Lock project UUID then shared route for short commit/reload; broad operations acquire
UUID locks in sorted order, never acquire a project lock while holding shared route.
No lock spans user input. Cleanup uses a fresh bounded context. Atomically replace
individual records; journal recovery handles multi-file boundaries and disk failures.
