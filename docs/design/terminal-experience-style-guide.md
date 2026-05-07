# StageServe Terminal Experience Style Guide

StageServe terminal output is a product interface, not incidental console text. Most users will meet StageServe through the terminal, so every command should feel calm, intentional, useful, and recognisably part of the same product.

This guide is for humans and agent designers. Use it before deciding exact wording, ANSI markup, helper functions, or implementation details.

## Identity

StageServe terminal UX should feel:

- **Clear:** every line helps the user understand state, importance, or action.
- **Competent:** the command knows what happened, what matters, and what should happen next.
- **Calm:** failures are direct without being dramatic, apologetic, or noisy.
- **Practical:** remediation is concrete, copy-pasteable, and close to the problem.
- **Consistent:** repeated concepts use repeated structure, language, and colour semantics.
- **Evolvable:** the system supports new workflows without freezing early implementation choices.

Avoid generic CLI habits: status dumps, internal names, decorative banners, raw enum values, unexplained counters, and vague "try again" messages.

## Design Layers

Think about every terminal interaction in three layers:

- **UI design:** hierarchy, spacing, colour semantics, icons, grouping, scanability, and rhythm.
- **UX design:** user journey, information order, default path, blocked path, next actions, follow-up commands, and automation boundaries.
- **Language:** labels, explanations, problem statements, remediation copy, verdicts, and the detail needed for the user's current state.

Do not solve one layer while ignoring the others. A tidy output that leaves the user unsure what to do is not successful. Correct remediation buried in noisy prose is not successful.

## Interaction Model

Each command should answer four questions in order:

1. **Where am I?** Name the command surface or product area.
2. **What is the state?** Give a human verdict before details.
3. **What matters?** Surface blockers, warnings, and confirmations in priority order.
4. **What can I do next?** Provide the next useful action when there is one.

Prefer progressive disclosure. Start with the outcome, then the reason, then the action. Do not force users to infer status from a table of raw checks.

## Information Hierarchy

Use whitespace, order, and weight to create hierarchy. Do not rely on boxes, repeated dividers, or decorative frames.

Detailed output should support reading top to bottom. Compact output should support quick scanning.

When there are problems, lead with problems. Passing checks can appear afterward as reassurance, not as a barrier before the user sees what needs attention.

When everything passes, the output should become quieter. Do not print ceremonial success text or a long proof list unless the command context needs it.

## Action Design

Commands shown to users must be exact shell invocations. Do not wrap commands in explanatory prose, placeholders, or half-commands.

When a single next action is clearly best, provide it directly.

When multiple next actions are valid, separate the decision from the commands. Explain the choice briefly, then show exact commands.

When a command can safely perform the next step automatically, prefer doing the work over asking the user to manually stitch steps together. When automation would be risky, explain the risk in human terms and stop at a clear command or confirmation.

## Colour And Markup

Colour carries semantic meaning only. It is not decoration, brand garnish, or a way to make output feel more designed.

Plain text output must carry the same information as styled output. Styled output may add emphasis and colour; it must not add or remove content.

Use the markup spec in `.github/instructions/terminal-markup-spec.instructions.md` when changing renderer code.

## Review Checklist

Before approving terminal output, check:

- The first meaningful line establishes context.
- The verdict is human-readable and appears before detailed diagnostics.
- The most important user action is obvious.
- Any command shown is exact and copy-pasteable.
- The same information exists in TUI and plain-text output.
- Colour has semantic purpose.
- Copy avoids internal status names and implementation details.
- Passing and failing states are both intentionally designed.
- The output gets quieter when there is less to do.
- A new or changed pattern has a prototype example.
