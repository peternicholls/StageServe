# Guided Easy Mode Prototype

This is a tracked, non-production prototype for spec 007. It exists to test the easy-mode flow, copy, keyboard model, and visible-default rules before production implementation.

The prototype is fixture-only. It does not read real StageServe state, run Docker, touch local DNS, or write `.env.stageserve`.

## Run

```bash
go run ./specs/007-harden-TUI-and-other-interactions/prototype --list-scenarios
go run ./specs/007-harden-TUI-and-other-interactions/prototype
go run ./specs/007-harden-TUI-and-other-interactions/prototype --scenario project_running
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario project_missing_config
go run ./specs/007-harden-TUI-and-other-interactions/prototype --cli --scenario drift_detected
```

## What It Demonstrates

- Four surfaces: status header, decision bar, tool work panel, persistent footer.
- Machine setup as a tool-owned checklist, not a menu.
- `.develop` local URL examples with visible site name, web folder, suffix, and URL before any write.
- Project settings preview and confirmation before `.env.stageserve` would be written.
- Running-project default is non-destructive: `enter` opens logs, not stop.
- Out-of-sync recovery previews the safe step and confirms before changing records.
- Direct commands and advanced troubleshooting live behind footer paths.

## Controls

- `up` / `down`: move through decision items.
- `enter`: run the highlighted/default item.
- `?`: show plain-language detail.
- `m`: show direct command equivalents.
- `a`: show advanced/troubleshooting detail.
- `tab` / `shift+tab`: switch between canned scenarios.
- `q`: quit.

Project edit screen:

- `up` / `down`: choose field.
- `enter`: cycle sample values.
- `s`: save edits back to preview.
- `esc`: discard edits.

Logs screen:

- `q` / `esc`: exit logs and return to the running-project screen.

## Verification

```bash
go test ./specs/007-harden-TUI-and-other-interactions/prototype
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario machine_not_ready
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario project_missing_config
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario project_running
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario drift_detected
go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario unknown_error
```

Manual TTY checks:

- Start at `machine_not_ready`; confirm setup is a checklist and `Find issues` is not a decision item.
- Start at `project_missing_config`; confirm defaults and local URL are visible before confirmation.
- Use `Edit before writing`; confirm edits return to the preview and do not write.
- Start at `project_running`; confirm `enter` opens logs and stop requires confirmation.
- Start at `drift_detected`; confirm the safe step previews what changes before applying.
- Press `m`; confirm direct commands are discoverable but not mixed into the decision bar.
