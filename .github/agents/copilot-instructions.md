# StageServe development context

Updated 2026-09-14 through Spec Kit plan refresh and authority reconciliation.
Constitution 3.1.0 and specs/012-apple-only-experience are current authority.

## Active technology and boundaries

Go 1.26/toolchain 1.26.2; existing Cobra, Bubble Tea, Bubbles, Huh and Lip Gloss.
Apple container subprocess adapter only; no Docker/Compose runtime fallback.
Atomic JSON registry, independent retained-identity ledger and apply journal are
the target persistence model. Linux OCI images run on tested Apple silicon Macs.

Source modules: cmd/stage/commands, core/{config,state,guidance,lifecycle,runtime},
infra/{applecontainer,gateway}, observability and platform. Existing legacy adapters
are retirement work, not a second supported runtime. No speculative src/tests tree.

## Contracts and checks

Read contracts/project-runtime-contract.md and contracts/release-qualification.md
under spec 012 before changing settings, identity, networking or updates. GUX in
ux-validation.md gates full dashboard implementation. Distinguish pending target
behaviour from observed code; never claim live compatibility from mocks.

Run python3 scripts/verify-planning.py after planning/archive changes. Runtime
changes require focused tests then full short tests, vet/build/lint and applicable
live gates. SPECIFY_FEATURE=012-apple-only-experience selects the feature scripts.
archive/** and previous-version-archive/** are immutable historical reference.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
