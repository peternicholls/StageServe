# Guided TUI Prototype

This is a tracked, non-production prototype for spec 007. It is intentionally separate from the shipped `stage` command and exists only to test flows, copy, and assumptions before production implementation.

## Run

```bash
go run ./specs/007-harden-TUI-and-other-interactions/prototype --list-scenarios
go run ./specs/007-harden-TUI-and-other-interactions/prototype
go run ./specs/007-harden-TUI-and-other-interactions/prototype --scenario project_running
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario machine_not_ready
go run ./specs/007-harden-TUI-and-other-interactions/prototype --cli --scenario drift_detected
```

## Controls

- `up` / `down`: move between actions
- `enter`: choose action
- `c`: show commands
- `esc`: go back
- `q`: quit

## Terminal Verification Loop

Run these commands while refining the prototype:

```bash
go run ./specs/007-harden-TUI-and-other-interactions/prototype --list-scenarios
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario machine_not_ready
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario project_missing_config
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario project_running
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario drift_detected
go test ./specs/007-harden-TUI-and-other-interactions/prototype
```

Manual TTY checks:

- Start at `machine_not_ready` and confirm the primary action is `Set up this computer`
- Move through config preview and confirmation without writing files
- Start at `project_running` and check status, logs, stop, and command escape paths
- Start at `drift_detected` and `unknown_error` and judge recovery wording
- Confirm `attach` and `detach` never appear as first-level labels
- Confirm `show commands` exposes the CLI equivalents
- Confirm `not_project` and multi-project scope notes do not imply a project switcher

## Notes

- The prototype uses canned fixtures only.
- It does not read real state, run Docker, or write `.env.stageserve`.
- `edit project settings` shows a path and sample content only.
