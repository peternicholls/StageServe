# Spec 007 Flow Redesign: Tool-Driven Easy Mode

## Why This Document Exists

The prototype under [prototype/](./prototype) showed a structural design problem, not a copy problem:

- Easy mode was rendered as a menu of peer actions (`Set up this computer`, `Find issues`, `Show commands`).
- That menu shape made the user the orchestrator and asked them to know when to run diagnostics or fall back to commands.
- "Set up this computer" then expanded into more menus, when it should have been a tool-owned checklist that executes step-by-step on the user's behalf.

The original spec 007 intent in [original-intentions-and-decisions.md](./original-intentions-and-decisions.md) is:

> The intended first-level experience is a simple guided StageServe surface that helps a normal user install, set up the machine, set up a project, run it, stop it, inspect it, and recover from problems without needing to understand the Docker implementation.

The previous prototype interaction model violated that intent in three concrete ways. This document fixes it before any production code is written.

## Diagnosis: Three Concrete Design Failures

### Failure 1: Diagnostics Were Made A User Choice

`Find issues` (`stage doctor`) was a peer secondary action under every situation. Easy-mode users do not know when running doctor is appropriate. Diagnostics are something the tool runs on the user's behalf when a blocker is detected — never a menu item the user is expected to reach for.

This contradicts spec 007 SC-008 in [spec.md](./spec.md), which requires that easy-mode labels describe user goals rather than implementation mechanics. `Find issues` is a tool-mechanic.

### Failure 2: `Show commands` Was Treated As An Action

`Show commands` belongs to power-user transparency, not the easy-mode action ladder. Putting it in the same list as "Run this project" forces the easy-mode user to scan options that are not theirs.

`docs/runtime-contract.md` and the [guided TUI contract](./contracts/guided-tui-contract.md) treat direct CLI commands as a parallel surface for power users. The TUI must reflect that separation in its layout, not by listing both surfaces side by side.

### Failure 3: Tool Setup Was Modelled As A Navigable Scenario Tree

The eight required situations in [contracts/guided-tui-contract.md](./contracts/guided-tui-contract.md) are *detection states the planner produces*, not *screens the user steps through*. The previous prototype treated them as nodes the user could traverse and inserted scenario-shaped sub-menus for `Set up this computer`. Setup is not a menu — it is an ordered checklist the tool executes, prompting the user only at concrete privileged or external steps (Docker Desktop install, DNS resolver write, certificate trust).

## Revised Interaction Model

### Three Surfaces, Not One Menu

Easy mode separates three concerns that the previous prototype merged into one menu:

| Surface | Owner | Purpose | Trigger |
|---|---|---|---|
| Status header | Tool | Tells the user where they are right now in plain language | Always present |
| Decision bar | User | One primary action, at most one alternative meaningful choice | Only when a real user choice exists |
| Tool work panel | Tool | Runs setup steps, shows progress, surfaces blockers with concrete next instructions | Driven by tool, not user navigation |
| Footer (persistent) | Both | Quit, back, help, show commands, advanced/troubleshooting | Always available, never in the decision bar |

`Find issues`, `Show commands`, `Edit project settings`, and `Advanced troubleshooting` are footer affordances. They are not peer actions to "Run this project".

### The Tool Drives, The User Confirms

Easy-mode interaction reduces to three event types the user ever sees:

1. **Watching** — the tool is detecting state or running a step. Spinner, progress, or short status line. No input expected.
2. **Confirming** — the tool needs a yes/no for a mutation (write `.env.stageserve`, run `stage up`, stop project). One key.
3. **Choosing** — the tool offers a real alternative ("Run project now" vs "Edit settings first"). Two or three labelled goals, never tool-mechanics.

The previous prototype had a fourth event type — *navigating between scenarios* — which is what produced the menu-of-menus shape. That event type is removed.

### Tool-Owned Setup Checklist

When the planner detects `machine_not_ready`, the easy-mode UI does not show a "Set up this computer" menu item. It enters a tool-owned setup phase that runs the existing `core/onboarding` checks (Docker binary, Docker daemon, state dir, ports, DNS, mkcert) sequentially, surfacing each check's result and only prompting the user at external blockers.

Each external blocker presents one of:

