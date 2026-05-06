---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe TUI & Output Style Guide

This file governs all terminal output produced by StageServe.
Apply it whenever touching `projection_tui.go`, `projection_text.go`, `projection_shared.go`, or any command that renders a `CommandResult`.

---

## Guiding principles

1. **Clarity first.** Every line must earn its place. Never print a status tag, summary counter, or internal state string that a human wouldn't naturally say aloud.
2. **Hierarchy through whitespace and weight.** Use indentation, blank lines, and font weight — not borders, boxes, or repeated dividers — to establish structure.
3. **Colour carries meaning, not decoration.** Each colour has a single consistent semantic. Do not add colour for visual interest.
4. **Commands must be copy-pasteable.** Any command shown in output must be the exact shell invocation, free of surrounding prose.
5. **The TUI path and the plain-text path must produce the same information.** The TUI path may add colour and weight; it must not add or remove content.

---

## Colour palette (lipgloss ANSI codes)

| Purpose | ANSI | Style var | When to use |
|---|---|---|---|
| Success / ready | `"2"` (green) | `styleReady` | ✓ icons, "All clear" section header, "ready" verdict, footer command when ready |
| Warning / needs action | `"3"` (yellow) | `styleNeedsAction` | Issue numbers, "Needs fixing" section header, `!` icon |
| Error | `"1"` (red) | `styleError` | ✗ icon and verdict text when any check is in error state |
| Primary action / command | `"14"` (bright cyan, bold) | `styleBrightCyan` | The actual command in "To fix:" lines; footer "next command" when not ready |
| Supporting command / link | `"6"` (cyan) | `styleCyan` | Header `◆` icon; compact-mode next-step arrow `▸`; inline command refs in prose |
| Structural accent | `"15"` (bright white, bold) | `styleWhite` | Page title; check labels in the issues section |
| Supplementary text | `"7"` (light grey) | `styleMuted` | One-sentence descriptions (italic); secondary prose |
| De-emphasised text | `"8"` (dark grey) | `styleDim` | Divider lines; problem messages; compact status words in the passed table |
| Neutral bold | default bold | `styleBold` | "To fix:" label; section labels in compact mode; column headers |

Rules:
- **Never** use colour for decorative purposes only.
- **Never** introduce a new named colour without updating this table.
- Issue numbers adopt warning or error colour to match their step's severity — not a fixed yellow.
- The verdict text (not just its icon) is styled in the same colour as the icon.

---

## Layout structure — detailed mode

```
<blank line>
  ◆  <Title>                                      ← styleCyan icon, styleWhite bold title
  ──────────────────────────────────────           ← styleDim divider
<blank line>
  ✗  Not ready — N of M checks need attention.    ← styleError bold (or styleReady bold when all pass)
<blank line>
── Needs fixing ────────────────────────           ← tuiSectionHeader("Needs fixing", "3")
<blank line>
  N  <Label>                                      ← issue-colour bold number, styleWhite bold label
     <one-sentence description>                   ← styleMuted italic
<blank line>
     <problem message>                            ← styleDim
     To fix:  <command>                           ← styleBold "To fix:", styleBrightCyan command
<blank line>  (between issues)
<blank line>  (after last issue)
── All clear ───────────────────────────          ← tuiSectionHeader("All clear", "2")
<blank line>
  ✓  <Label padded to 18>  <compact status>      ← styleReady ✓, styleBold label, styleDim status
<blank line>
  ──────────────────────────────────────           ← styleDim divider
  Fix the issues above, then run: <command>       ← styleBrightCyan command (or styleReady when ready)
<blank line>
```

Section title variants:
- Issues present: `"Needs fixing"` (yellow `"3"`) + `"All clear"` (green `"2"`)
- All passing: `"Checks passed"` (green `"2"`) — no "Needs fixing" section

---

## Layout structure — compact mode

Used by `stage init` and inline readiness checks. Does **not** show descriptions or a footer.

