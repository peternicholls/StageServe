# Out-Of-Sync Flow

This flow runs when the planner detects the situation it internally calls `drift_detected`. The user-facing name is **"this project doesn't match what StageServe expects"**. The word `drift` never appears on screen.

This is a first-class screen, not an edge case. Keeping the project trustworthy after partial failures, manual edits, or external changes is a major part of why StageServe exists.

## What Triggers This Screen

The planner reports `drift_detected` when at least one of the following is true (the user does not see this list):

- StageServe's records say the project is running, but the URL doesn't respond.
- StageServe's records say the project is running, but no underlying runtime is listed.
- StageServe's records say the project is stopped, but the URL is responding.
- DNS for the project's URL is missing.
- Routing for the project's URL is missing.
- `.env.stageserve` was edited in a way that conflicts with what's currently routed.

The user is told what is observably true ("StageServe expected this to be reachable, but it isn't"), not the internal check name.

## Defaults That Must Be Visible

Before the user does anything, they must see:

- The project name and URL.
- A short, plain-language summary of what doesn't match.
- The safe next step, named in advance.
- What that step will and will not change on disk.

## Top-Level Screen

```
StageServe 0.7.0  pete-site looks out of sync

  https://pete-site.stage.local                          (not responding)

  Here's what StageServe found:
    StageServe thinks this project is running, but
    https://pete-site.stage.local isn't responding.

  Safe next step (the highlighted choice):
    Treat this project as stopped, then let you start it again
    if you want. Nothing in your folder will be deleted.

▶ Use the safe next step
    StageServe will forget the running record. You can run it again next.

  Try to start it again
    StageServe will run this project with its current settings.

  Show what doesn't match in detail
    Read a longer plain-language explanation

  More…
    Show direct commands, plain text output, advanced and troubleshooting

  ↑/↓ navigate • enter use safe next step • → details • esc quit
```

Notes:

- The word "drift" never appears.
- The safe next step is named in the body before it appears in the action list.
- The action list has at most three primary items, all phrased as user goals.
- The default is the lowest-risk option. It does not start, stop, or delete anything destructively.

## Sub-Flow: Show What Doesn't Match In Detail

This is opt-in. A normal user never has to read it.

```
StageServe 0.7.0  What doesn't match for pete-site

  StageServe checks several things. Here's what each one says right now:

    Project record         StageServe thinks this is running.
    Local URL response     https://pete-site.stage.local isn't responding.
    DNS for stage.local    Working.
    Local routing          No route is set up for pete-site.stage.local.

  This usually happens when:
    • The project was stopped from outside StageServe.
    • A file under StageServe's hidden working folder was changed by hand.
    • Your computer was restarted while a project was running.

  press any key to go back
```

Notes:

- Plain language for every check name. No "container", "compose", "gateway", or "routing table".
- The "this usually happens when" list helps the user understand without blaming.
- The screen is purely informational. No actions on it; the user goes back to the choices.

## Sub-Flow: Use The Safe Next Step (Confirm Before Acting)

```
StageServe 0.7.0  Reset pete-site's running record?

  StageServe will:
    Forget that this project is running.
    Leave .env.stageserve in your folder as it is.
    Leave all your project files as they are.

  After this, you can choose "Run this project" to start it again.

  ▶ Yes, do this    No, cancel

  ←/→ choose • enter confirm • esc cancel
```

Notes:

- The screen tells the user exactly what will and will not change before they commit.
- After confirming, StageServe applies the safe reconciliation, the planner re-runs, and the user lands on `project_down` (or `project_ready_to_run` if there is no retained down record).

## Sub-Flow: Try To Start It Again

This is the same as choosing "Run this project" from `project_ready_to_run`. StageServe runs `stage up` with the existing settings. If it succeeds, the user lands on `project_running`. If it fails again, the planner reports either `out_of_sync` again (with new details) or `cannot_decide` (covered in [08-recovery-and-help.md](./08-recovery-and-help.md)).

## State Transitions Out Of This Flow

| Trigger | Goes To |
|---|---|
| `Use the safe next step` confirmed | re-detect → typically `project_down` |
| `Try to start it again` succeeded | `project_running` |
| `Try to start it again` failed | re-detect → `out_of_sync` again or `cannot_decide` |
| `Show what doesn't match in detail` | back to this screen |
| `esc` | exit TUI; the situation is unchanged on disk |

## What This Flow Does Not Do

- It does not say "drift", "reconcile", "diagnose", "doctor", or "state mismatch".
- It does not delete project files, ever.
- It does not silently apply the safe step. The user always confirms.
- It does not require the user to read the detail screen to understand the choices. The first screen has everything they need.
- It does not call the safe step "the cheap option" or "the lazy option". It calls it the safe option.
