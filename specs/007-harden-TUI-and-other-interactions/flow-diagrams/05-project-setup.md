# Project Setup Flow

This flow runs when the planner reports `project_missing_config` (the user is in a folder StageServe could treat as a project, but `.env.stageserve` isn't there yet) or `not_project` (the user explicitly chose "Set up this folder as a project" from the not-a-project screen).

It is the only place StageServe writes user-editable config. The output is one file: `.env.stageserve` in the project root. Nothing else is written, and nothing is changed on disk before the user confirms.

## Source Of Truth

The fields, defaults, and validation come from `core/onboarding` `project_env.go` (the existing `stage init` engine). This flow does not invent any fields. It presents the existing engine's defaults to the user and lets them accept or change them before the file is written.

## Defaults That Must Be Visible

The user must see these values on the first screen, before they decide:

| Field | Default Source | What The User Sees |
|---|---|---|
| Site name | folder name, lowercased, hyphenated | `pete-site` |
| Web folder | `./public_html` if it exists, else `./` | `./public_html` |
| Domain suffix | machine-wide setting, default `.stage.local` | `.stage.local` |
| Resulting URL | combined preview | `https://pete-site.stage.local` |
| Target file | always `<project>/.env.stageserve` | `/Users/pete/sites/pete-site/.env.stageserve` |

If any of these is missing or invalid (for example no `./public_html` folder), the screen shows what was found and how StageServe filled the gap.

## Top-Level Screen (Preview)

```
StageServe 0.7.0  /Users/pete/sites/pete-site

  This folder doesn't have StageServe settings yet.
  StageServe will create one file: .env.stageserve

  Site name           pete-site                       (default: folder name)
  Web folder          ./public_html                   (default: found here)
  Domain suffix       .stage.local                    (default: your machine setting)
  Local URL           https://pete-site.stage.local   (this is what you'll visit)

▶ Use these settings
    Write .env.stageserve and continue

  Edit before writing
    Change site name, web folder, or domain suffix first

  ↑/↓ navigate • enter confirm • → edit one field • esc cancel
```

Notes:

- Every value is visible before the user commits.
- `Local URL` is shown as a preview so the user knows what they will end up typing into a browser.
- `→ edit one field` jumps straight to editing whatever row the cursor is on, which is the fastest path for a user who only wants to change one thing.
- `esc` here writes nothing.

## Sub-Flow: Edit Before Writing (Inline Form)

When the user picks "Edit before writing", or presses `→` on a specific row, they get an inline form. The form is keyboard-only. Each field shows its current value as the placeholder, so the user can hit enter to keep it.

```
StageServe 0.7.0  Edit project settings

  Site name
    pete-site
    Used for the local URL. Lowercase, hyphens, no spaces.

  Web folder
    ./public_html
    The folder StageServe should serve. Type a path relative to this project.

  Domain suffix
    .stage.local
    Used to build the local URL. Most people leave this as is.

  Local URL preview
    https://pete-site.stage.local

▶ Save and preview
    Go back to the confirmation screen with these values

  Cancel
    Discard changes

  ↑/↓ navigate • tab next field • enter save • esc cancel
```

Notes:

- The user can edit one field, all of them, or none. Pressing enter on a row keeps its current value.
- The URL preview updates live as the user types.
- Saving does not write the file. It returns to the preview screen with the new values, where the user still has to confirm.

## Sub-Flow: Confirmation (After Edit Or Direct Accept)

If the user picked "Use these settings" on the preview, or "Save and preview" then "Use these settings" after editing, StageServe shows a one-screen confirmation before writing.

```
StageServe 0.7.0  About to write project settings

  StageServe will create:
    /Users/pete/sites/pete-site/.env.stageserve

  with these settings:
    Site name      pete-site
    Web folder     ./public_html
    Domain suffix  .stage.local
    Local URL      https://pete-site.stage.local

  StageServe will not change any other file in this folder.

▶ Yes, create it    No, go back

  ←/→ choose • enter confirm • esc cancel
```

Notes:

- The exact path that will be written is shown. There is no ambiguity about where the file goes.
- The reassurance "StageServe will not change any other file in this folder" is part of the screen, not buried in docs.
- `Yes` is highlighted by default. The user can press enter twice from the very first preview to commit (preview → confirm → enter).

## Sub-Flow: After Writing

After the file is written, the planner re-runs and the user typically lands on `project_ready_to_run`. The previous screen briefly shows a one-line success line and what's next, then transitions automatically when the user presses any key.

```
StageServe 0.7.0  Project settings created

  ✓ Wrote /Users/pete/sites/pete-site/.env.stageserve

  Next: run this project at https://pete-site.stage.local

  press any key to continue
```

If writing failed (permissions, disk full, etc.), the user gets a blocker view modelled on [04-machine-setup.md](./04-machine-setup.md), with the failure described in plain language and a numbered next-step list.

## State Transitions Out Of This Flow

| Trigger | Goes To |
|---|---|
| `Yes, create it` confirmed and write succeeded | re-detect → `project_ready_to_run` |
| Write failed | blocker screen (in this flow), with a numbered fix and a re-try button |
| `esc` from preview, edit form, or confirmation | exit TUI; no file is written |

## What This Flow Does Not Do

- It does not write any file other than `.env.stageserve` in the project root.
- It does not modify project `.env`.
- It does not write or edit anything under `.stageserve-state`.
- It does not propose changes to other projects.
- It does not run `stage up` automatically. The user is dropped on `project_ready_to_run`, where running is the highlighted default but still requires one keypress.
- It does not validate the project against frameworks, package managers, or anything else outside `core/onboarding`.
