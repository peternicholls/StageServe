---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe Terminal Pattern Catalog

This catalog is the example pack for StageServe terminal UX. Use it as precedent when designing, reviewing, or revising CLI output.

Examples are not frozen templates. They show intent, structure, and reasoning. As StageServe evolves, update the catalog so future work follows current product thinking rather than stale output.

Use `.github/instructions/terminal-copy-style.instructions.md` when adapting the wording in these examples. The catalog demonstrates patterns; the copy style guide governs voice, vocabulary, labels, and remediation phrasing.

## Example Format

Each reusable pattern should include:

- **Context:** where the pattern applies.
- **User state:** what the user likely needs at this moment.
- **Output sketch:** representative terminal output, not necessarily byte-perfect.
- **Why it works:** the design reasoning.
- **Rules demonstrated:** links back to the design contract or markup spec concepts.

## Current StageServe Patterns

### Detailed Readiness With Blockers

**Context:** `stage setup` or `stage doctor` finds checks that need attention.

**User state:** The user needs a quick verdict, the exact blocker, and a command they can run.

**Output sketch:**

```text
  ◆  StageServe Doctor
  --------------------------------------

  ✗  Not ready - 2 of 7 checks need attention.

-- Needs fixing ------------------------

  1  Docker daemon
     The Docker daemon must be running before any container can start.

     Docker is installed, but the daemon is not running.
     To fix:  open -a Docker

  2  Port 443
     Port 443 must be free for the local HTTPS gateway to bind to it.

     Another process is already listening on port 443.
     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

-- All clear ---------------------------

  ✓  State directory    exists
  ✓  mkcert             installed

  --------------------------------------
  Fix the issues above, then run: stage setup
```

**Why it works:** The verdict appears before details. Problems come before confirmations. Each issue pairs reason, current problem, and exact remediation.

**Rules demonstrated:** human verdict, problem-first ordering, command copy-pasteability, semantic sections.

### Detailed Readiness All Clear

**Context:** `stage setup` or `stage doctor` confirms everything required is ready.

**User state:** The user needs confirmation and a natural handoff, not a long celebration.

**Output sketch:**

```text
  ◆  StageServe Setup
  --------------------------------------

  ✓  Ready - all checks passed.

-- Checks passed -----------------------

  ✓  State directory    exists
  ✓  Port 80            available
  ✓  Port 443           available
  ✓  mkcert             installed

  --------------------------------------
  Next: stage init
```

**Why it works:** Success is quiet and actionable. The output still provides evidence, but it does not over-explain a healthy state.

**Rules demonstrated:** quieter success state, concise proof, direct next action.

### Compact Inline Check

**Context:** `stage init` performs a readiness check as part of a larger flow.

**User state:** The user is trying to continue and only needs blockers that affect the current action.

**Output sketch:**

```text
✓  State directory
!  DNS resolver
   *.test domains are not resolving to localhost.
   Next:  sudo brew services restart dnsmasq

▸  Fix DNS, then run stage init again.
```

**Why it works:** Compact mode removes descriptions and detailed footers. It keeps the failure close to the next command and prints only one bottom-line next step.

**Rules demonstrated:** compact hierarchy, one next step, failing checks get details, passing checks stay quiet.

## Anticipated Patterns

### First-Run Onboarding

**Context:** The user runs StageServe before local prerequisites or project state exist.

**User state:** The user needs orientation, not a manual.

**Output sketch:**

```text
  ◆  StageServe
  --------------------------------------

  !  Local setup is not ready yet.

-- Start here --------------------------

  1  Install trusted local certificates
     Creates HTTPS certificates without browser warnings.

     Next:  stage setup

  2  Register this project
     Adds the current app to the local StageServe registry.

     Next:  stage init
```

**Why it works:** It frames first run as a short path. It does not expose every underlying check before the user knows the product shape.

**Rules demonstrated:** orientation before detail, ordered path, direct commands.

### Guided Interactive Screen

**Context:** Bare `stage` opens an interactive flow, or a command report hands off into guided help.

**User state:** The user wants StageServe to show the next safe action without needing to remember command names.

**Output sketch:**

```text
  ◆  StageServe                         Project
  --------------------------------------

  This project is ready to run.

-- Key facts ---------------------------

  Local URL       http://pete-site.develop
  Web folder      ./public_html
  Status          not running yet

▶ Run this project
    Start the project and open it in your browser.

  Edit project settings
    Change site name, web folder, or domain suffix first.

  ↑/↓ navigate • enter run • ? details • esc quit
```

**Why it works:** The verdict appears before choices. Default values and the default action are visible before commitment. Secondary choices stay small and user-goal oriented.

**Rules demonstrated:** visible defaults, lowest-risk default action, plain language, semantic hierarchy, context-specific footer.

### Report To Assisted Help

**Context:** `stage doctor` or another command report finds issues in an interactive terminal.

**User state:** The user may understand the report, but may prefer StageServe to walk through the issues one at a time.

**Output sketch:**

```text
  ◆  StageServe Doctor
  --------------------------------------

  ✗  Not ready - 2 of 7 checks need attention.

-- Needs fixing ------------------------

  1  Port 443
     Something else on your computer is using port 443.

     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

  2  Local DNS resolver
     Your computer cannot yet open local project URLs.

     To fix:  stage setup

-- Assistance --------------------------

  StageServe can help with the issues above.

▶ Help me fix these
    Walk through each issue one at a time.

  Leave it here
    Exit without changing anything.
```

