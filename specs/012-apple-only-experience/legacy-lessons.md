# Legacy planning lessons and disposition

Reviewed 2026-09-11 before archive. All source files retain their original content and historical status. [Archive manifest](../../archive/2026-09-11-pre-apple-only/manifest.json) supplies original paths and checksums.

| Source | Retain | Retire / evidence caution |
|---|---|---|
| [002-project-rebrand](../../archive/2026-09-11-pre-apple-only/specs/002-project-rebrand/research.md) | One product identity, canonical stage and consistent help/config names. | No further rebrand milestone. Draft and checked ledger are not release proof. |
| [003-rewrite-language-choices](../../archive/2026-09-11-pre-apple-only/specs/003-rewrite-language-choices/tasks.md) | Go modular core, thin presentation, named failures, machine contracts and real-project gates. | Reject Docker/Intel assumptions; historical live lifecycle, health and installation tests remained deferred. |
| [004-workflow-and-lifecycle](../../archive/2026-09-11-pre-apple-only/specs/004-workflow-and-lifecycle/implementation-review.md) | Config ownership, readiness before bootstrap, rollback, cancellation, isolation and useful failures. | Review reports implementation/bookkeeping divergence; do not copy Compose resource naming as a product contract. |
| [005-installer-and-onboarding](../../archive/2026-09-11-pre-apple-only/specs/005-installer-and-onboarding/tasks.md) | Deterministic install path, checksum verification, typed readiness and shared projections. | Disconnected setup/init/up commands did not complete the guided journey; later 007 review corrects completion claims. |
| [006-project-and-command-renaming](../../archive/2026-09-11-pre-apple-only/specs/006-project-and-command-renaming/plan.md) | Inventory before cutover, PATH/install/release/rollback rehearsal and small delivery packets. | No more rename migration; never use Docker prune for migration. Tickets contain stale mechanical naming claims. |
| [007-harden-TUI-and-other-interactions](../../archive/2026-09-11-pre-apple-only/specs/007-harden-TUI-and-other-interactions/quickstart.md) | Continuous context-aware flow, safe defaults, previews, user language, shared core and terminal acceptance. | Quickstart around line 369 explicitly says live lifecycle was not rerun. Completion marks are not live proof; missing prototype directory is not restored. |
| [008-version-management-releases-and-update-workflows](../../archive/2026-09-11-pre-apple-only/specs/008-version-management-releases-and-update-workflows/spec.md) | Version source, reproducible artifacts, checksums, update/rollback and release rehearsal. | Placeholder only, not implemented release machinery. |
| [009-documentation-update-and-declutter](../../archive/2026-09-11-pre-apple-only/specs/009-documentation-update-and-declutter/spec.md) | User-first README, command reference and separate contributor detail. | Placeholder only; obsolete man stack wording; docs must ship with each milestone. |
| [010-extract-local-dns-project](../../archive/2026-09-11-pre-apple-only/specs/010-extract-local-dns-project/spec.md) | Typed provider boundary, readiness, previewed host changes, ownership and unsupported states. | Entire task ledger open; standalone DNS product is not a prerequisite. Assess Apple DNS on the pinned release first. |
| [011-guided-experience-and-runtime-hardening](../../archive/2026-09-11-pre-apple-only/specs/011-guided-experience-and-runtime-hardening/tasks.md) | Prerequisites, running-project defaults, service selection, logs viewport, progress/cancellation and confirmations. | T041–T050 remain open: docs/copy and final automated/manual acceptance. Replace Compose seams. |

## The recurring lesson

004 established lifecycle contracts; 005 delivered onboarding commands; 007 recovered the intended guided journey; 011 hardened daily operations. Useful components were delivered while the complete operator journey and final evidence lagged. The new roadmap therefore gates usable vertical slices rather than counting modules or old checked tasks.

Spec 007's `spec-planning-departures.md` and `current-implementation-review.md` are the clearest retrospectives. Spec 004's planning/implementation reviews show why task bookkeeping cannot be treated as runtime truth. Spec 003's deferred tests explain why a compiled rewrite was not itself distribution readiness.

## Concrete carry-forward

- Config/identity/isolation/error semantics -> FR-002–004, FR-007–009 and M0–M2.
- One continuous guided TUI and shared projections -> FR-005–006, FR-010–011 and M3.
- Real browser, DB and terminal evidence -> SC-001–005, G1–G3.
- Installer/update/rollback and doc parity -> FR-012, FR-015 and G4.
- DNS boundary and exact host-change previews -> FR-009, FR-013; no separate product.
- Canonical product identity and one backlog -> spec 012 and this archive; no new renaming effort.

## Disposition outside specs

Older `docs/plan.md`, recommended-roadmap documents, project/architecture analysis, embedded-runtime asset plan and superpowers prototype plans remain source material with superseded notices. They are not extra active roadmaps. Current `docs/design/` and `mockups/` remain design references; their Docker copy and unverified rendering are explicit M3 work. `.omx/plans/` contains historic hybrid analysis, not authority for the Apple-only scope.