- a one-keystroke confirmation when StageServe can perform the action itself
- explicit copy-pasteable instructions when the action is outside StageServe's authority (install Docker Desktop, click Allow on a TLS prompt)

The user never sees `Find issues` as an option here. If a check fails non-recoverably, the tool itself surfaces the equivalent of doctor output inline, with the next concrete instruction.

### Planner Situations Become Display Headings, Not Menus

The eight situations from the contract remain as planner output. They map to the status header in plain language, and they determine which decision bar to show — but they are no longer screens the user navigates between. The flow stack of scenarios from the previous prototype is removed.

### Defaults Are Always Visible

The Ollama TUI shows the active default inline next to its action label (`Chat with a model (gemma4:31b)`). StageServe must do the same: every screen that has a sensible default must surface that default inline, before the user presses anything. This includes:

- The currently-selected model on a chooser.
- The proposed project route on the init preview (`pete-site.stage.local`).
- The current docroot path (`./public_html`).
- The active TLS source and trust state.
- The currently-detected DNS resolver mode.
- For decision bars: which action is highlighted as the default if the user just presses `enter`.

A user must never have to drill into a sub-screen to discover what the tool is going to do for them. If the tool has a value to use, it shows that value on the screen where the choice is offered.

## Screen Mockups (Ollama-Inspired)

Each mockup represents one full TUI screen at one moment. The visual model follows the Ollama TUI shown in the user's reference screenshots: a short status line at the top, a small list of plain-language items with one-line descriptions, an inline default shown in brackets next to the relevant item, and a single footer line of keys.

Convention used in the mockups below:

- `▶` marks the highlighted item that `enter` will activate.
- `(default)` after a value means it is what the tool will use if the user just presses `enter`.
- Items in `[brackets]` after a label are inline default values for that choice.
- The footer line is always present and always identical across screens of the same kind.
- Items below `More…` are progressive disclosure, exactly like Ollama's `More…` row.

### Screen: Bare `stage` — Project Ready To Run

```
StageServe 0.7.0  /Users/pete/sites/pete-site

▶ Run this project                       [pete-site.stage.local]
    Start the project and open it in your browser

  Edit project settings                  [./public_html, suffix .stage.local]
    Change the site name, web folder, or domain suffix

  More…
    Show direct commands, advanced and troubleshooting

  ↑/↓ navigate • enter run • → configure • esc quit
```

Notes:

- The default route, docroot, and suffix appear inline so the user knows what `enter` will do.
- `More…` hides power-user affordances (`Show direct commands`, `Advanced and troubleshooting`) until the user opts in. They are not peer actions in the primary list.
- `Find issues` does not appear. The planner has determined the project is ready; if it were not, a different screen would be shown.

### Screen: Bare `stage` — Machine Not Ready

```
StageServe 0.7.0  Setting up your computer

  Docker Desktop                                            ✓ ready
  State folder                                              ✓ ready
  Local DNS resolver                                        ✓ ready
▶ Local HTTPS certificates                                  needs your approval
    StageServe will install a local certificate authority so
    https://*.stage.local works without browser warnings.
    Default action on enter: install and trust the local CA.

  Network ports 80 and 443                                  pending

  enter approve • s skip this step • → details • esc quit
```

Notes:

- The tool drives the checklist and only stops on the first item that needs the user.
- The default action on `enter` is spelled out in plain English on the active row.
- Other rows show their state (`✓ ready`, `pending`, `needs your approval`) so the user can see progress and what is left.
- No `Find issues` action appears: this screen *is* the find-issues experience.

### Screen: Machine Not Ready — Blocker Outside StageServe's Authority

```
StageServe 0.7.0  Setting up your computer

  Docker Desktop                                            not installed

    StageServe needs Docker Desktop to run your sites.

    1. Open https://www.docker.com/products/docker-desktop
    2. Install Docker Desktop for macOS
    3. Open Docker Desktop and wait for the whale icon to settle

    When you've done this, press enter and StageServe will check again.

  enter check again • → why does StageServe need this • esc quit
```

Notes:

- A single concrete instruction with numbered steps is shown.
- The default action (`enter` re-checks) is spelled out in the body, not just the footer.
- `→ why does StageServe need this` is progressive disclosure for users who want to understand before acting.

