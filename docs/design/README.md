# StageServe Design Reference

These artifacts define how StageServe terminal interfaces should be designed, written, reviewed, and evolved.

Use them when changing command output, creating a new guided flow, reviewing agent-generated terminal UX, or writing user-facing copy for the CLI.

## Artifact Map

1. [Terminal Experience Style Guide](terminal-experience-style-guide.md)
   - Product identity, interface principles, information hierarchy, action design, and guardrails.
2. [Terminal Copy Style Guide](terminal-copy-style-guide.md)
   - Voice, vocabulary, sentence patterns, labels, remediation, warnings, and copy review rules.
3. [Terminal Interface Prototypes](terminal-interface-prototypes.md)
   - Reference interface sketches for readiness checks, onboarding, empty states, dry runs, destructive actions, long-running work, and plain-text parity.

## Who Should Use This

- Human designers defining a new terminal interaction.
- Engineers changing `stage` command output.
- Agent designers prompting Codex, Copilot, or another coding agent to revise terminal UX.
- Reviewers checking whether copy and output structure still feel like StageServe.

## Source Of Truth

The human-facing docs in this folder are the readable reference surface.

The agent-facing instruction files under `.github/instructions/` mirror the same system for tools that can automatically load scoped guidance:

- `.github/instructions/terminal-experience-index.instructions.md`
- `.github/instructions/terminal-design-contract.instructions.md`
- `.github/instructions/terminal-markup-spec.instructions.md`
- `.github/instructions/terminal-copy-style.instructions.md`
- `.github/instructions/terminal-pattern-catalog.instructions.md`
- `.github/instructions/terminal-experience-goal.instructions.md`

When changing the design system, update both the human docs and the matching agent instruction file in the same change.
