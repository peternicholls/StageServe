# Cutover Runbook Skeleton

## Purpose

This file is the Gate A runbook skeleton for spec 006.

It defines the owner matrix, release-day sequence, abort criteria, and rollback checklist before implementation begins.

## Role Matrix

| Area | Owner role | Verification role | Notes |
|---|---|---|---|
| Contract acceptance | maintainer / spec owner | reviewer | Final naming, shim decision, and closeout rule |
| Inventory accuracy | implementation owner | reviewer | `inventory.md` must classify all active hits |
| CLI and installer cutover | implementation owner | runtime verifier | Includes help text, install path, and binary naming |
| Internal naming migration | implementation owner | code reviewer | Slice by owning subsystem |
| Docs and normative specs | docs owner | reviewer | Maintained docs only; archive stays archived |
| CI and release rehearsal | release owner | verifier | Cache rotation, assets, checksums, installer retrieval |
| Clean/dirty-machine validation | runtime verifier | reviewer | Must remain separate |
| Zero-active-reference sweep | implementation owner | reviewer | Final go/no-go gate before closeout |
| Rollback drill | release owner | reviewer | Must be rehearsed before public release |

## Release-Day Sequence

1. Confirm Gate A acceptance:
   - contract approved
   - runbook approved
   - inventory approved
2. Confirm Gate B evidence:
   - active hits classified
   - namespace risks documented
3. Confirm Gate C and D evidence:
   - local install rehearsal green
   - help/version rehearsal green
   - docs/spec rehearsal green
4. Confirm internal rename slices planned and validated.
5. Confirm Gate E and F evidence:
   - CI/release rehearsal green
   - rollback rehearsal green
6. Apply the final cutover packets.
7. Run verification matrix:
   - clean-machine validation
   - dirty-machine validation
   - focused tests
   - docs copy-paste audit
   - zero-active-reference sweep
8. Publish release notes and migration guidance.
9. Hold the post-release verification window.

## Abort Criteria

Abort or pause the cutover if any of the following is true:

- the final replacement names for active internal surfaces are still undecided
- local install rehearsal does not produce a reliable `stage` command path
- a shim is required for transition but cannot preserve stdout purity and exit-code parity
- CI or release rehearsal depends on stale cache state or an old binary path
- any maintained doc or normative spec still teaches `stacklane` as current behavior
- the final zero-active-reference sweep finds active code or maintained doc hits that are not explicitly historical or archival
- rollback depends on undocumented manual recovery

## Rollback Checklist

1. Restore prior installer target and published asset naming.
2. Restore the prior canonical command in release guidance if cutover has already been announced.
3. Revert the last rename slice that introduced the failure.
4. Re-run the focused validation for that slice.
5. Re-open the inventory and disposition table if the failure reveals a missing dependency.

## Evidence Checklist

- accepted rename contract
- accepted inventory table
- local install rehearsal transcript
- help/version rehearsal notes
- completion and shell-cache rehearsal notes
- CI/release rehearsal logs
- rollback drill notes
- clean-machine validation notes
- dirty-machine validation notes
- focused test outputs
- docs copy-paste audit log
- zero-active-reference search log

## Gate A Decisions Still Open

1. Final replacement names for `.env.stacklane`, `.stacklane-state`, and `STACKLANE_*`
2. Whether a temporary `stacklane` shim exists at all
3. Final replacement module path and any repository-rename dependency it creates
4. Final names for gateway headers, route sentinels, and health endpoints that currently carry `stacklane`
5. Whether `stln-*` remains the runtime prefix after review