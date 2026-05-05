# Original Intent From Spec 004 And Spec 005

This file is the source of authority for the diagrams in this folder. Where a diagram makes a choice, it should be traceable back to one of these statements. The implementations completed under specs 004 and 005 partially departed from the original intent in two ways: the simple-first guided surface was not built (only its prerequisites were), and onboarding commands grew TUI projections without the planner-driven entry that bare `stage` was meant to provide.

## From The Project Constitution

- **Ease of use is a product requirement.** The primary operator experience must stay simple enough to use from memory.
- **Reliability must be boring and predictable.** Repeated runs and config precedence must be deterministic and visible.
- **Robustness must hold under real failure.** Failures must be actionable through status, logs, health checks, or recovery instructions.
- **Remove pinch points.** New steps, prompts, and stateful exceptions must remove recurring pain rather than move it elsewhere.

## Pulled From Spec 004 (Workflow And Lifecycle)

Spec 004 hardened the lifecycle. It deliberately scoped TUI changes out, but it also locked in the rules that the simple-first surface must respect:

- One bootstrap phase only, and it runs only after StageServe-owned readiness succeeds.
- `STAGESERVE_POST_UP_COMMAND` is sourced only from project-root `.env.stageserve`.
- `.env.stageserve` is the canonical stack-owned defaults file.
- Project `.env` remains application-owned.
- StageServe-owned runtime resource names use `stage-<slug>`.
- Shared routing is StageServe-managed and named `stage-shared`/`stage-gateway`.
- Failures must be classified by phase, not lumped together.
- Project rollback must be project-scoped; one project's failure must not affect another.
- Operator cancellation (`Ctrl-C`) during the bootstrap phase must abort and roll back cleanly under the same step name.

Implication for the simple-first surface: every flow in this folder must trust the lifecycle classification, never duplicate it, and never invent an extra step. The TUI's job is to show those classifications in plain language.

## Pulled From Spec 005 (Installer And Onboarding)

Spec 005 built the underlying onboarding mechanics — `stage setup`, `stage init`, `stage doctor`, normalized step results, JSON output, exit codes 0/1/2/3 — but it explicitly listed "GUI/TUI onboarding experiences" as out of scope. As a result, the planner-driven simple-first surface that bare `stage` was meant to expose never landed. Spec 007 is finishing that job.

Locked-in pieces from spec 005 that the simple-first surface must reuse:

- `stage setup` for machine readiness, idempotent, normalized step statuses (`ready`, `needs_action`, `error`).
- `stage init` for project `.env.stageserve`, never overwriting without explicit confirmation.
- `stage doctor` for read-mostly diagnostics.
- Privileged operations are explicit and bounded; nothing escalates silently.
- All three commands must support text, JSON, and TUI modes through one projector.
- Exit codes: 0 ready, 1 needs action, 2 error, 3 unsupported OS.
- `install -> setup -> init -> up` as the documented onboarding sequence.

Implication for the simple-first surface: the bare `stage` TUI is a thin orchestrator over these existing commands. It should never run its own setup logic, never write its own config, never invent its own diagnostic checks.

## What The Departure Was

Spec 004 did not build a TUI. Spec 005 added TUI projections to three onboarding commands but did not add the bare-`stage` guided entrypoint. The current effect is:

- A new user runs `stage` with no subcommand and gets help text, not the guided surface that the constitution and the spec 005 documentation roadmap implied.
- Each onboarding command has its own TUI projection, but there is no single screen that detects context and offers the right next action.
- Power-user terminology (`attach`, `detach`, `doctor`, `compose`, `gateway`) leaks into first-level docs and command help because there has been no easy-mode language gate.

Spec 007 is the catch-up work. The flow diagrams in this folder are the design that catch-up should follow.

## What Stays Off-Limits Even In The Catch-Up

These are not negotiable, because specs 004 and 005 already locked them in:

- The TUI does not write any user-editable config file other than `.env.stageserve`.
- The TUI does not expose Docker, compose, gateway, container, or volume names in any first-level user message. They live only in the advanced view.
- The TUI does not bypass lifecycle rollback semantics.
- The TUI does not introduce its own state files. It reads what `.stageserve-state` holds and shows it through plain language.
- The TUI does not silently elevate privileges. Any privileged action is explicit, named in plain language, and confirmed.
- The TUI does not fork the JSON output contract. JSON modes remain pure.

## What This Means For Every Diagram In This Folder

Each diagram is a translation layer. It takes existing planner output, existing onboarding step results, existing lifecycle outcomes, and existing config, and renders them in language a normal user can act on. The diagrams describe what the user sees and chooses; they never describe new mechanics that StageServe would have to build from scratch.

If a diagram looks like it is asking StageServe to do something new, that is the wrong diagram. The correct diagram is the one that takes the existing capability and makes it easier for the user to use.
