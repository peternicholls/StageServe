---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe TUI & Output Style Guide

This file is the legacy entrypoint for StageServe terminal output guidance.

The TUI guidance now lives in a small indexed design system so the product identity, rendering mechanics, examples, and long-form goal prompt can evolve independently.

## Start Here

Read `.github/instructions/terminal-experience-index.instructions.md`.

Use the files in this order:

1. `.github/instructions/terminal-design-contract.instructions.md`
   - Product identity, UX principles, copy rules, and guardrails.
2. `.github/instructions/terminal-markup-spec.instructions.md`
   - Concrete TUI/plain-text rendering rules, colour palette, layout, helper functions, and output constraints.
3. `.github/instructions/terminal-copy-style.instructions.md`
   - Voice, vocabulary, labels, verdicts, remediation, next actions, and copy review checks.
4. `.github/instructions/terminal-pattern-catalog.instructions.md`
   - Current and anticipated StageServe terminal patterns, examples, anti-examples, and reasoning.
5. `.github/instructions/terminal-experience-goal.instructions.md`
   - Goal prompt for agents designing, reviewing, or revising terminal interactions.

## Compatibility Note

The concrete rendering rules that used to live here have moved to `terminal-markup-spec.instructions.md`.

Do not add new rules to this file. Update the appropriate design-system file instead, then revise the index if the reading order or authority model changes.
