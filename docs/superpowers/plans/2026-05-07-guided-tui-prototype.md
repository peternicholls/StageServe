# Guided TUI Prototype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the spec 007 prototype into a polished design reference for guided StageServe interactions, including the transition from `stage doctor`-style reports into assisted help.

**Architecture:** Keep the prototype fixture-only and private to `specs/007-harden-TUI-and-other-interactions/prototype`. Introduce small composable render helpers inside the prototype before changing screen output. Add one new fixture scenario for report-to-assistance so the UX can be tested without changing production `stage doctor`.

**Tech Stack:** Go, Bubble Tea, ANSI/Lipgloss-compatible terminal styling, existing Makefile prototype targets, Markdown design docs.

---

## File Structure

- Modify: `.github/instructions/terminal-pattern-catalog.instructions.md`
  - Add a reusable pattern for interactive guided screens.
  - Add a reusable pattern for report-to-assistance handoff.
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main.go`
  - Add private component render helpers.
  - Polish existing guided screens to use the shared grammar.
  - Add a `doctor_report_needs_help` fixture scenario and an assisted-help mode.
  - Keep all behavior fixture-only.
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main_test.go`
  - Add UX invariant tests before implementation.
  - Keep existing tests passing.
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/README.md`
  - Document the new scenario and manual review points.
- Create or update if needed: `.omx/notepad.md`
  - Record code-side concerns discovered during design without mixing them into design decisions.

## Task 1: Document The New Terminal Patterns

**Files:**
- Modify: `.github/instructions/terminal-pattern-catalog.instructions.md`

- [ ] **Step 1: Add the guided interactive screen pattern**

Append this section under `## Anticipated Patterns`, before `### Multiple Valid Fix Paths`:

````markdown
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
````

- [ ] **Step 2: Add the report-to-assistance handoff pattern**

Append this section after `### Guided Interactive Screen`:

````markdown
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
````

- [ ] **Step 3: Run a docs scan**

Run:

```bash
rg -n "Help me fix these|Guided Interactive Screen|Report To Assisted Help" .github/instructions/terminal-pattern-catalog.instructions.md
```

Expected: all three phrases are found.

- [ ] **Step 4: Commit**

```bash
git add .github/instructions/terminal-pattern-catalog.instructions.md
git commit -m "Document guided terminal assistance patterns" -m "The prototype is now exploring how passive reports hand off into assisted interactive flows, so the pattern catalog needs examples before code follows them.

Constraint: Terminal design docs are the source of truth for prototype UX
Confidence: high
Scope-risk: narrow
Tested: rg pattern scan
Not-tested: No renderer output changed"
```

## Task 2: Add Failing UX Tests For Assisted Reports

