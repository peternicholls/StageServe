---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# Goal: StageServe Terminal Experience

Give this goal to Codex or another agent when asking for long-form work on any StageServe CLI interaction. StageServe treats terminal output as a TUI surface even when users only see plain terminal text.

This goal is for humans, Codex, and other agent platforms. It is an operating prompt for doing the work, not the documentation entrypoint and not the full design system.

## Source Documents

Read these in order:

1. `.github/instructions/terminal-experience-index.instructions.md`
2. `.github/instructions/terminal-design-contract.instructions.md`
3. `.github/instructions/terminal-markup-spec.instructions.md`
4. `.github/instructions/terminal-copy-style.instructions.md`
5. `.github/instructions/terminal-pattern-catalog.instructions.md`

Use the design contract to decide what the experience should be. Use the markup spec only after the experience is clear. Use the copy style guide for labels, verdicts, explanations, remediation, and next actions. Use the pattern catalog for precedent and examples.

## Goal

Produce terminal interactions that are clear, consistent, useful, copy-pasteable, and recognisably StageServe.

The output must consider:

- UI design: hierarchy, rhythm, colour semantics, grouping, and scanability.
- UX design: flow, information order, default path, blockers, next actions, and automation boundaries.
- Language: labels, explanations, verdicts, remediation, and clarity.

## Workflow

1. Identify the interaction: command, user journey, current state, target outcome, and failure modes.
2. Decide the task type: design new output, review existing output, or revise an established pattern.
3. Read the design contract before proposing output.
4. Check the copy style guide before writing or revising labels, verdicts, descriptions, warnings, or remediation.
5. Check the pattern catalog for a matching or adjacent precedent.
6. Draft the interaction in plain text first, with the information hierarchy visible without colour.
7. Map the design to the markup spec: colour, weight, glyphs, spacing, helper functions, and text/TUI parity.
8. Implement or recommend changes using existing renderer helpers and local patterns.
9. Add or update catalog examples when a new pattern should guide future work.
10. Verify with tests, snapshots, or targeted command output checks where possible.
11. Report changed files, design choices, verification performed, and remaining risks.

## Review Mode

When reviewing existing terminal output, lead with findings:

- Does the verdict appear early and use human language?
- Are blockers ordered before passing details?
- Is each remediation exact and copy-pasteable?
- Does the output avoid internal status names and implementation details?
- Does colour carry semantic meaning only?
- Do TUI and plain-text modes preserve the same information?
- Is the success state quieter than the failure state?
- Would a user know the next action without reading source code or docs?
- Does the wording follow the copy style guide's voice, vocabulary, and remediation rules?
- Does the pattern catalog need a new or revised example?

## Revision Mode

When revising output:

- Preserve established StageServe semantics unless there is a clear design reason to change them.
- Prefer deleting noisy lines over adding explanatory prose.
- Prefer existing helpers over one-off rendering.
- Keep TUI and plain-text paths aligned.
- Update docs and examples when the change alters a reusable pattern.
- Run `go test ./core/onboarding ./cmd/stage/commands` when changing onboarding or command output behavior.

## Guardrails

Do not:

- Print internal enum names, raw check IDs, or status tags as user-facing copy.
- Add decorative colour, borders, banners, or boxes.
- Add a command that is not an exact shell invocation.
- Add TUI-only information that plain text does not contain.
- Keep old output patterns alive after the design contract has moved past them.
- Treat examples as rigid templates when the user journey needs a better pattern.

Do:

- Start from user state and next action.
- Make the output readable without colour.
- Use styling to reinforce meaning.
- Keep copy calm, short, and specific.
- Make evolution explicit through design reasons and catalog updates.
