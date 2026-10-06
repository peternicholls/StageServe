# Eleven-finding repair record

Completed authoring 2026-09-13; completion audit 2026-09-14. Scope: the eleven
planning/research/authority defects in the review, not implementation of v1.
The original review superseded the earlier broad planning approval. This record
requires a concrete contract, owned work and verification for each missing item.
Runtime/user/release cases remain open until executed; defining a case is not a pass.

| Finding | Concrete repair | Owned implementation/research and acceptance | Planning disposition |
|---|---|---|---|
| 1. Settings application undefined | [Project contract](contracts/project-runtime-contract.md): keyed desired/applied fingerprints, field effect table, stale revision rejection, stopped/running transaction, journal and rollback preserving data | T041/T025; CFG-01–04; FR-016; SC-008 | Definition and task gap resolved |
| 2. Identity lost on detach | [Retained ledger](contracts/project-runtime-contract.md): immutable installation/project UUID, ledger independent of registration, exact labels/IDs, copy/move/missing-state/crash rules | T040/T013/T019; ID-01–04; FR-017; SC-008 | Durable location and lifecycle resolved |
| 3. App health/config ambiguous | [Application contract](contracts/project-runtime-contract.md): documented three-key .env fallback with explicit provenance; StageServe-owned PHP and SELECT 1 probes; optional app-health separate; no app writes | T004/T009/T011; APP-01–04; FR-018; SC-002 | Scope/probe/source-write decisions resolved |
| 4. Isolation not measurable | [Connectivity matrix](contracts/project-runtime-contract.md): allowed/denied paths, private application/ingress candidate, IPv6/host-alias negative tests, explicit threat boundary and candidate rejection | T003/T014/T016; NET-01–04; R-02/R-03; FR-007; SC-005 | Policy/experiment defined; feasibility still G0 |
| 5. Mixed TLS unspecified | [Per-route policy](contracts/project-runtime-contract.md): shared listener ownership, exact scheme/SNI/redirect/cert rules, renewal/removal, retain other routes on failure | T014/T016/T020; TLS-01–04; FR-013 | Mixed-policy contract and tests resolved |
| 6. Rollback excludes persistent format | [Compatibility contract](contracts/release-qualification.md): qualified N/N+1 state read/write, image/engine provenance, existing-volume engine pin, new-volume migration and post-write rollback/refusal | T031/T032/T033; UPD-01–04; FR-019; SC-009 | Compatibility/data-loss policy resolved |
| 7. Actual artifacts not qualified | [Candidate pipeline](contracts/release-qualification.md): build once, real production download/install outside checkout, digest inventory, protected evidence gate and same-digest promotion; signing status honest | T032/T045/T036; REL-01–04; FR-019; SC-009 | Release pipeline specification/task gap resolved; current workflow changes remain implementation |
| 8. Conflicting authority | [Active docs instruction](../../.github/instructions/docs-contract.instructions.md), [runtime instruction](../../.github/instructions/runtime-go.instructions.md), [archive rule](../../.github/instructions/archive.instructions.md) now follow constitution 3.1.0/spec 012, exclude archive from active scope and reject Compose authority | T038 completed; verifier checks authority and scope; T035 checks future release docs parity | Applied now, not deferred |
| 9. Untested operational promises | [Release/capacity policy](contracts/release-qualification.md) and [responsiveness](ux-validation.md): exact offline cache/recreate boundary, minimum host targets, CPU/RAM/disk/log/cancel bounds, sleep/reboot and supervised uninstall | T042/T026/T044; OFF/CAP/HOST families; FR-020–021; SC-010 | Requirements/owners/measurements defined; measured results remain G1/G3/G4 |
| 10. Late UX and incomplete accessibility | [Early study](ux-validation.md): five actual participants, desktop discovery, thresholds and alternative-entry decision before T024; complete interactive text/VoiceOver journey and final trials | T039/T043/T028/T029; UX/A11Y families; FR-022–023; SC-011 | Research timing and accessible-mode contract resolved; actual participation remains GUX/G3 |
| 11. Archive not reproducible in Git | [Content inventory](../../archive/2026-09-11-pre-apple-only/manifest.json): 88 content entries; original 90 snapshot and two optional Finder metadata entries preserved separately; no original hashes changed | T038 completed; verifier checks exact partition, all content hashes and Git eligibility; T037 repeats | Applied now and reproducibly verifiable |

## Integration and review scope

Spec now has 23 functional requirements and 11 measurable outcomes. Tasks T001–T045
have explicit dependencies and implementation/verification mappings. Stable IDs are
preserved; dependency edges, not numeric order, determine execution. T038 is the
only completed task because runtime implementation and observed studies have not
been performed here. G0, GUX and G1–G4 are not marked complete.

Constitution 3.1.0, product direction, plan, data model, command contract, roadmap,
research register and acceptance runbook incorporate these decisions. Earlier
6–12.5-week allowance is replaced by 9.5–17.5 engineering weeks, with actual
re-estimates after G0/GUX/G1. Ten research entries have inputs, outputs, owners and
failure decisions; primary Apple sources are versioned to candidate 1.4.1.

## What this repair does not claim

No production runtime feature, published release, benchmark, participant study or
rendered mockup acceptance is claimed. These were implementation gates in the
original roadmap and remain explicit gates. The repaired planning gaps are closed
by precise decisions and assigned proof, not by fabricating that proof. Source and
runtime work already in the checkout are preserved. No dependency was installed.

The attempted independent follow-up was unavailable due to an agent usage limit;
final author-side review and structural negative controls are recorded in analysis.md.
The earlier independent audit supplied the findings, not approval of these repairs.