**Files:**
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main_test.go`

- [ ] **Step 1: Add canonical scenario coverage**

Add `doctorReportNeedsHelp` to the `required` list in `TestFixturesContainCanonicalPlannerSituations` after `unknownError`:

```go
doctorReportNeedsHelp,
```

Expected compile failure before implementation:

```text
undefined: doctorReportNeedsHelp
```

- [ ] **Step 2: Add the assisted report text fallback test**

Append this test:

```go
func TestDoctorReportOffersAssistanceWithoutHidingCommands(t *testing.T) {
	plans := planFixtures()
	plan, ok := plans[doctorReportNeedsHelp]
	if !ok {
		t.Fatal("missing doctor report assistance scenario")
	}

	var b strings.Builder
	renderText(&b, plan)
	text := b.String()

	for _, want := range []string{
		"StageServe Doctor",
		"Not ready - 2 of 7 checks need attention.",
		"Needs fixing",
		"Port 443",
		"To fix: sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
		"Local DNS resolver",
		"To fix: stage setup",
		"Assistance",
		"Help me fix these",
		"Walk through each issue one at a time.",
		"Leave it here",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("doctor assistance fallback missing %q:\n%s", want, text)
		}
	}
}
```

- [ ] **Step 3: Add the one-blocker-at-a-time invariant**

Append this test:

```go
func TestAssistedDoctorFlowStartsWithOneFocusedBlocker(t *testing.T) {
	m := newModel(planFixtures(), doctorReportNeedsHelp)
	next := m.handleDecision(m.currentPlan().Decisions[0])

	if next.mode != modeAssist {
		t.Fatalf("mode=%v want modeAssist", next.mode)
	}
	view := next.View()
	for _, want := range []string{
		"Port 443",
		"Something else on your computer is using port 443.",
		"Check with sudo",
		"Run a read-only command to identify the process.",
		"Skip this issue",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("assist view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Local DNS resolver") {
		t.Fatalf("assist view should focus on one blocker at a time:\n%s", view)
	}
}
```

Expected compile failures before implementation:

```text
undefined: modeAssist
```

- [ ] **Step 4: Run tests to confirm failure**

Run:

```bash
go test ./specs/007-harden-TUI-and-other-interactions/prototype
```

Expected: FAIL with missing `doctorReportNeedsHelp` and `modeAssist`.

## Task 3: Add Prototype Render Components

**Files:**
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main.go`

- [ ] **Step 1: Add report section data types**

Insert after `type workItem struct`:

```go
type reportItem struct {
	Label       string
	Description string
	Message     string
	Command     string
	Ready       bool
	Status      workStatus
}

type actionItem struct {
	Label       string
	Description string
}
```

- [ ] **Step 2: Extend `plan` for report sections**

Add fields to `type plan struct` after `WorkItems []workItem`:

```go
	ReportAttention []reportItem
	ReportReady     []reportItem
	AssistanceTitle string
```

- [ ] **Step 3: Add semantic styling helpers**

Replace the existing `bold` and `dim` helpers at the bottom with:

```go
func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	return "\033[90m" + s + "\033[0m"
}

func cyan(s string) string {
	return "\033[36m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}
```

- [ ] **Step 4: Add component render helpers**

Insert before `func (m model) renderMain() string`:

```go
func renderScreenHeader(b *strings.Builder, title, surface string) {
	if surface == "" {
		fmt.Fprintf(b, "%s  %s\n", cyan("◆"), bold(title))
		return
	}
	fmt.Fprintf(b, "%s  %-32s %s\n", cyan("◆"), bold(title), dim(surface))
}

func renderVerdict(b *strings.Builder, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	fmt.Fprintf(b, "\n%s\n", text)
}

func renderFactRows(b *strings.Builder, title string, defaults []defaultValue) {
	if len(defaults) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s\n", bold(title))
	for _, item := range defaults {
		note := ""
		if item.Note != "" {
			note = "  " + dim("("+item.Note+")")
		}
		fmt.Fprintf(b, "  %-16s %-34s%s\n", item.Label, item.Value, note)
	}
}

func renderActionList(b *strings.Builder, decisions []decisionItem, cursor int) {
	if len(decisions) == 0 {
		return
	}
	fmt.Fprintln(b)
	for i, item := range decisions {
		prefix := " "
		if i == cursor {
			prefix = yellow("▶")
		}
		fmt.Fprintf(b, "%s %s\n", prefix, item.Label)
		if item.Description != "" {
			fmt.Fprintf(b, "    %s\n", item.Description)
		}
	}
}

func renderFooterHelp(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	fmt.Fprintf(b, "\n%s\n", dim(text))
}

func renderReportSections(b *strings.Builder, attention, ready []reportItem) {
	if len(attention) > 0 {
		fmt.Fprintf(b, "\n%s\n", bold("Needs fixing"))
		for i, item := range attention {
			fmt.Fprintf(b, "\n  %s  %s\n", yellow(fmt.Sprintf("%d", i+1)), bold(item.Label))
			if item.Description != "" {
				fmt.Fprintf(b, "     %s\n", item.Description)
			}
			if item.Message != "" {
				fmt.Fprintf(b, "\n     %s\n", dim(item.Message))
			}
			if item.Command != "" {
				fmt.Fprintf(b, "     %s %s\n", bold("To fix:"), cyan(item.Command))
			}
		}
	}
	if len(ready) > 0 {
		fmt.Fprintf(b, "\n%s\n", bold("All clear"))
		for _, item := range ready {
			status := string(item.Status)
			if status == "" {
				status = item.Message
			}
			fmt.Fprintf(b, "  %s  %-18s %s\n", green("✓"), item.Label, dim(status))
		}
	}
}
```

- [ ] **Step 5: Run gofmt and tests**

Run:

```bash
gofmt -w specs/007-harden-TUI-and-other-interactions/prototype/main.go specs/007-harden-TUI-and-other-interactions/prototype/main_test.go
go test ./specs/007-harden-TUI-and-other-interactions/prototype
```

Expected: tests still fail only for missing `doctorReportNeedsHelp` and `modeAssist` until Task 4.

## Task 4: Add Doctor Assistance Scenario And Focused Assist View

**Files:**
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main.go`
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main_test.go`

- [ ] **Step 1: Add the scenario constant**

Add after `unknownError situation = "unknown_error"`:

```go
	doctorReportNeedsHelp situation = "doctor_report_needs_help"
```

- [ ] **Step 2: Add the view mode**

Add after `modeEdit`:

```go
	modeAssist
```

- [ ] **Step 3: Add active assistance state to `model`**

Add after `lastScreenName string`:

```go
	assistIndex int
```

- [ ] **Step 4: Route assist mode in `Update` and `View`**

Add to the `Update` mode switch after `case modeEdit`:

```go
	case modeAssist:
		return m.updateAssist(key)
```

Add to the `View` switch after `case modeEdit`:

```go
	case modeAssist:
		return m.renderAssist()
```

- [ ] **Step 5: Add `updateAssist`**

Insert after `updateEdit`:

```go
func (m model) updateAssist(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeMain
	case "enter":
		m.resultTitle = "Read-only check selected"
		m.resultBody = "Prototype only: StageServe would ask before running sudo lsof to identify the process using port 443."
		m.mode = modeMain
	case "s":
		m.resultTitle = "Skipped Port 443"
		m.resultBody = "StageServe left this issue unresolved and would continue to the next blocker."
		m.mode = modeMain
	}
	return m, nil
}
```

- [ ] **Step 6: Extend `handleDecision` for assistance**

At the start of `handleDecision`, before `if item.Opens != modeMain`, add:

```go
	if item.Opens == modeAssist {
		m.mode = modeAssist
		m.assistIndex = 0
		return m
	}
```

- [ ] **Step 7: Add the fixture plan**

Add this entry to `planFixtures()`:

```go
		doctorReportNeedsHelp: {
			Situation:    doctorReportNeedsHelp,
			StatusHeader: "StageServe Doctor",
			Context:      "Read-only machine check",
			Summary:      "Not ready - 2 of 7 checks need attention.",
			ReportAttention: []reportItem{
				{
					Label:       "Port 443",
					Description: "Port 443 must be free for the local HTTPS gateway to bind to it.",
					Message:     "Owner requires sudo to identify.",
					Command:     "sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
				},
				{
					Label:       "Local DNS resolver",
					Description: "Your computer cannot yet open local project URLs.",
					Message:     "dnsmasq config missing",
					Command:     "stage setup",
				},
			},
			ReportReady: []reportItem{
				{Label: "Docker CLI", Status: statusReady, Message: "docker found at /usr/local/bin/docker"},
				{Label: "Docker Desktop", Status: statusReady, Message: "running"},
				{Label: "State directory", Status: statusReady, Message: "exists"},
				{Label: "Port 80", Status: statusReady, Message: "available"},
				{Label: "mkcert local CA", Status: statusReady, Message: "installed"},
			},
			AssistanceTitle: "Assistance",
			Decisions: []decisionItem{
				{ID: "assist", Label: "Help me fix these", Description: "Walk through each issue one at a time.", Opens: modeAssist},
				{ID: "leave", Label: "Leave it here", Description: "Exit without changing anything.", ResultTitle: "No changes made", ResultBody: "Prototype only: StageServe would exit without changing anything."},
			},
			DetailsTitle:   "Doctor report overview",
			Details:        []string{"StageServe shows the report first so commands remain copy-pasteable.", "Guided help starts only when the user asks for it."},
			DirectCommands: []string{"sudo lsof -nP -iTCP:443 -sTCP:LISTEN", "stage setup", "stage doctor"},
			Advanced:       []string{"Advanced view would include internal check IDs and raw command output."},
			Footer:         baseFooter,
		},
```

- [ ] **Step 8: Render report sections and assistance actions in `renderMain`**

In `renderMain`, after the summary block and before defaults, add:

```go
	renderReportSections(&b, p.ReportAttention, p.ReportReady)
	if p.AssistanceTitle != "" {
		fmt.Fprintf(&b, "\n%s\n", bold(p.AssistanceTitle))
	}
```

- [ ] **Step 9: Render report sections and assistance actions in `renderText`**

In `renderText`, after the summary block and before the defaults block, add:

```go
	if len(p.ReportAttention) > 0 || len(p.ReportReady) > 0 {
		fmt.Fprintf(w, "\nNeeds fixing\n")
		for i, item := range p.ReportAttention {
			fmt.Fprintf(w, "\n%d. %s\n", i+1, item.Label)
			if item.Description != "" {
				fmt.Fprintf(w, "   %s\n", item.Description)
			}
			if item.Message != "" {
				fmt.Fprintf(w, "   %s\n", item.Message)
			}
			if item.Command != "" {
				fmt.Fprintf(w, "   To fix: %s\n", item.Command)
			}
		}
		if len(p.ReportReady) > 0 {
			fmt.Fprintf(w, "\nAll clear\n")
			for _, item := range p.ReportReady {
				status := item.Message
				if status == "" {
					status = string(item.Status)
				}
				fmt.Fprintf(w, "- %s: %s\n", item.Label, status)
			}
		}
	}
	if p.AssistanceTitle != "" {
		fmt.Fprintf(w, "\n%s\n", p.AssistanceTitle)
	}
```

- [ ] **Step 10: Add `renderAssist`**

Insert after `renderEdit`:

```go
func (m model) renderAssist() string {
	var b strings.Builder
	renderScreenHeader(&b, "StageServe", "Port 443")
	renderVerdict(&b, "Something else on your computer is using port 443.")
	fmt.Fprintf(&b, "\nStageServe can check which process owns the port. Your computer\n")
	fmt.Fprintf(&b, "will ask for your password because macOS hides this detail by default.\n")
	renderActionList(&b, []decisionItem{
		{Label: "Check with sudo", Description: "Run a read-only command to identify the process."},
		{Label: "Skip this issue", Description: "Leave port 443 unresolved for now."},
	}, 0)
	renderFooterHelp(&b, "enter check • s skip • esc back • q quit")
	return b.String()
}
```

- [ ] **Step 11: Run the focused tests**

Run:

```bash
gofmt -w specs/007-harden-TUI-and-other-interactions/prototype/main.go specs/007-harden-TUI-and-other-interactions/prototype/main_test.go
go test ./specs/007-harden-TUI-and-other-interactions/prototype
```

Expected: PASS.

- [ ] **Step 12: Commit**

```bash
git add specs/007-harden-TUI-and-other-interactions/prototype/main.go specs/007-harden-TUI-and-other-interactions/prototype/main_test.go
git commit -m "Prototype assisted help from doctor reports" -m "The design prototype now demonstrates how a passive doctor-style report can hand off into focused guided assistance without changing production command behavior.

Constraint: Prototype remains fixture-only
Rejected: Change production stage doctor first | the prototype is the design surface for validating interaction semantics
Confidence: medium
Scope-risk: moderate
Tested: go test ./specs/007-harden-TUI-and-other-interactions/prototype
Not-tested: Manual TTY review"
```

## Task 5: Polish Existing Prototype Screens With Shared Grammar

**Files:**
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main.go`
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/main_test.go`

- [ ] **Step 1: Rewrite `renderMain` around helper order**

Replace the body of `renderMain` with:

```go
func (m model) renderMain() string {
	p := m.currentPlan()
	var b strings.Builder
	renderScreenHeader(&b, "StageServe", p.StatusHeader)
	if p.Context != "" {
		fmt.Fprintf(&b, "%s\n", dim(p.Context))
	}
	renderVerdict(&b, p.Summary)
	renderReportSections(&b, p.ReportAttention, p.ReportReady)
	if p.AssistanceTitle != "" {
		fmt.Fprintf(&b, "\n%s\n", bold(p.AssistanceTitle))
	}
	renderFactRows(&b, "Key facts", p.Defaults)
	renderWorkPanel(&b, p)
	renderActionList(&b, p.Decisions, m.cursor)
	if m.resultTitle != "" {
		fmt.Fprintf(&b, "\n%s\n%s\n", bold("Latest outcome"), m.resultTitle)
		if m.resultBody != "" {
			fmt.Fprintf(&b, "%s\n", m.resultBody)
		}
	}
	renderFooterHelp(&b, "↑/↓ navigate • enter use highlighted • ? details • m more • a advanced • tab next scenario • q quit")
	return b.String()
}
```

- [ ] **Step 2: Keep `renderDefaults` unused only if no references remain**

Run:

```bash
rg -n "renderDefaults|renderDecisionBar" specs/007-harden-TUI-and-other-interactions/prototype/main.go
```

Expected after Step 1: only function definitions remain.

If only definitions remain, delete both `renderDefaults` and `renderDecisionBar`.

- [ ] **Step 3: Update tests for renamed surface heading**

If `TestTextFallbackUsesSurfaceLanguage` still passes, leave it unchanged. If it fails because the heading changed from `Visible defaults` to `Key facts`, change this expected string:

```go
"Visible defaults",
```

to:

```go
"Key facts",
```

- [ ] **Step 4: Run tests**

Run:

```bash
gofmt -w specs/007-harden-TUI-and-other-interactions/prototype/main.go specs/007-harden-TUI-and-other-interactions/prototype/main_test.go
go test ./specs/007-harden-TUI-and-other-interactions/prototype
```

Expected: PASS.

## Task 6: Update Prototype README And Manual Review Checklist

**Files:**
- Modify: `specs/007-harden-TUI-and-other-interactions/prototype/README.md`

- [ ] **Step 1: Add the new scenario to run examples**

Add this line to the `Run` code block:

```bash
go run ./specs/007-harden-TUI-and-other-interactions/prototype --scenario doctor_report_needs_help
```

- [ ] **Step 2: Add the new demonstration bullet**

Under `## What It Demonstrates`, add:

```markdown
- Doctor-style reports can offer guided help without hiding copy-pasteable commands.
```

- [ ] **Step 3: Add manual TTY checks**

Under `Manual TTY checks`, add:

```markdown
- Start at `doctor_report_needs_help`; confirm the passive report appears before assistance choices.
- Choose `Help me fix these`; confirm the next screen focuses only on Port 443 and explains the read-only sudo check.
- Confirm `Leave it here` exits the assistance path without implying a change was made.
```

- [ ] **Step 4: Run README scan**

Run:

```bash
rg -n "doctor_report_needs_help|guided help|Port 443" specs/007-harden-TUI-and-other-interactions/prototype/README.md
```

Expected: all terms are found.

- [ ] **Step 5: Commit**

```bash
git add specs/007-harden-TUI-and-other-interactions/prototype/README.md
git commit -m "Document assisted doctor prototype scenario" -m "The prototype README now tells reviewers how to exercise the report-to-assistance flow and what UX invariants to inspect manually.

Constraint: Prototype review is visual and fixture-only
Confidence: high
Scope-risk: narrow
Tested: rg README scan
Not-tested: Manual TTY review not run in this task"
```

## Task 7: Final Verification And Notes

**Files:**
- Modify if needed: `.omx/notepad.md`

- [ ] **Step 1: Run prototype tests**

Run:

```bash
make prototype-test
```

Expected:

```text
go test ./specs/007-harden-TUI-and-other-interactions/prototype
ok  	github.com/peternicholls/stageserve/specs/007-harden-TUI-and-other-interactions/prototype
```

- [ ] **Step 2: Run key text fallbacks**

Run:

```bash
make prototype-text PROTOTYPE_SCENARIO=doctor_report_needs_help
make prototype-text PROTOTYPE_SCENARIO=project_missing_config
make prototype-text PROTOTYPE_SCENARIO=project_running
make prototype-text PROTOTYPE_SCENARIO=drift_detected
```

Expected:

- `doctor_report_needs_help` includes `Needs fixing`, exact `To fix:` commands, and `Help me fix these`.
- `project_missing_config` still shows proposed settings before write.
- `project_running` still defaults to logs, not stop.
- `drift_detected` still uses plain-language safe recovery wording.

- [ ] **Step 3: Run scenario list**

Run:

```bash
make prototype-list
```

Expected: includes `doctor_report_needs_help`.

- [ ] **Step 4: Record deferred code-side issues**

If implementation reveals production concerns, append them to `.omx/notepad.md` under a `Code-side issues for later` heading. If no new code-side issue is discovered, leave the notepad unchanged. For the currently known passive-doctor concern, use this exact entry if it is not already recorded:

```markdown
## Code-side issues for later

- 2026-05-07: Production `stage doctor` is currently a passive report and does not offer an interactive assistance handoff. Context: `cmd/stage/commands/doctor.go` and `core/onboarding/projection_tui.go`. Do not solve during the prototype design pass.
```

- [ ] **Step 5: Commit verification notes if `.omx/notepad.md` is tracked**

Run:

```bash
git status --short -- .omx/notepad.md
```

If Git shows `.omx/notepad.md`, commit it:

```bash
git add .omx/notepad.md
git commit -m "Record guided TUI prototype follow-up notes" -m "The design pass separates production code concerns from prototype UX decisions so later implementation work can discuss them without interrupting visual design.

Constraint: Code-side concerns should not derail prototype design
Confidence: medium
Scope-risk: narrow
Tested: git status --short -- .omx/notepad.md
Not-tested: No runtime behavior changed"
```

If Git shows nothing, do not commit `.omx/notepad.md`.

- [ ] **Step 6: Final status**

Run:

```bash
git status --short
```

Expected: clean or only intentionally ignored `.omx` local state.
