# Recovery And Help Flow

This flow runs when the planner reports `cannot_decide` (the internal name is `unknown_error`). It happens when StageServe could not finish gathering context, or when an action failed in a way that doesn't fit any of the other states.

The user-facing name is **"StageServe couldn't safely choose a next step"**. The user is never abandoned: even when the tool can't decide, the screen shows an ordered, named recovery path the user can follow one step at a time.

## Defaults That Must Be Visible

Before the user does anything, they must see:

- A short, plain-language summary of what went wrong, in the user's terms (not the internal error message).
- The ordered list of recovery steps StageServe will run on their behalf.
- Which step is next (highlighted as the default).
- An explicit way to give up safely (no partial state left behind).

## Top-Level Screen

```
StageServe 0.7.0  StageServe couldn't safely choose a next step

  Something went wrong while StageServe was checking on this project.
  StageServe doesn't want to guess. Here's what it can try, in order.

  Recovery steps StageServe will run for you:

    1. Look at this project's current state and report what it finds.
    2. Look at the running log for this project.
    3. Stop and forget the running record (your files won't be touched).
    4. Run this project from scratch.

  Each step is safe on its own. StageServe will pause after each one
  so you can decide whether to continue.

▶ Run step 1: look at this project's current state
    Read-only. Nothing on your computer will be changed.

  Show what went wrong in detail
    Read a longer plain-language explanation

  More…
    Show direct commands, plain text output, advanced and troubleshooting

  ↑/↓ navigate • enter run step 1 • → details • esc quit
```

Notes:

- The recovery list is ordered from least invasive to most invasive. Step 1 is always read-only.
- The first step is the highlighted default. The user can press enter and the safe thing happens.
- After each step, StageServe pauses and re-runs the planner. If the planner can now decide, the user moves to the appropriate normal screen and the recovery flow ends.
- The user can quit at any point. Quitting changes nothing.

## Sub-Flow: After A Recovery Step Runs

```
StageServe 0.7.0  Step 1 finished

  StageServe looked at this project's current state.

  What StageServe saw:
    The project has settings (.env.stageserve found).
    StageServe can't reach the URL right now.
    StageServe doesn't have a running record for this project.

  StageServe thinks the safest next move is:
    Run step 4: run this project from scratch.

▶ Continue with the suggested next step
    Run step 4: run this project from scratch

  Choose a different step
    Pick from the recovery list yourself

  Stop here
    Leave the project as it is and exit StageServe

  ↑/↓ navigate • enter continue • esc quit
```

Notes:

- After every step, StageServe says what it learned, in plain language.
- It then suggests the next step from the recovery list and highlights it as the default.
- The user can override and pick any other step from the recovery list, or stop.
- If the planner now has enough information to choose a normal state, the screen instead transitions to that state. The user lands on `project_ready_to_run`, `project_running`, or `project_down` and the recovery flow ends.

## Sub-Flow: Show What Went Wrong In Detail

This is opt-in. A normal user does not need to read it.

```
StageServe 0.7.0  What went wrong in detail

  StageServe was checking this project and got an error it didn't expect.
  Here is what it tried and what came back, in plain language:

    Tried: read this project's recorded state.
    Result: read it OK.

    Tried: ask the operating system about the project's URL.
    Result: timed out after 5 seconds.

    Tried: ask StageServe's own routing what it thinks.
    Result: didn't get an answer.

  This is the kind of error that usually clears up when you let StageServe
  walk through the recovery steps. If it keeps happening, the advanced
  view (More… → Advanced and troubleshooting) has more detail and the
  exact commands a power user might run by hand.

  press any key to go back
```

Notes:

- Even the "detail" view is in plain language. The internal error name and stack do not appear here. They appear only in the advanced view.
- The detail view never has actions on it. The user reads it and goes back to the recovery list.

## Sub-Flow: Stop Here (Confirm Before Quitting)

```
StageServe 0.7.0  Stop recovery for now?

  StageServe will leave this project as it is.
    Your files won't be touched.
    No change will be made to StageServe's records.
    You can come back to this any time by running stage.

  ▶ Yes, stop here    No, keep going

  ←/→ choose • enter confirm • esc cancel
```

Notes:

- Stopping is always safe. The screen says so explicitly.
- After stopping, StageServe exits. Re-running `stage` re-detects the situation; the user lands back here if nothing else has changed.

## State Transitions Out Of This Flow

| Trigger | Goes To |
|---|---|
| A recovery step fixed the situation | re-detect → normal state, recovery flow ends |
| User chose to continue with the suggested next step | run that step, then back to "after a step" screen |
| User chose a different step | run that step, then back to "after a step" screen |
| `Stop here` confirmed | exit TUI; nothing partial is left |
| `esc` from any screen | same as `Stop here` (with confirmation) |

## What This Flow Does Not Do

- It does not show stack traces, error codes, or internal exception messages on the primary screens.
- It does not require the user to know what a doctor, diagnosis, or check is.
- It does not let StageServe loop on its own. Every step requires a user keypress so the user is always in control.
- It does not delete files, ever.
- It does not skip the safe-by-default order. It always offers the read-only step first.
- It does not give the user the impression that they are stuck. The screen always shows what to do next.