### Screen: Bare `stage` — Project Missing Config (Preview)

```
StageServe 0.7.0  /Users/pete/sites/pete-site

  This folder doesn't have StageServe settings yet.
  StageServe will create one file: .env.stageserve

  Site name           pete-site                       (default)
  Web folder          ./public_html                   (default)
  Domain suffix       .stage.local                    (default)
  Local URL           https://pete-site.stage.local   (preview)

▶ Use these settings
    Write .env.stageserve and continue

  Edit before writing
    Change site name, web folder, or domain suffix first

  ↑/↓ navigate • enter confirm • → edit one field • esc cancel
```

Notes:

- All defaults are visible *before* the user picks a path.
- `Local URL` is shown as a preview so the user knows what the result will look like.
- `→ edit one field` jumps straight into an inline editor for the highlighted row, which is how Ollama's `→ configure` works.

### Screen: Bare `stage` — Project Running

```
StageServe 0.7.0  pete-site is running

  https://pete-site.stage.local                             ↗ open
  Started 4 minutes ago • Apache + PHP 8.3 • healthy

▶ Stop this project
    Free up the local URL and shut down the project

  View project logs
    Watch the live log stream until you press q

  More…
    Show direct commands, restart, advanced and troubleshooting

  ↑/↓ navigate • enter run • → open in browser • esc quit
```

Notes:

- The live URL is shown at the top with the open hint.
- The default action (`enter` stops) is the most likely user goal at this point and matches the highlighted row.
- `Restart` is hidden under `More…` because it is rarely the right first instinct.

### Screen: Bare `stage` — Drift Detected

`drift_detected` is a first-class planner situation, not an edge case. It fires when the recorded state, runtime state, DNS, gateway, or config disagree. The user must be able to see *what* disagrees, what StageServe thinks the safe next step is, and confirm or override that step. This screen exists for exactly that.

```
StageServe 0.7.0  pete-site looks out of sync

  StageServe expected this project to be running, but the gateway
  is not routing to it. The project's container is not listed.

  What StageServe found
    Recorded as running          yes
    Container present            no
    Gateway route present        no
    DNS for pete-site.stage.local resolved

  Safe next step (default on enter)
    Forget the recorded run and treat the project as stopped.
    Nothing on disk will be deleted.

▶ Use the safe next step                 [forget recorded run]
  Try to start the project again         [keeps existing settings]
  Show what disagreed in detail
  More…
    Show direct commands, advanced and troubleshooting

  ↑/↓ navigate • enter confirm • → details • esc quit
```

Notes:

- Drift is treated as a real situation with its own dedicated screen. It is not delegated to a generic `Find issues` action.
- The default-on-enter action is the lowest-risk recovery and is named in the body before the user sees the action list.
- `Show what disagreed in detail` is the user's escape hatch when the summary is not enough; it is *not* the default and is *not* required for the safe path.

### Screen: Bare `stage` — Not A Project

```
StageServe 0.7.0  /Users/pete/Downloads

  This folder isn't a StageServe project yet.

▶ Set up this folder as a project        [name: Downloads, suffix .stage.local]
    StageServe will create a .env.stageserve and propose a local URL

  Pick a different folder
    Type a path to look at instead

  More…
    Show direct commands, advanced and troubleshooting

  ↑/↓ navigate • enter set up • → details • esc quit
```

Notes:

- The proposed defaults for the new project (name from folder, suffix from machine settings) are visible inline before the user commits.
- `Pick a different folder` keeps the user from dead-ending in an unintended directory.

### Screen: Footer Affordances (Persistent)

The footer is the same set of keys on every screen of the same kind, so the user learns it once. It is rendered as a single dim line, never as a peer of decision-bar items.

```
  ↑/↓ navigate • enter <verb> • → <reveal> • esc quit
```

When the user opens `More…` (or presses the discoverable key bound to it), they see:

```
StageServe 0.7.0  More options for this screen

▶ Show direct commands
    See the equivalent stage commands for what this screen does

  Advanced and troubleshooting
    Open stage doctor, log streams, and runtime details

  Switch to the command-line interface
    Re-run this command with --cli for plain text output

  ↑/↓ navigate • enter open • esc back
```

Notes:

## Screen Transition Map

