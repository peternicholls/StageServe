# The Defaults-Visible Rule

This file states one rule that every other diagram in this folder must follow:

> **Every screen with a default value or a default action must show that value or action inline, before the user has to commit anything.**

The user must never have to drill down to discover what StageServe is going to do for them. If StageServe has a value, that value is on the screen. If `enter` will trigger an action, the screen says what that action is.

Example screens use `.develop`, but the rule is configuration-driven: if the active project uses `.test`, `.dev`, a full hostname, HTTPS, or a non-default port, the screen shows that exact value.

## What "Default" Means In This TUI

Two distinct things, both must always be visible:

1. **Default value** — a setting StageServe will use unless the user changes it (site name, web folder, domain suffix, port, etc.).
2. **Default action** — what happens if the user just presses `enter` without doing anything else (run, stop, confirm, continue, etc.).

## Display Conventions

These conventions are used uniformly across every screen in this folder so the user learns them once.

### Inline Default Values

When a row represents a value StageServe will use, the value appears next to its label:

```
  Site name           pete-site                       (default: folder name)
  Web folder          ./public_html                   (default: found here)
  Domain suffix       .develop                        (default: your machine setting)
  Local URL           http://pete-site.develop        (this is what you'll visit)
```

The right column is always the value. The note in parentheses explains where the default came from, so the user knows why StageServe picked it.

### Inline Default Action Description

When a row represents an action, the row is followed by a one-line plain-language description of what that action will do:

```
▶ Run this project
    Start the project. StageServe will open it in your browser.
```

The user can read what `enter` will do before they press it. The description is the contract for the action.

### Highlighted Default Choice

When the user has to pick between options, the option that is the default is marked with `▶` and is the first item in the list:

```
▶ Use the safe next step
    StageServe will forget the running record. You can run it again next.

  Try to start it again
    StageServe will run this project with its current settings.
```

For yes/no confirmations:

- `Yes` is the default for actions the user explicitly chose to come to (write `.env.stageserve`, run a project, stop a project).
- `No` is the default for destructive actions where the user might have arrived by accident (remove a project from StageServe).

## Rules Per Flow

| Flow | Defaults That Must Be Visible |
|---|---|
| Top-level routing | Not applicable (no screen). |
| Machine setup | Each row's current state, the one active step's description, what `enter` will do, the file path or external action that step will trigger. |
| Project setup | Site name, web folder, domain suffix, URL scheme, port when needed, resulting URL, target file path, what "Use these settings" will write, what "Edit" will open. |
| Run, stop, inspect | Project name, current URL, current status, what the highlighted action will do (the default is non-destructive). |
| Out-of-sync | What StageServe found, what the safe next step is, what it will and will not change on disk. |
| Recovery and help | The ordered list of recovery steps, which step is next, that step is read-only or names what it will change. |
| Power-user paths | The opt-out flags, the file paths for `.env.stageserve` and `.stageserve-state`, the equivalent direct command. |

## Anti-Patterns This Rule Forbids

- **Hidden defaults.** A screen that says "use defaults" without showing the values is wrong. The user must see the values.
- **Mystery enter.** A screen where the user can press enter and something happens, but the screen didn't say what would happen, is wrong.
- **Drilled-down truth.** A screen that requires the user to open a sub-screen to find out what StageServe will do is wrong. The summary must be enough.
- **Inconsistent default highlighting.** A screen where the default option isn't first or isn't marked with `▶` is wrong.
- **Destructive defaults.** A screen where pressing enter accidentally stops, deletes, or detaches anything is wrong. Destructive actions are never the default unless the entire screen exists for that purpose (and even then, they confirm).

## How To Audit A New Screen Against This Rule

For any new screen, the reviewer asks four questions:

1. What value will StageServe use if the user does nothing? Is that value on the screen?
2. What will happen if the user presses enter? Is that explained on the screen, in plain language?
3. If the user has to choose, is the default option the lowest-risk one?
4. If the user wants to change a value, can they do it without leaving the screen they came from for more than one step?

If any answer is "no", the screen is not finished.