**Why it works:** The command remains useful as a report, but interactive users get a safe handoff into guided help. The wording avoids promising that StageServe can fix every blocker automatically.

**Rules demonstrated:** report-first design, progressive disclosure, no hidden mutation, assistance without noisy menus.

### Multiple Valid Fix Paths

**Context:** A blocked check can be fixed in more than one reasonable way.

**User state:** The user needs to choose based on intent.

**Output sketch:**

```text
  !  Port 443 is already in use.

-- Choose a fix ------------------------

  Stop the process if it should not be using HTTPS locally:
  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

  Use a different gateway port for this project:
  stage config set https-port 8443
```

**Why it works:** The choice is explained in human terms before the exact commands. The commands remain copy-pasteable.

**Rules demonstrated:** decision before command, no ambiguous remediation, exact shell invocations.

### Long-Running Operation

**Context:** A command starts work that may take several seconds or more.

**User state:** The user needs to know what is happening and what completion will look like.

**Output sketch:**

```text
  ◆  StageServe Setup
  --------------------------------------

  !  Preparing local HTTPS support.

-- Working -----------------------------

  ✓  mkcert installed
  !  Creating local certificate authority
     This can take a few seconds the first time.
```

**Why it works:** The output explains the current step without streaming noisy implementation logs. It names why waiting is normal.

**Rules demonstrated:** progress as product state, concise expectation setting.

### Destructive Action Warning

**Context:** A command may delete, overwrite, or unregister local state.

**User state:** The user needs the consequence and a clear confirmation path.

**Output sketch:**

```text
  ◆  StageServe
  --------------------------------------

  !  This will unregister api.test from StageServe.

-- Impact ------------------------------

  The project will stop routing through the local gateway.
  Certificates and source files will not be deleted.

  Next:  stage project remove api.test --confirm
```

**Why it works:** It states the concrete impact and what is not affected. Confirmation is explicit in the command.

**Rules demonstrated:** risk clarity, bounded consequence, explicit confirmation.

### Dry-Run Preview

**Context:** A command can show planned changes before applying them.

**User state:** The user needs to compare intent with effect.

**Output sketch:**

```text
  ◆  StageServe Plan
  --------------------------------------

  !  Previewing changes. Nothing has been changed.

-- Will update -------------------------

  Project domain     api.test
  HTTPS port         8443
  Gateway route      /Users/example/api

  Next:  stage apply
```

**Why it works:** The dry-run state is explicit. Planned changes use labels that map to user concepts, not internal keys.

**Rules demonstrated:** state clarity, user-facing labels, safe handoff.

### Empty State

**Context:** A list command finds no registered projects.

**User state:** The user needs to know whether this is normal and how to create the first item.

**Output sketch:**

```text
  ◆  StageServe Projects
  --------------------------------------

  !  No projects are registered yet.

  Next:  stage init
```

**Why it works:** It does not print an empty table. It gives the next action.

**Rules demonstrated:** empty state as UX, no empty scaffolding.

### Post-Success Handoff

**Context:** A command succeeds and there is one natural next workflow step.

**User state:** The user is ready to continue.

**Output sketch:**

```text
  ✓  api.test is registered.

  Next:  stage up
```

**Why it works:** It avoids ceremony. The state is clear and the next command is exact.

**Rules demonstrated:** quiet success, direct continuation.

### Plain-Text Parity

**Context:** `--no-tui`, redirected output, or non-interactive mode.

**User state:** The user or automation needs the same information without styling.

**Output sketch:**

```text
StageServe Doctor

Not ready - 1 of 7 checks needs attention.

Needs fixing

1. Docker daemon
   The Docker daemon must be running before any container can start.

   Docker is installed, but the daemon is not running.
   To fix: open -a Docker

All clear

- State directory: exists
- mkcert: installed

Fix the issues above, then run: stage setup
```

**Why it works:** It preserves information and order without relying on colour, glyphs, or weight.

**Rules demonstrated:** text/TUI parity, semantic content independence.

## Anti-Examples

### Internal State Dump

```text
Result: Needs action (exit 1)
Summary: 5 ready, 2 needs_action, 0 errors
[docker.daemon] StatusNeedsAction
Run: open -a Docker (starts docker)
```

**Why it fails:** It exposes implementation names, duplicates summary math, and pollutes the command with prose.

### Decorative Terminal UI

```text
╭────────────────────────────────────╮
│        STAGESERVE CHECKS!!!        │
╰────────────────────────────────────╯
```

**Why it fails:** It uses decoration instead of hierarchy and makes the terminal feel generic rather than useful.

### Vague Remediation

```text
Something went wrong with DNS.
Try checking your settings and run again.
```

**Why it fails:** It gives no specific state, no likely cause, and no exact next action.

## Catalog Maintenance

Add or revise an example when:

- A new command introduces a reusable interaction shape.
- A changed output pattern should guide future work.
- A review finds repeated ambiguity in how agents or humans interpret the style.
- Product direction changes and old examples no longer represent the desired experience.

Keep examples short enough to scan. Prefer one strong example with reasoning over many near-duplicates.
