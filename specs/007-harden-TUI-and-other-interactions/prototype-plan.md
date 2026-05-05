# Prototype Plan: Guided TUI Flow Mock

## Purpose

Build a tracked, non-production terminal prototype to test spec 007 flows, copy, and assumptions before any production CLI changes are made.

This prototype exists to answer:

- Does bare `stage` as a guided entrypoint feel natural?
- Are the easy-mode labels clear enough for a front-end developer or hobbyist?
- Do first-run, run, stop, drift, and recovery flows read coherently?
- Do command equivalents stay useful without becoming the primary mental model?

## Boundaries

- Keep the prototype outside `cmd/stage`, `core/`, and shipped command paths.
- Do not import or reshape production lifecycle or config code.
- Do not execute Docker, write `.env.stageserve`, or mutate state.
- Reuse the existing Go module and existing Charm dependencies already present in `go.mod`.

## Prototype Shape

- Location: `specs/007-harden-TUI-and-other-interactions/prototype/`
- Form: runnable terminal mock using Bubble Tea
- Scope: core guided flows plus text fallback
- Inputs: canned scenario fixtures, not live system inspection
- Outputs: interactive flow behavior, text fallback output, and recorded findings

## Scenarios

The prototype must support these scenarios:

- `machine_not_ready`
- `project_missing_config`
- `project_ready_to_run`
- `project_running`
- `project_down`
- `drift_detected`
- `not_project`
- `unknown_error`

## Interaction Rules

- TUI is the default when the prototype is run in a TTY.
- `--notui` and `--cli` force text fallback.
- `--scenario` selects a starting scenario.
- `--list-scenarios` prints the supported fixtures.
- `show commands` always reveals direct command equivalents.
- `edit project settings` shows the `.env.stageserve` path and sample content, but never opens an editor.
- Multi-project switching is not implemented; the prototype explicitly states that v1 stays scoped to the current directory.

## Flow Goals

- `machine_not_ready` should guide to `project_missing_config`
- `project_missing_config` should preview config and guide to `project_ready_to_run`
- `project_ready_to_run` should guide to `project_running`
- `project_running` should allow status, logs, stop, doctor, commands, and advanced
- `project_down` should allow rerun, status, remove-from-StageServe, and doctor
- `drift_detected` and `unknown_error` should prioritise recovery clarity

## Verification Loop

Use terminal verification first:

1. Run `go run ./specs/007-harden-TUI-and-other-interactions/prototype --list-scenarios`
2. Run text fallback for each core scenario with `--notui`
3. Run the TUI in a TTY and manually exercise setup, config preview, run, stop, logs, commands, and recovery paths
4. Fix prototype issues immediately
5. Re-run the same terminal checks
6. Use focused automated tests only as supporting checks

## Acceptance

- The prototype runs separately from the main CLI.
- Easy-mode labels stay in plain language.
- Direct command equivalents are visible but secondary.
- Config preview and confirmation feel credible without mutating files.
- Drift, unknown-error, and stop flows have a clear next step.
- The prototype README contains the exact terminal test loop used during review.
