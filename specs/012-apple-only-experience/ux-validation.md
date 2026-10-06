# Early product study and accessible interaction acceptance

Revised 2026-09-13. Protocol, not completed research. FR-011 and FR-022–023.
TUI-first remains the working design; GUX is required before full M3 implementation.

## Participants and study boundary

Recruit five intended PHP/web developers: at least two who rarely use Terminal,
two comfortable with terminal tools, and one who routinely uses VoiceOver or an
accessible text workflow. Do not replace participants with agents or simulated
personas. Ask for separate permission before recording screens/voices. Use fictional
disposable projects and credentials; retain anonymised observations, not personal
project paths. A facilitator observes without teaching the command sequence.

If participants are unavailable, continue independent runtime and fixture work but
leave GUX open. Do not mark a recruiting plan as usability evidence. Existing source
mockups and runnable terminal style guide can support the study; they are fixtures
and must be labelled as such. No rendered result is assumed from source inspection.

## Early tasks (T039, before T024)

Start at the Mac desktop, with a short install/product welcome and two sample project
folders. Do not start with Terminal already in the right directory.

| Case | Participant instruction | Observe |
|---|---|---|
| UX-01 | Find how to run the sample site locally | Product discovery, Terminal entry and folder navigation without hidden prerequisites |
| UX-02 | Check settings and run site A | Can user identify web folder, URL and consequences, and cancel before writing? |
| UX-03 | Return later, open site B and inspect site A | Project discovery/switching, selected scope and return path |
| UX-04 | Explain a failed start and choose recovery | Distinguish computer setup, infrastructure and application failure; locate safe next step |
| UX-05 | Stop then remove A, explaining what happens to data | Difference between stop/unregister/data deletion; no accidental destructive selection |

Record task start/end, completion without assistance, errors, exact hesitation,
assistance given, participant explanation and severity. Early fixture tasks measure
comprehension/navigation only; do not count mocked site startup as runtime acceptance.
Run a 20–30 minute session per participant and a short post-task explanation.

## Decision thresholds

Working TUI hypothesis passes GUX only if >=4/5 complete discovery and project entry
within five minutes without coaching; all five identify the selected project and
correct data effects; and the accessible participant completes the essential text
journey in a labelled line-oriented prototype. GUX does not depend on T043; G3
repeats the journey using the actual implementation. Zero unresolved critical safety/accessibility issues. A task completed with
facilitator help is recorded as assisted, not a pass. Five participants provide
formative evidence, not population-level statistical confidence.

If failure is copy/discoverability within the terminal, revise onboarding/fixtures
and retest affected tasks with at least three participants including original failing
profiles. If two or more remain blocked by Terminal entry/folder navigation itself,
freeze full dashboard work and compare a minimal native folder-launcher/GUI concept
against improved terminal onboarding before choosing a revised surface. That study
requires a Spec Kit decision amendment and new estimate; it does not authorize
building an entire second interface. Runtime/core work remains useful independently.
Publish decision, evidence and remaining limitations in evidence.md. Revisit before
v1, not only after release, if later trials contradict GUX.

## Accessible mode contract (T043)

Interactive plain-text mode is a complete supported journey. With TTY stdin/stdout,
`stage --cli` or `--notui` uses numbered, line-oriented steps: context/verdict,
settings/effects, choices, exact Enter default, typed response. No alternate-screen
buffer, cursor-up redraw or colour-only status. Invalid input repeats only the
relevant prompt. Enter on confirmations defaults to Cancel. `q` leaves without
stopping running sites. Newline-delimited progress emits only phase changes and a
summary; optional verbosity reveals detail. Redirected input/output remains strictly
noninteractive with a useful plan or direct command, never a hidden prompt. JSON
still wins over interactive/plain output.

The styled TUI must keep focus visible, show modal target/scope, restore focus on
close, avoid unsolicited focus changes on refresh, and pause auto-scrolling logs
when the user scrolls. Offer a discoverable switch to the text journey that retains
selected scope and draft only with an explicit unsaved-changes choice. Never promise
screen-reader support for rapid full-screen redraw without testing it.

| Case | VoiceOver / keyboard / text acceptance |
|---|---|
| A11Y-01 | Discover accessible mode from welcome/help; identify selected project, verdict and next action |
| A11Y-02 | Create and edit settings, read values/errors, cancel and recover focus without writing |
| A11Y-03 | Follow phase changes without announcement flooding; interrupt with Ctrl-C and hear final state |
| A11Y-04 | Inspect/read logs, pause/resume output and exit; no trap or destructive action caused by Enter |
| A11Y-05 | Understand error recovery and exact destructive target; default Cancel, rejection leaves data unchanged |
| A11Y-06 | Complete same run/status/stop/unregister journey in accessible text and styled TUI semantics |

Test Terminal.app with VoiceOver and keyboard-only at 52x24 and 80x24, light/dark,
NO_COLOR and live resize. Text mode must be usable with VoiceOver for all cases.
If styled TUI fails VoiceOver, clearly declare text mode the supported accessible
surface; do not let that excuse missing semantic actions or safety in text mode.
Record actual OS/terminal/VoiceOver versions. No participant or visual evidence is
fabricated. Read-only keyboard snapshots are not substitutes for the full workflow.

## Responsiveness budgets (T026/T028)

Target p95 key-to-feedback <=200 ms over 100 navigation events while an injected
runtime call is delayed 5 s. Show action-start feedback <=200 ms; show phase changes
<=1 s after receiving an event. Ctrl-C acknowledgment <=200 ms and subprocess
cancellation <=2 s; cleanup uses a separate <=15 s budget with visible result.
Do not claim cleanup succeeded if it times out. Log view retains at most 5,000 lines
and 2 MiB (whichever limit first), with explicit truncation; test 10 MiB/min input
for five minutes while scrolling/cancelling. The stream reader cannot block lifecycle
progress or grow memory without bound. Compare text/JSON semantic results.

These are proposed acceptance targets, not measurements. Failures require a fix or
an explicit evidence-backed requirement amendment before G3. No silent relaxation.

## Final trials and evidence

T029 repeats end-to-end discovery through the actual installed candidate with at
least three representative participants, including low-terminal-experience and
accessible-mode coverage. Target first setup/start/visit/stop <=10 min excluding
prerequisite/image downloads; record those exclusions separately. All complete
without undocumented command ordering and no critical unresolved safety/accessibility
blocker. The final distribution gate repeats core tasks on qualified release assets.

Evidence row format: case ID, anonymous profile, candidate/fixture revision,
OS/terminal/mode, expected/observed, elapsed time, assistance, blocker severity,
recording consent if applicable, remediation and retest result. UX and accessibility
cases stay pending until observed. Study preparation does not satisfy GUX or G3.