```
✓  <Label>                                        ← only icon + label for passing checks
!  <Label>                                        ← icon + label + message + "Next:" for failing ones
   <problem message>                              ← styleDim
   Next:  <command>                               ← styleBold "Next:", styleCyan command

▸  <next step>                                    ← styleCyan arrow + styleBold text
```

Rules:
- Only print message/remediation rows for checks that are not `StatusReady`.
- At most one "next step" line at the bottom.

---

## Section headers

`tuiSectionHeader(title, colorCode)` is the canonical function.

Structure: `dim "── " + coloured-bold title + dim " ─────..."` totalling ~40 visible columns.

- Always use `tuiSectionHeader` in TUI path; `sectionHeader` (plain) in text path.
- Do not create freeform section headers inline in render functions.

---

## Column alignment in tables

The passed-checks table must align. Apply `fmt.Sprintf("%-18s", label)` to the **raw string** before passing it to `styleBold.Render(...)`. ANSI escape codes inflate byte counts and break `%-Ns` format verbs.

```go
paddedLabel := fmt.Sprintf("%-18s", s.Label)
fmt.Fprintf(w, "  %s  %s  %s\n", styleReady.Render("✓"), styleBold.Render(paddedLabel), styleDim.Render(compactMessage(s)))
```

If the column width changes, update it here and in the code together.

---

## Compact status words (passed table)

`compactMessage(s StepResult)` in `projection_shared.go` controls the short status string for each check ID. Current values:

| Check ID | Compact word |
|---|---|
| `state.dir` | `exists` |
| `port.80`, `port.443` | `available` |
| `mkcert.binary` | `installed` |
| `dns.resolver` (ready) | `configured` |
| anything else | `s.Message` verbatim |

When adding a new check, add its compact word here.

---

## Check descriptions

`checkDescription(id string)` in `projection_shared.go` maps a step ID to one italic sentence explaining *why* the check matters. This line appears beneath the check label in the issues section only.

Format: plain statement of fact ending with a full stop. No "Note:", no imperative, no "you must".

Current descriptions:

| ID | Description |
|---|---|
| `docker.binary` | `Docker CLI — the command-line tool used to manage containers.` |
| `docker.daemon` | `The Docker daemon must be running before any container can start.` |
| `state.dir` | `Stores StageServe runtime data: ports, certs, project registry.` |
| `port.80` | `Port 80 must be free for the local HTTP gateway to bind to it.` |
| `port.443` | `Port 443 must be free for the local HTTPS gateway to bind to it.` |
| `dns.resolver` | `Routes *.test domains to your stack — needs dnsmasq configured.` |
| `mkcert.binary` | `Creates trusted local HTTPS certificates without browser warnings.` |

When adding a new check, add its description here and in `checkDescription`.

---

## Remediation strings

Remediation values stored on `StepResult.Remediation` must be:
- The exact shell command only — no `"Run: "` prefix, no trailing explanation in parentheses.
- `cleanRemediation()` strips legacy `"Run: "` / `"run: "` prefixes automatically, but new code should not add them.
- The rendered label is always `"To fix:"` (detailed mode) or `"Next:"` (compact mode).

---

## Things never to print

- Internal status strings: `needs_action`, `StatusNeedsAction`, `OverallReady`, exit codes in prose.
- Summary counters: `"Summary: N ready, N need attention, N errors"`.
- `(s)` pluralisation: use `plural(n, sing, plur string)` from `projection_shared.go`.
- Step ID tags like `[docker.binary]` or `[Needs action]`.
- `"Result: Needs action (exit 1)"` or any machine-facing status field.
- Box borders (`╭─╮`, `│`, `╰─╯`).

---

## Updating this guide

When you change the output layout, colour use, or shared helper behaviour:
1. Update the relevant section of this file in the same commit.
2. Update the table in "Check descriptions" or "Compact status words" if you add a new check.
3. If you add a new named style var, add a row to the colour palette table.
4. Run `go test ./core/onboarding ./cmd/stage/commands` to confirm tests still pass.
