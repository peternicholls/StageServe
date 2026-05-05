# Top-Level Routing: What Bare `stage` Does

This is the first decision StageServe makes when the user types `stage` and presses enter. Nothing else in this folder runs until this routing has chosen a path.

## Inputs Available At This Point

- The exact command line (`stage`, `stage --help`, `stage up`, `stage init --json`, etc.).
- Whether stdin and stdout are interactive terminals.
- Whether `STAGESERVE_NO_TUI=1` is set in the shell environment.
- Whether `NO_COLOR=1` is set.
- Whether `--notui` or `--cli` is on the command line.
- The current working directory.

## Output Of This Step

One of:

- Open the guided TUI in this terminal.
- Print plain text guidance and exit (no prompts, no input).
- Print Cobra help and exit.
- Run a direct subcommand and exit.

This step does not touch Docker, the planner, or any project files.

## Routing Table

| Command Line | TTY? | `STAGESERVE_NO_TUI`? | `--notui` or `--cli`? | Result |
|---|---|---|---|---|
| `stage` | yes | no | no | Open guided TUI |
| `stage` | yes | yes | no | Print text guidance, exit 0 |
| `stage` | yes | no | yes | Print text guidance, exit 0 |
| `stage` | no | any | any | Print text guidance, exit 0 |
| `stage --help` or `stage -h` | any | any | any | Show Cobra help, exit 0 |
| `stage <subcommand>` | any | any | any | Run that subcommand directly |
| `stage <subcommand> --json` | any | any | any | Run that subcommand, JSON output, no TUI |

`STAGESERVE_NO_TUI` is read only from the shell environment. It is intentionally not honoured if it appears in `.env.stageserve`, because `.env.stageserve` is project/stack config, and TUI behaviour is a per-invocation operator preference.

`--cli` is an alias for `--notui` so users can pick whichever word feels right.

## Mockup: Guided TUI Path

The user types `stage` in an interactive terminal. After the routing step decides "open the guided TUI", control passes to the planner state machine described in [03-state-machine.md](./03-state-machine.md). The first screen the user sees depends on what the planner finds. Examples are in [04-machine-setup.md](./04-machine-setup.md), [05-project-setup.md](./05-project-setup.md), and [06-run-stop-inspect.md](./06-run-stop-inspect.md).

## Mockup: Plain Text Guidance Path (Non-TTY Or TUI Disabled)

```
$ stage
StageServe 0.7.0

Your computer isn't ready yet.
  StageServe needs Docker Desktop and a local DNS resolver.

Next step:
  Run: stage setup

Other commands:
  stage init     set up the current folder as a project
  stage up       run this project
  stage status   see what's running
  stage --help   full command list

$
```

Rules for the plain text path:

- One short status line at the top.
- One named "Next step" with the literal command to type.
- A short list of other commands, in the order a new user is most likely to need them.
- Exit 0 unless context collection itself fatally failed.

This path never prompts. It is safe to redirect, pipe, or run from CI.

## Mockup: Direct Subcommand Path

```
$ stage up
... existing stage up output ...
$
```

The routing step recognises `up` as a subcommand and never enters the TUI. The same is true for every documented direct subcommand listed in `contracts/guided-tui-contract.md`.

## Edge Cases Handled At This Layer

- `stage` is run inside a tmux/screen session that reports limited colour: still opens the TUI, but `NO_COLOR` rules apply if set.
- `stage` is run with stdout piped to `tee`: stdout is non-TTY, so the plain text guidance path is used.
- `stage` is run from an editor "Run" button that allocates a TTY for stdout but not stdin: treated as non-TTY for this routing step. The user still gets the plain text guidance path; they can re-run from a real terminal.
- `stage --notui --json`: `--json` only makes sense for subcommands. The bare-`stage` form with `--json` is documented as "use a subcommand if you need JSON output" in the plain text guidance.

## What This Layer Does Not Do

- It does not detect machine readiness.
- It does not read `.env.stageserve`.
- It does not look for project files.
- It does not call Docker.
- It does not write anything.

All of that belongs to layers downstream. This layer's only job is to decide "guided, text, help, or direct".
