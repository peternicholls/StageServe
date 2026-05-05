# Power-User Paths

Spec 007's primary user is the simple-first user. This file is for everyone else: scripts, CI, and people who already know what they want and want StageServe to get out of their way.

The rule is: **the simple-first user gets the guided path; the power user gets every direct command, every flag, every file.** The two paths must not interfere with each other.

## Opting Out Of The Guided TUI

There are three equivalent ways to opt out of the guided TUI for an invocation:

| Opt-Out | Scope |
|---|---|
| `--notui` flag on the command line | This invocation only |
| `--cli` flag on the command line (alias for `--notui`) | This invocation only |
| `STAGESERVE_NO_TUI=1` in the shell environment | All invocations in that shell session |

`STAGESERVE_NO_TUI` is read **only** from the shell environment. It is intentionally not honoured in `.env.stageserve`, because TUI behaviour is a per-operator preference, not project or stack config.

There is no `--tui` flag. TUI is the default in interactive terminals; the opt-outs above turn it off.

## When The Guided TUI Is Skipped Automatically

The guided TUI is skipped without any flag when:

- stdout is not a terminal (`stage > out.txt`, `stage | tee log`, etc.).
- stdin is not a terminal (`echo y | stage`, etc.).
- the user typed `stage --help` or `stage -h`.
- the user typed any direct subcommand (`stage up`, `stage init --json`, etc.).

In each of these, StageServe falls through to plain text or to the requested subcommand. See [02-top-level-routing.md](./02-top-level-routing.md) for the full table.

## Direct Command Surface

The full direct command list is documented in `contracts/guided-tui-contract.md`. The diagrams in this folder do not redefine it. Every action a guided screen performs has an equivalent direct command, and any user who knows the direct command can skip the guided path entirely.

The `Show direct commands` panel under `More…` on any guided screen lists the direct commands relevant to the current situation. That panel is the bridge for users who want to learn the CLI as they go.

## What `.env.stageserve` Owns

`.env.stageserve` is the only user-editable StageServe config surface. There are two scopes:

| Scope | Path | Purpose |
|---|---|---|
| Stack-wide | `<stack-home>/.env.stageserve` | Defaults that apply to every project on this machine |
| Project-local | `<project>/.env.stageserve` | Overrides for one specific project |

Precedence (highest first):

1. Command-line flags
2. Project-local `.env.stageserve`
3. Shell environment
4. Stack-wide `.env.stageserve`
5. Built-in defaults

Power users can edit either file in their text editor. The guided TUI's project setup flow ([05-project-setup.md](./05-project-setup.md)) writes only the project-local file, and only with explicit confirmation. It never edits the stack-wide file.

## What `.stageserve-state` Holds (And Why The Simple User Never Sees It)

`.stageserve-state` lives under the stack home. It contains:

- Per-project recorded state (running, down, last started, route, etc.).
- Generated runtime env files for projects.
- Generated routing/gateway state for the shared front door.
- Working files StageServe needs to keep its records consistent.

The simple-first user never has to look at this folder. The guided TUI never asks the user to edit anything in it. If the planner ever produces a recovery step that requires touching it, the step is described in plain language ("forget the running record") and the file change is internal — the user never sees the path.

Power users can inspect `.stageserve-state` directly. The advanced view in `More… → Advanced and troubleshooting` shows the path so an operator can `cd` into it.

## What Stays Hidden Even From Power Users In Project Directories

Inside a project directory, the only StageServe-owned file is `.env.stageserve`. There are no:

- Compose files
- Dockerfiles
- Container definitions
- Network or volume definitions
- Generated build outputs

All of those live under `.stageserve-state` (machine-generated, hidden) or under the StageServe binary itself (built-in templates). They are accessible to the power user who goes looking for them in the advanced view, but they are never written into the project directory.

This is a hard rule from spec 004: project directories stay clean. The guided TUI flows in this folder do not break it.

## The `More…` Footer Screen

Every guided screen offers `More…` as a non-default option. When the user picks it, they get a small power-user surface for the current screen:

```
StageServe 0.7.0  More options for this screen

▶ Show direct commands
    See the equivalent stage commands for what this screen does

  Switch to plain text output
    Re-run this command with --cli for plain text

  Open the project's settings file
    Show the path to .env.stageserve and how to edit it by hand

  Open StageServe's hidden working folder
    Show where StageServe keeps its records (advanced)

  Advanced and troubleshooting
    The full diagnostic view, including check names, internal IDs,
    Docker terminology, and exact recovery commands

  ↑/↓ navigate • enter open • esc back
```

Notes:

- `Show direct commands` is the bridge for users who want to learn the CLI from the guided UI.
- `Open the project's settings file` and `Open StageServe's hidden working folder` show paths; they do not open external editors.
- `Advanced and troubleshooting` is the only place internal vocabulary (`drift`, `gateway`, `compose`, `container`, `bootstrap`, `attach`, `detach`, `state`) is allowed to appear.

## Advanced And Troubleshooting View

This is the one screen in the guided TUI where the plain-language rules are relaxed. It is opt-in (the user has to walk through `More… → Advanced and troubleshooting` to get to it). Its purpose is to give a power user the precise internal information they need without forcing every other user to see it.

What it shows for the current situation:

- The internal planner situation name (for example `drift_detected`).
- The internal step IDs (`docker_cli`, `dns_resolver`, etc.) and their last known status.
- The exact lifecycle classification (for example `post-up-hook failure`).
- The exact `stage` commands that match the recovery options.
- The path to `.env.stageserve` for the current project.
- The path to `.stageserve-state` and the relevant per-project record file.
- The exact compose project name and resource names that StageServe is using internally for this project.

This is the only place those names appear in the TUI. They are never used in any first-level guided screen.

## How The Guided And Direct Surfaces Stay Consistent

- The planner is the single source of truth for "what situation is this?".
- The lifecycle/onboarding/status modules are the single source of truth for "what should happen?".
- The guided TUI does not own any logic. It renders planner output in plain language and calls existing commands to act.
- The direct CLI does not own any extra logic either. It calls the same modules with the same arguments.

This means a user can move freely between the guided path and the direct path within the same session. They will not see different answers depending on which surface they used.