A plain list of which screen leads where. The planner re-runs after every action that changes state, so most arrows return to "re-detect" rather than to a hard-coded next screen.

```
bare stage
    -> planner detects situation
        machine_not_ready       -> Setup checklist
        project_missing_config  -> Init preview
        project_ready_to_run    -> Project ready
        project_running         -> Project running
        project_down            -> Project down
        drift_detected          -> Drift screen
        not_project             -> Not a project
        unknown_error           -> Recovery screen

Setup checklist
    enter on a "needs your approval" row -> tool runs that step -> re-detect
    enter on a "not installed" row       -> show external instructions -> re-detect on next enter
    s skip                               -> mark step skipped, move to next -> re-detect at end

Init preview
    enter (Use these settings)   -> write .env.stageserve -> re-detect
    enter (Edit before writing)  -> inline edit -> back to Init preview with new defaults
    -> on a row                  -> inline editor for that one field -> back to Init preview
    esc                          -> leave files unchanged -> exit

Project ready
    enter (Run this project)     -> stage up -> re-detect (likely Project running)
    enter (Edit project settings)-> Init preview with current values
    More...                      -> footer screen
    esc                          -> exit

Project running
    enter (Stop this project)    -> confirm -> stage down -> re-detect (likely Project down)
    enter (View project logs)    -> log stream until q -> back to Project running
    -> open in browser           -> launch URL -> stay on Project running
    More...                      -> footer screen
    esc                          -> exit (project keeps running)

Project down
    enter (Run this project)     -> stage up -> re-detect
    Remove from StageServe       -> confirm -> deregister -> re-detect (likely Not a project)
    More...                      -> footer screen
    esc                          -> exit

Drift screen
    enter (Use the safe next step)        -> tool applies safe recovery -> re-detect
    enter (Try to start the project)      -> stage up -> re-detect
    Show what disagreed in detail         -> details screen -> back to Drift screen
    More...                               -> footer screen
    esc                                   -> exit, drift unresolved

Not a project
    enter (Set up this folder)   -> Init preview (with proposed defaults)
    Pick a different folder      -> path prompt -> re-detect at new path
    More...                      -> footer screen
    esc                          -> exit

Recovery screen (unknown_error)
    follows the same pattern as Drift screen, but lists ordered steps
    the tool will run on the user's behalf, with the lowest-risk step as default

Footer screen (More...)
    Show direct commands         -> command panel for this situation -> back
    Advanced and troubleshooting -> doctor / logs / runtime details -> back
    Switch to CLI                -> exit and re-run with --cli
```

## Concrete Easy-Mode Wording Rules

These rules apply to every status header, decision bar, and confirmation prompt:

- Status header is one sentence describing what is true now, in plain English.
- Decision bar items are user goals (`Run this project`, `Stop this project`, `Use these settings`), never tool actions (`Run stage up`, `Run doctor`).
- Tool work panels describe what the tool is doing in plain language (`Checking Docker`, `Writing .env.stageserve`), never command names.
- Blockers describe the external problem and the concrete next instruction (`Docker Desktop is not running. Open Docker Desktop, then press any key to retry.`), never the tool mechanic.
- Every screen with a default value or default action shows that default inline, before the user has to drill in. The user can always see what `enter` will do.
- The footer is the only place where command names appear by default.

## Mapping To Existing Planner Situations

The eight situations from [contracts/guided-tui-contract.md](./contracts/guided-tui-contract.md) remain authoritative for what the planner produces. Their UI mapping changes:

| Situation | Status header | Decision bar | Tool work panel |
|---|---|---|---|
| `machine_not_ready` | "This computer is not ready yet." | hidden | tool-owned setup checklist runs |
| `project_missing_config` | "This project does not have StageServe settings yet." | Use these settings / Edit before writing | preview panel above decision bar |
| `project_ready_to_run` | "This project is ready to run." | Run this project / Edit settings first | hidden |
| `project_running` | "This project is running at \<route\>." | View project logs / Stop this project | hidden |
| `project_down` | "This project is stopped." | Run this project / Remove from StageServe | hidden |
| `drift_detected` | "Something about this project does not match what StageServe expects." | hidden | tool runs diagnostics inline |
| `not_project` | "This directory is not a StageServe project yet." | Set up this directory / Get setup help | hidden |
| `unknown_error` | "StageServe could not safely choose a next step." | hidden | tool runs ordered recovery inline |

