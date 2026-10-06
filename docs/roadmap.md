# StageServe Apple-only roadmap

Revised 2026-09-13 • Target roadmap • [Product decision](product-direction.md) • [Executable task order](../specs/012-apple-only-experience/tasks.md)

## Baseline and planning assumptions

Apple-only configuration and production CLI/TUI wiring exist in the dirty working tree. The core guidance model, terminal design system and adapters are reusable. The product is not yet verified on Apple `container`: it is absent from PATH on the reviewed macOS 27.0/arm64 host. Two focused suites initially failed on old Docker/hostname expectations; concurrent work resolved them, and the refreshed full short suite, vet and build pass. Staticcheck is unavailable, so lint remains open. Existing completion checkboxes do not count as new release evidence.

Estimate basis: one experienced Go developer-equivalent, focused work, access to a suitable test Mac and representative PHP projects. These are engineering-week ranges, not dates or commitments. M0/M1 must re-estimate the remaining work using measured runtime behaviour. Discovery can expand scope; do not hide that by reducing acceptance.

## Ordered milestones

| Milestone | Outcome and principal work | Estimate | Prerequisite | Exit gate |
|---|---|---|---|---|
| M0 — Establish identity and feasibility | Qualified CLI fixtures, durable identity/ledger, configuration semantics, exact connectivity/TLS experiments and early study preparation | 1.5–2.5 weeks | Planning repairs; candidate test host | G0: baseline green; CLI/schema/topology evidence; ledger refusal/recovery tested |
| GUX — Validate operator entry | Five-person discovery, project switching and accessible text study; fixture corrections | Included across M0/M3 | T004; recruited participants | GUX: thresholds in UX protocol pass before T024; otherwise revise interface decision |
| M1 — One safe Apple site | Correct endpoints/probes, transactional settings, cache reuse and minimum-host resource qualification | 2–4 weeks | G0 | G1: fixture PHP/DB/browser + real app health distinction; CFG/APP/OFF/CAP cases; 20 cycles |
| M2 — Safe multi-project operations | Scoped ownership/recovery, negative connectivity, mixed TLS, logs and exact deletion/reattach | 2–4 weeks | G1 | G2: ID/NET/TLS and failure matrix passes; no unrelated mutation |
| M3 — Complete both operator modes | Existing TUI plus accessible interactive text, settings apply, progress/log bounds, final trials | 2–4 weeks | GUX; fixture work after G0; live acceptance after G2 | G3: A11Y/latency and actual terminal/first-use cases pass |
| M4 — Qualify and distribute v1 | Actual candidate binary/bundle, compatibility after writes, host lifecycle/uninstall and gated promotion | 2–3 weeks | G2/G3; compatibility design starts at G0 | G4: REL/UPD/HOST and installed-candidate matrix; exact digests promoted only after evidence |

Revised allowance: **9.5–17.5 engineering weeks**, approximately **12–23 calendar
weeks** for one full-time experienced Go developer with integration/review allowance.
The increase accounts for the missing settings/identity transaction work, accessible
text journey, measured offline/resource work and actual artifact qualification.
These are planning estimates, not measured throughput or delivery dates. No double
counting of GUX: its effort is included in M0/M3. Participant availability and access
to minimum-spec Apple hardware are explicit external scheduling dependencies.

Re-estimate after G0 topology/CLI evidence, after GUX, and after G1 resource/runtime
results. Record actual effort by task, remaining work and unresolved risk. A failed
network policy or terminal-entry study can require redesign; this is not absorbed
by silently weakening requirements. The previous 6–12.5-week allowance is superseded.

## Dependency and parallel work

Critical path: M0 -> M1 -> M2 -> final M3 acceptance -> M4; GUX independently blocks full dashboard implementation. Fixture preparation and study can proceed after T004; full TUI work waits for GUX and G0 alongside M1/M2; packaging design and documentation can also start early. Shared lifecycle/config/state files need one owner at a time. Parallel work changes elapsed time only when genuinely separate contributors are available; it does not reduce total effort.

M1 is a developer vertical slice, not a public alpha. M2 plus the essential M3 journey can support a limited pilot. V1 requires GUX and every G0–G4 gate. If time is constrained, defer debug tools and cosmetic enhancements through a spec amendment; never defer isolation, data safety, real routing evidence or CLI/TUI truth parity silently.

## Decisions at the risky gates

- **Runtime version:** use Apple container 1.4.1 as the initial candidate, capture its real output and explicitly tested macOS versions, then expand the support matrix only with evidence. macOS 26 is an upstream floor, not proof every newer release works.
- **Network:** execute the candidate private application/ingress topology, allowed/denied matrix and NET cases in the project-runtime contract using inspected per-interface addresses. Do not rely on bare `apache` name lookup or equate network creation with use. If isolated topology fails, halt G0/G1 and redesign within Apple-only scope; no Docker fallback.
- **DNS:** retain the existing provider boundary; compare existing dnsmasq path with Apple DNS on the pinned version. Prefer qualifying the existing provider unless an Apple-native path demonstrably removes host setup. Do not create a separate DNS product.
- **Legacy state/data:** inspect records without mutation; never relabel old Docker state as Apple. Retain originals, export/import DB with reconciliation, and offer manual migration steps. No prune commands or direct cross-runtime volume adoption.
- **Distribution:** retain current Go toolchain intent (`go.mod` Go 1.26, toolchain 1.26.2) until explicitly changed. Checksum-verified binary and asset pairs first; no new dependency or auto-updater service is assumed.

## Post-v1 options, outside the committed plan

1. A separate native GUI remains outside the assumed delivery scope. Early GUX can trigger a comparison/decision amendment before M3 if terminal entry fails; later demand may justify another feature. No GUI implementation date or toolkit is preselected.
2. Additional stack profiles: only after one-stack operational proof and concrete demand; add a separately tested profile, not arbitrary Compose compatibility.
3. Streamlined Apple-native DNS or embedded assets: adopt only if the measured setup/update burden justifies the change. Keep these out of the runtime critical path.

## How progress is recorded

Each gate owns a dated evidence entry with commit/worktree identity, host/CLI versions, command or manual scenario, expected/observed result and unresolved limitation. Tasks remain unchecked until the relevant evidence exists. A passing unit suite cannot close a live gate. A blocked runtime gate permits independent design/docs work, but blocks release claims. See [quickstart acceptance matrix](../specs/012-apple-only-experience/quickstart.md).

## Detailed research and qualification

[Project/runtime contract](../specs/012-apple-only-experience/contracts/project-runtime-contract.md)
settles field effects, durable identity, app configuration/health, private connectivity
and mixed TLS. [Release qualification](../specs/012-apple-only-experience/contracts/release-qualification.md)
settles version/DB rollback, actual candidate promotion, offline/cache, resource and
host lifecycle policy. [UX validation](../specs/012-apple-only-experience/ux-validation.md)
defines early participants, task scripts, decision thresholds, VoiceOver/text and
responsiveness. [Research ledger](../specs/012-apple-only-experience/research.md)
separates versioned documentation from still-open empirical experiments.
