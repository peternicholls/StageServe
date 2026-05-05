# Machine Setup Flow

This flow runs when the planner reports `machine_not_ready`. The user does not navigate it as a menu. The tool walks the checklist itself, only stopping on items that need the user.

## Source Of Truth

The checklist items come from `core/onboarding`. The flow does not invent any checks. It runs the existing setup steps and reports them in plain language.

## Step Order

1. Docker Desktop installed
2. Docker Desktop running
3. StageServe's working folder writable
4. Network ports 80 and 443 free
5. Local DNS resolver ready (so `*.stage.local` resolves)
6. Local HTTPS certificates trusted (so the browser doesn't warn)

The order matters: each step depends on the previous one being satisfied.

## Top-Level Screen (Walking The Checklist)

```
StageServe 0.7.0  Setting up your computer

  Docker Desktop                                          ✓ ready
  Docker Desktop is running                               ✓ ready
  StageServe's working folder                             ✓ ready
  Network ports 80 and 443                                ✓ ready
▶ Local DNS resolver                                      needs your approval
    StageServe will add a small file so your browser can open
    URLs that end in .stage.local. This needs your password.
    On enter: StageServe will ask your computer for permission.

  Local HTTPS certificates                                pending

  enter approve • s skip this step • → details • esc quit
```

Notes:

- Each row shows its current state in plain words: `ready`, `pending`, `needs your approval`, `not installed`, or `skipped`.
- The active row (marked `▶`) is the next thing the tool wants to do, with a one-line plain-language description of what enter will do.
- `s skip this step` lets the user move on if a step is acceptable to defer (for example HTTPS trust). Skipped steps remain visible so the user knows what is missing.
- `→ details` opens a longer plain-language explanation for the active row. It is opt-in.

## Sub-Flow: An External Blocker (Cannot Be Fixed By StageServe)

When a step requires the user to do something StageServe can't do itself, the screen replaces the checklist with a focused blocker view. There is no menu of options.

```
StageServe 0.7.0  Setting up your computer

  Docker Desktop is not installed.

    StageServe needs Docker Desktop to run your sites.

    1. Open https://www.docker.com/products/docker-desktop in your browser.
    2. Download Docker Desktop for macOS.
    3. Install it and open it. Wait until the whale icon in your menu bar
       has stopped animating.

    When you've done this, press enter and StageServe will check again.

  enter check again • → why does StageServe need this • esc quit
```

Notes:

- One blocker at a time. The user is never asked to mentally juggle multiple problems.
- Steps are physical actions, numbered, in order.
- The default action (`enter`) is spelled out in the body, not just the footer.
- `→ why does StageServe need this` is progressive disclosure for users who want to understand before acting. The default user does not have to read it.

## Sub-Flow: A Step StageServe Can Do Itself (One-Key Confirm)

Some steps require user permission but can be performed by StageServe directly. They get a yes/no confirmation, with `Yes` highlighted by default.

```
StageServe 0.7.0  Setting up your computer

  Local DNS resolver

    StageServe will write one small file so your browser can open
    URLs that end in .stage.local. This file lives at:

      /etc/resolver/stage.local

    Writing here needs your password. macOS will ask you next.

▶ Yes, set this up    No, skip for now

    ←/→ choose • enter confirm • → details • esc cancel
```

Notes:

- The exact file path and the exact reason for the password prompt are visible before the user commits.
- `Yes` is the default. The user can press enter once and the right thing happens.
- `No, skip for now` does not abort the whole setup. It marks this step skipped and continues to the next step.

## Sub-Flow: When All Steps Are Done

When every step is `ready` or explicitly `skipped`, the setup screen shows a one-line summary and the planner re-runs.

```
StageServe 0.7.0  Setup finished

  Docker Desktop                              ✓ ready
  Docker Desktop is running                   ✓ ready
  StageServe's working folder                 ✓ ready
  Network ports 80 and 443                    ✓ ready
  Local DNS resolver                          ✓ ready
  Local HTTPS certificates                    skipped (you can finish this later)

▶ Continue
    StageServe will look at the current folder next.

  enter continue • esc quit
```

Notes:

- Skipped items are visible, so the user can come back to them.
- The user always sees what StageServe is about to do next. They are not dropped silently into a different screen.

## State Transitions Out Of This Flow

| Trigger | Goes To |
|---|---|
| Every step `ready` or `skipped`, user presses enter on the summary | re-detect → typically `not_project`, `project_missing_config`, or `project_ready_to_run` |
| User presses esc at any point | exit TUI; nothing partial is left in a confused state |
| A step succeeded, more remain | stay in this flow, advance to the next step |
| A step requires the user and they pressed `s` | mark skipped, advance |

## What This Flow Does Not Do

- It does not show Docker, daemon, container, image, or volume names anywhere. The user never sees `dockerd`, `com.docker.docker`, `bridge network`, or anything like it.
- It does not present a "Find issues" button. This screen *is* the find-issues experience, owned by the tool.
- It does not show a multi-tab or tree menu. There is one active row, one blocker at a time.
- It does not require the user to know what to do next. Even on hard blockers, the body text tells them physically what to do.