Everywhere the decision bar is hidden, the user is watching the tool work and is offered cancel/back through the footer. They are never asked to pick `Find issues`.

## What This Removes From The Previous Prototype

- The flow stack of scenarios.
- The "navigate to next scenario" action kind.
- `Find issues` as a peer secondary action under every situation.
- `Show commands` as an action ladder item.
- `Edit project settings` as a peer action separate from the init preview.
- The completed-work trail (re-detection after each tool action gives the user a fresh, accurate status header instead).

## What This Preserves From The Spec

- The eight planner situations remain authoritative.
- `core/guidance` planner stays the single non-UI decision layer (FR-005).
- `core/onboarding` continues to own setup/doctor mechanics; the TUI just orchestrates them.
- Direct command equivalents remain available through the footer's `show commands` (FR-011).
- `--notui` and `--cli` text fallback covers the same situations and decisions, but linearizes them as plain prose with one decision prompt at a time (NFR-004).
- `STAGESERVE_NO_TUI=1`, `NO_COLOR=1`, JSON purity, and direct-subcommand bypass behavior are unchanged (FR-012, FR-013, FR-014).

## Self-Review Against Prior Research

Checked against [research.md](./research.md) and the patterns it endorses:

- **DDEV no-args dashboard** — matched: bare `stage` shows status and at most a small set of meaningful next actions, never a full command map.
- **Fly Launch preview-then-confirm** — matched: project init shows preview before any write, with explicit edit-or-accept choice.
- **Vercel and GitHub CLI automation safety** — preserved: the redesign does not change non-TTY, JSON, or direct-subcommand behavior.
- **Bubble Tea / Huh accessibility** — improved: removing the menu-of-menus shape reduces cursor management, and the inline form for init replaces a separate scenario.

Checked against [original-intentions-and-decisions.md](./original-intentions-and-decisions.md):

- "A primary simple user is never left at a dead end" — preserved by the always-present footer and re-detection after every tool action.
- "Power users can bypass the guided path" — unchanged; direct subcommands still bypass the TUI.
- Easy-mode language rule — strengthened by removing tool-mechanic items from the decision bar entirely.

Checked against [recovery-plan.md](./recovery-plan.md) Phase C exit criteria:

- "A user can start from bare stage, understand context, and choose an action" — supported.
- "A user can understand the first screen without knowing attach/detach, gateway, compose, container, registry, runtime, or state terminology" — supported, since those terms are now confined to footer-driven advanced/troubleshooting and direct CLI help.

## Implications For Tasks And Contract

When this redesign is accepted, the following spec artifacts need a paired update:

- [contracts/guided-tui-contract.md](./contracts/guided-tui-contract.md): replace the "Required Guided Situations" table's `Secondary Actions` column with the surface mapping above. The current column lists `find issues, show commands, quit` as peer secondaries, which is exactly the shape this redesign rejects.
- [tasks.md](./tasks.md): T032, T038–T041, and T041d need to be reframed around the three-surface model rather than action lists.
- [data-model.md](./data-model.md): `GuidedAction.Kind` should be narrowed to `confirm`, `choose`, `tool_step`, `inline_form`. The previous `navigate` kind disappears.
- [prototype/](./prototype): the prototype should be deleted or rebuilt against this redesign. Its current shape now actively misleads design review.

## Open Questions To Resolve Before Production Code

1. Should the footer key for advanced/troubleshooting be `a` or `?`? `?` is more discoverable; `a` avoids colliding with help conventions.
2. On the Drift screen, should the safe next step be applied immediately on `enter`, or should StageServe show a one-line preview of exactly what it will change first? The safer-by-default principle says preview; the Ollama-style brevity principle says inline-on-the-row description is enough.
3. Should the inline init form support a "skip and use defaults" return without re-confirmation, or always re-show preview after edits? Preview-after-edit is closer to Fly Launch behavior.

These questions are explicitly deferred until this redesign is accepted in principle. Note: drift is a first-class planner situation defined in the spec, not an implementation detail to hand-wave; the open question above only concerns *how* the Drift screen confirms its safe step, not *whether* drift deserves a dedicated screen.
