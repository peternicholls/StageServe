# Guided TUI Prototype Design

Date: 2026-05-07

## Purpose

The spec 007 prototype is a design reference implementation for StageServe terminal UX. It separates visual UI, interaction flow, language, and assisted workflow design from operational code.

The prototype should help StageServe answer two related questions:

1. How should a user move through an assisted workflow from bare `stage`?
2. How should an individual command such as `stage doctor` move from a passive report into guided assistance when the terminal is interactive?

The prototype remains fixture-only. It must not run Docker, touch DNS, open browsers, write `.env.stageserve`, or change StageServe state.

## Design Direction

Use one shared terminal grammar for both guided screens and command reports, while preserving their different jobs.

Guided interactive screens are for choosing and confirming the next step. They should feel like a calm local-development cockpit:

- a clear StageServe header;
- one human verdict near the top;
- visible facts before action;
- one highlighted default action;
- a small number of secondary actions;
- context-specific footer controls.

Command report screens are for diagnosis and evidence. `stage doctor` is the current example:

- verdict first;
- blockers before passing checks;
- exact remediation commands;
- semantic sections such as `Needs fixing` and `All clear`;
- plain text parity.

The prototype should show how these surfaces connect. A report may offer guided help, but it should not become a noisy menu.

## Report-To-Assistance Handoff

When an interactive command report finds issues, it may end with an assistance invitation:

```text
StageServe can help with the issues above.

▶ Help me fix these
    Walk through each issue one at a time.

  Leave it here
    Exit without changing anything.
```

The wording is intentionally assistance-oriented, not a guarantee that StageServe can repair everything automatically. Some blockers need information, approval, or physical user action.

Rules:

- Show the passive report first so power users can copy commands and leave.
- Offer assistance only when stdin and stdout are interactive.
- `Leave it here` exits without changing anything.
- `Help me fix these` starts a guided flow ordered by least invasive step first.
- Each assisted step explains what StageServe can do, what it cannot do, and whether it needs approval.
- Any privileged read, privileged write, or mutation must get its own confirmation.
- After each assisted step, re-check or return to the report with updated state.

## Assisted Blocker Pattern

Each blocker should become a focused screen, not a list of raw diagnostics.

For a read-only command that needs elevated permission, such as identifying the owner of port 443:

```text
◆ StageServe                         Port 443

Something else on your computer is using port 443.

StageServe can check which process owns the port. Your computer
will ask for your password because macOS hides this detail by default.

▶ Check with sudo
    Run a read-only command to identify the process.

  Skip this issue
    Leave port 443 unresolved for now.

enter check • esc back
```

For a setup step that changes machine configuration, such as local DNS:

```text
◆ StageServe                         Local DNS

Your computer can't yet open local project URLs.

StageServe will write one resolver file so addresses ending in the
active suffix open on this computer.

Will update:
  File            /etc/resolver/develop
  Domain suffix   .develop

▶ Yes, set this up
    Your computer will ask for your password.

  No, skip for now
    Leave local DNS unresolved.

←/→ choose • enter confirm • esc cancel
```

## Interactive Screen Grammar

Guided screens should use this order:

```text
◆ StageServe                         <surface>

<Human verdict sentence.>

Key facts:
  <Label>         <value>             (<short note when useful>)
  <Label>         <value>

▶ <Default action>
    <What pressing enter will do.>

  <Secondary action>
    <Why someone would choose it.>

<context-specific footer>
```

Rules:

- Do not use boxes or decorative borders.
- Use spacing, alignment, section rules, and semantic glyphs for scanability.
- Keep first-level language plain. Internal vocabulary belongs only in advanced/troubleshooting.
- Every default value and default action must be visible before commitment.
- The default action should be the lowest-risk likely goal.
- Running-project screens must never default to stopping the project.
- Text fallback must preserve the same information without relying on colour or glyphs.

## Prototype Component Vocabulary

The prototype should introduce small private render helpers rather than one-off string assembly:

- `screenHeader`
- `verdict`
- `factRows`
- `workChecklist`
- `actionList`
- `confirmation`
- `reportSection`
- `footerHelp`

These helpers are prototype scaffolding. Production can lift only the pieces that prove durable.

## Documentation Updates

Update the terminal design docs where the prototype clarifies a reusable rule:

- Add a guided interactive TUI pattern to the pattern catalog.
- Explain the relationship between report surfaces and assisted interactive flows.
- Document the report-to-assistance handoff.
- Keep `stage doctor` report examples aligned with the evolving language, especially where current output still exposes implementation-heavy phrasing.

## Testing

Prototype tests should protect UX invariants:

- canonical planner situations still render;
- diagnostic actions are not first-level menu noise;
- reports can offer assistance without hiding exact commands;
- assistance steps are one blocker at a time;
- confirmations show what changes and what does not;
- visible defaults remain visible before writes;
- running-project default remains non-destructive;
- text fallback carries the same semantic information as the TUI.

## Out Of Scope

- Real machine changes.
- Real Docker, DNS, browser, or lifecycle calls.
- Production command rewrites.
- New runtime dependencies.
- JSON contract changes.
