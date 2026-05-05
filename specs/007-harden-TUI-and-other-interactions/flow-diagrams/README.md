# Spec 007 Flow Diagrams

This folder holds textual flow diagrams for the spec 007 guided TUI redesign. Each file covers one concept in isolation so we can review and revise them piece by piece.

The diagrams are deliberately text-only. They use ASCII screen mockups, indented step lists, and small state-transition tables. No Mermaid, no SVGs, no images.

## Files

- [00-vocabulary.md](./00-vocabulary.md) — plain-language word list. Defines what we say to the user and what we never say. Read this first.
- [01-intent-from-004-005.md](./01-intent-from-004-005.md) — original intent pulled together from spec 004, spec 005, and the project constitution. Justifies every diagram in this folder.
- [02-top-level-routing.md](./02-top-level-routing.md) — what bare `stage` does, how it decides which screen to show, and how non-TTY/automation paths still get a useful answer.
- [03-state-machine.md](./03-state-machine.md) — the planner state machine: every situation, every transition, every default action.
- [04-machine-setup.md](./04-machine-setup.md) — first-install and machine-readiness flow. Tool-driven checklist, not a menu.
- [05-project-setup.md](./05-project-setup.md) — per-project setup flow. `.env.stageserve` preview, edit, confirm.
- [06-run-stop-inspect.md](./06-run-stop-inspect.md) — day-2 menu when the project is configured. Run, stop, view logs, open in browser.
- [07-out-of-sync.md](./07-out-of-sync.md) — what to do when StageServe finds something it cannot trust. Replaces the previous "drift" framing with plain language.
- [08-recovery-and-help.md](./08-recovery-and-help.md) — recovery flow when StageServe cannot decide a safe next step on its own.
- [09-power-user-paths.md](./09-power-user-paths.md) — how power users opt out, where hidden artifacts live, and how `.env.stageserve` remains the only user-editable surface.
- [10-defaults-rules.md](./10-defaults-rules.md) — the rule that every screen must show the values it will use before the user commits.

## Reading Order

1. Vocabulary (so the language rules are clear).
2. Intent (so the justification is clear).
3. Top-level routing (so the entry point is clear).
4. State machine (so the situations are clear).
5. Each individual flow (4 through 10) in any order.

## Hard Rules That Apply To Every Diagram

- Plain language only. No `drift`, `gateway`, `compose`, `container`, `daemon`, `attach`, `detach`, `runtime`, `state`, `registry`, or `bootstrap` in any first-level user-visible text.
- The tool drives. The user confirms. The user only chooses when there is a real alternative goal, not a real alternative tool action.
- Defaults are visible. Every screen with a value shows that value before the user is asked to commit.
- The user is never left at a dead end. Every screen has a documented next action, even when StageServe itself cannot decide.
- `.env.stageserve` is the only user-editable StageServe config surface. Hidden artifacts under `.stageserve-state` exist for the power user to inspect, but no normal user flow requires touching them.
- Docker, compose, and image vocabulary do not appear in any project directory or any first-level user message. They appear only in the optional advanced/troubleshooting view.
