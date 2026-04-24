# Feature Specification: Workflow And Lifecycle Hardening

**Feature Branch**: `004-workflow-and-lifecycle`
**Created**: 2026-04-24
**Status**: Draft
**Input**: Close sprint 003 with a handoff that captures the workflow and lifecycle gaps discovered during live validation, and carry those into a focused follow-up spec.

## Summary

Spec 003 got the Go runtime into a usable state and cleared the main startup blockers for real projects. During live validation, the next layer of work became clear: Stacklane now needs a tighter operator workflow around bootstrap, lifecycle recovery, and real-project verification.

This follow-up spec is intentionally narrower than the rewrite itself. It is not another architecture rewrite. It is the operator workflow and lifecycle hardening pass that turns the current runtime into a more repeatable daily tool.

## Problem Statement

The current runtime now starts projects, provisions per-project databases correctly, routes multiple attached sites through the shared gateway, and distinguishes stack-wide config from app-owned project config. However, the operator workflow still has gaps:

- project bootstrap remains partly manual unless `STACKLANE_POST_UP_COMMAND` is configured
- bootstrap behavior is not yet a fully documented, design-locked contract
- real-daemon lifecycle validation still lives partly in ad hoc manual checks rather than a formal verification path
- some app-facing follow-up work, such as seeding and post-migration setup, is still outside the current lifecycle contract
- app failures discovered during Stacklane validation need a clear boundary so Stacklane infrastructure issues and project code issues are not conflated

## Goals

- Define the intended operator workflow after `stacklane up`, especially for bootstrap-sensitive applications.
- Turn the new post-up bootstrap hook into a documented lifecycle contract with explicit failure behavior.
- Define how Stacklane should support common app bootstrap actions such as migrations, seeds, and setup commands without baking framework-specific assumptions into core runtime code.
- Formalize the live validation workflow for multi-project attach/up/status/down checks against real applications.
- Capture the remaining lifecycle and workflow gaps as scoped follow-up work instead of continuing to extend spec 003 informally.

## Non-Goals

- Re-open the Go rewrite architecture decision.
- Re-introduce legacy `20i-*` compatibility behavior.
- Fix application-specific schema or seeding bugs inside sibling projects unless explicitly requested as separate work.
- Introduce a new UI or TUI workflow.
- Solve release/distribution pipeline work unless it directly blocks the workflow/lifecycle contract.

## Candidate User Stories

### User Story 1 - Bootstrap A Real App Predictably

An operator wants `stacklane up` to bring a real application to a usable state without requiring undocumented manual follow-up commands.

Success means:

- the operator can declare an app bootstrap command deliberately
- Stacklane runs it in a predictable place and phase
- failures are surfaced as lifecycle errors, not silent app breakage
- the contract is documented clearly enough to reuse across projects

### User Story 2 - Separate Stacklane Failures From App Failures

An operator validating a project wants to know whether a failure belongs to Stacklane infrastructure or the application itself.

Success means:

- Stacklane validates its own readiness independently of app routes
- bootstrap hook failures are reported as hook failures, not gateway or health noise
- known app-level failures can be recorded without muddying Stacklane runtime status

### User Story 3 - Validate Multi-Project Workflow Against Real Projects

An operator running several local sites wants attach/up/status/down behavior to be verified against realistic project shapes rather than only mocked tests.

Success means:

- the workflow is validated against at least one representative app and one multi-site scenario
- DNS, gateway, runtime env, and DB provisioning checks are part of the workflow checklist
- manual validation steps that remain are captured explicitly

## Open Questions

- Should `STACKLANE_POST_UP_COMMAND` stay as a single hook, or should Stacklane define multiple phases such as post-up and post-attach?
- Which config scopes should be allowed to define bootstrap behavior: only `.stacklane-local`, or also stack-wide `.stackenv` and shell env?
- Should bootstrap failure always roll back the project runtime, or should Stacklane support a degraded but inspectable state?
- Should Stacklane add first-class guidance for common Laravel patterns such as `migrate`, `db:seed`, or `composer run setup`, while keeping the runtime framework-agnostic?
- What minimum real-daemon validation should be required before lifecycle changes are considered complete?

## Initial Backlog Candidates

- Document the bootstrap hook contract, precedence, rollback semantics, and examples.
- Add real-daemon lifecycle validation coverage for representative projects.
- Decide whether seed/setup commands belong in the same hook or a separate documented workflow.
- Add better operator diagnostics around post-up hook execution and output.
- Capture a boundary policy for app-level defects discovered during Stacklane validation.
- Reconcile spec 003 deferred lifecycle validation tasks with the narrower workflow/lifecycle scope here.

## Success Criteria

- The operator workflow after `stacklane up` is explicit and documented.
- A project can declare bootstrap behavior without ambiguity about config source or failure handling.
- At least one representative real-app workflow is reproducible without ad hoc manual steps.
- Remaining app-specific issues are clearly separated from Stacklane runtime issues in docs and handoff artifacts.
