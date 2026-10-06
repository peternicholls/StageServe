# Spec Kit completion audit — planning repair

Date: 2026-09-14. Inputs: constitution 3.1.0, active spec/plan/tasks, detailed
contracts, UX protocol, research register, roadmap, instruction scopes and archive.
This supersedes the 2026-09-11 four-finding approval, which was too broad. The later
review identified eleven substantive planning gaps; see [resolution record](review-resolution.md).

## Scope and result

All eleven reviewed planning defects now have concrete definitions, sources where
needed, owned tasks and explicit acceptance. Authority/archive repairs are applied.
The package is ready for the ordered feasibility/research and implementation work;
it is not a claim that the entire architecture is empirically qualified or that
StageServe v1 is complete. G0, GUX and G1–G4 remain open.

No production runtime code or release workflow was changed by this planning repair.
Their known defects remain represented by unchecked implementation tasks. Existing
working-tree implementation and historical source bytes were preserved.

## Requirement-by-requirement check

[review-resolution.md](review-resolution.md) maps findings 1–11 to concrete artifacts,
FRs, tasks and cases. Author-side review checked the following transitions against
actual current code boundaries and the revised contracts:

- Editing running/stopped settings: desired/applied revisions, field effects,
  preparation before downtime, stale preview, failure rollback and DB preservation.
- Detach/fresh-process reattach: independent retained ledger and exact labels/IDs,
  rather than the deleted registration or a reconstructed name prefix.
- Existing application: explicit source precedence and no app writes; optional
  app health cannot tear down healthy infrastructure. Fixtures own sentinel writes.
- Multi-project networking/TLS: negative private paths and explicit local host trust
  boundary; per-route policy cannot follow whichever project ran last.
- N/N+1 update/rollback: state readability and pinned engine, latest writes retained,
  refusal of incompatible downgrade, candidate digest qualification before promotion.
- Early UX and accessible text: fixture study after T004, before T024; actual T043
  text implementation is not a circular prerequisite of the prototype study.

These checks establish contract completeness for the reviewed defects; runtime
behaviour is still demonstrated only by the later live cases.

## Automated planning evidence

Executed 2026-09-14 with Python standard library, no installed dependencies:

- `python3 scripts/verify-planning.py`: passed 45 unique tasks, 23 FR coverage rows,
  11 SC mappings, acyclic dependencies, 19 linked documents, 88 original content
  hashes, inventory partition and Git-ignore eligibility, and scoped authority.
- Fresh Git-shaped temporary copy with no `.DS_Store` metadata: passed.
- Negative controls in isolated copies: changed archived content, dependency cycle,
  broken link, invalid FR task mapping and reintroduced Compose authority each
  produced the expected failure; restoring the copy passed again.
- Python syntax parse passed. The first temporary-copy trial correctly failed for
  an omitted constitution file; after including the required .specify tree, the
  clean-copy and all negative controls passed. No fixture edits touched the repo.
- Spec Kit `check-prerequisites.sh --json --require-tasks --include-tasks` passed
  with `SPECIFY_FEATURE=012-apple-only-experience`.
- `git diff --check` passed after the repair.

The verifier checks structure and provenance, not semantic truth or live gates. The
manual transition review above and original issue mapping provide the semantic
check. It cannot certify DNS, data recovery, usability or release readiness.

## Research and verification limits

Research now uses Apple container 1.4.1 versioned primary documentation and a
candidate qualification matrix. R-01–R-10 specify inputs, outputs, owners and
failure decisions. No actual runtime, participant, VoiceOver, benchmark, signing
or release results are fabricated. The previously recorded Go test/vet/build
passes were from 2026-09-11; this documentation repair did not rerun them or use
them to certify the new targets. Staticcheck was previously absent.

Independent repair/re-review agents could not complete because of an agent usage
limit. The repairs and final review were completed directly by the author. No new
independent approval is claimed. The initial independent audits remain the source
of the eleven findings, not validation of the current revision.

## Handoff

Start with T001 baseline refresh (T038 is already complete), then candidate/runtime
and ledger work. Preserve numerical task IDs and follow dependencies, not line
order. Full dashboard work waits for GUX; actual release promotion waits for the
same-digest qualification cases. Report a failed experiment as failed and replan
within the stated product boundaries; do not hide it with a broader success claim.
