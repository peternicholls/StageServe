---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe Terminal Design Contract

StageServe terminal output is a product interface, not incidental console text. It is the StageServe sub-brand most users will see most often, so every command should feel calm, intentional, useful, and recognisably part of the same product.

This contract defines design intent. It is platform-agnostic: humans, Codex, and other agents should use it before reasoning about ANSI codes, helper functions, or exact markup.

Use `.github/instructions/terminal-copy-style.instructions.md` for concrete voice, vocabulary, label, verdict, and remediation rules after this contract establishes the experience.

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

Think about every terminal interaction in three complementary layers.

**UI design:** visual hierarchy, spacing, colour semantics, icons, grouping, scanability, and rhythm.

**UX design:** user journey, information order, default path, blocked path, next actions, follow-up commands, and whether the command should continue automatically.

**Language:** labels, explanations, problem statements, remediation copy, verdicts, and the level of detail needed for the user's current state.

Do not solve one layer while ignoring the others. A visually tidy output that leaves the user unsure what to do is not successful. A correct remediation buried in noisy prose is not successful.

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

When a single next action is clearly best, provide that action directly.

When multiple next actions are valid, separate the decision from the commands. Explain the choice briefly, then show exact commands.

When a command can safely perform the next step automatically, prefer doing the work over asking the user to manually stitch steps together. When automation would be risky, explain the risk in human terms and stop at a clear command or confirmation.

## Copy Rules

Write like a capable product surface, not a logger.

- Use natural human language, not internal state names.
- Keep explanations factual and short.
- Prefer present-tense descriptions of state.
- Avoid "please", apologies, cheerleading, jokes, and theatrical failure language.
- Avoid "Note:" and "you must" in check descriptions.
- Use "To fix:" or "Next:" for remediation labels consistently.
- End explanatory sentences with punctuation.
- Do not print status tags such as `needs_action`, `StatusReady`, or `[docker.binary]`.

The user should never need to know the internal data model to understand the output.

## Colour And Meaning

Colour carries semantic meaning only. It is not decoration, brand garnish, or a way to make output feel more designed.

Each colour must have one clear purpose. If a new state needs a new colour, update the markup spec and add an example before using it.

Plain text output must carry the same information as styled output. Styled output may add emphasis and colour; it must not add or remove content.

## Guardrails

Protect against these forms of drift:

- **Brand drift:** output becomes generic, noisy, decorative, or mechanically "CLI-like".
- **UX drift:** commands dump state instead of guiding users through a useful flow.
- **Copy drift:** wording becomes verbose, apologetic, robotic, cute, or over-explanatory.
- **Markup drift:** styling changes break semantic colour, hierarchy, parity, or copy-pasteability.
- **Pattern drift:** old examples continue shaping new work after the product has moved on.

When changing an established pattern, name the design reason in the change description and update the pattern catalog if the change should guide future work.

## Design Review Checklist

Before approving terminal output, check:

- The first meaningful line establishes context.
- The verdict is human-readable and appears before detailed diagnostics.
- The most important user action is obvious.
- Any command shown is exact and copy-pasteable.
- The same information exists in TUI and plain-text output.
- Colour has semantic purpose and follows the markup spec.
- Copy avoids internal status names and implementation details.
- Passing and failing states are both intentionally designed.
- The output gets quieter when there is less to do.
- A new or changed pattern has an example when it should guide future work.
