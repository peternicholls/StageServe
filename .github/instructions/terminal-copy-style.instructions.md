---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe Terminal Copy Style Guide

This guide governs the words inside StageServe terminal interfaces: labels, verdicts, explanations, remediation, next actions, warnings, and empty states.

Use it when designing output, changing command text, reviewing agent-written copy, or deciding whether a message belongs in terminal output at all.

## Voice

StageServe sounds like a capable local-development tool.

- **Plain:** use common words before implementation terms.
- **Specific:** name the state, blocker, or action.
- **Calm:** describe failure without drama, apology, or blame.
- **Operational:** write copy that helps the user continue.
- **Short:** remove lines that do not change user understanding or action.

Do not write like logs, docs, marketing, or a chat assistant. Terminal copy is product UI.

## Copy Hierarchy

Write terminal output in this order:

1. **Context:** the command surface or product area.
2. **Verdict:** the human state of the interaction.
3. **Reason:** the blocker, confirmation, or relevant detail.
4. **Action:** the exact next command or next decision.

If a line does not serve one of those roles, remove it.

## Vocabulary

Use stable user-facing terms.

| Concept | Preferred copy | Avoid |
|---|---|---|
| Product | `StageServe` | `stage runtime`, `gateway manager` |
| Success verdict | `Ready - all checks passed.` | `Success!`, `OK`, `OverallReady` |
| Blocked verdict | `Not ready - N of M checks need attention.` | `Failed`, `needs_action`, `Result: Needs action` |
| Fix label | `To fix:` | `Run:`, `Please run`, `Try` |
| Next step label | `Next:` | `You should`, `Continue by`, `Recommended` |
| No items | `No projects are registered yet.` | `0 projects`, `Empty list` |
| Planned change | `Will update` | `Mutation plan`, `Diff payload` |

Use internal IDs only in tests, logs, or developer-facing diagnostics. Do not print them as user copy.

## Sentence Patterns

Verdicts:

```text
Ready - all checks passed.
Not ready - 2 of 7 checks need attention.
Local setup is not ready yet.
No projects are registered yet.
Previewing changes. Nothing has been changed.
```

Reasons:

```text
Docker is installed, but the daemon is not running.
Another process is already listening on port 443.
*.test domains are not resolving to localhost.
The project will stop routing through the local gateway.
```

Actions:

```text
To fix:  open -a Docker
Next:  stage init
Next:  stage project remove api.test --confirm
```

## Remediation Copy

Remediation must be an exact shell command, not an instruction sentence.

Good:

```text
To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN
```

Bad:

```text
To fix:  Check what is using port 443 and close it.
Run: sudo lsof -nP -iTCP:443 -sTCP:LISTEN (shows listeners)
```

When there are multiple valid fixes, explain the choice in a sentence above the commands. Keep each command copy-pasteable.

## Labels

Labels should match user concepts and fit compact terminal layouts.

- Prefer `Docker daemon` over `docker.daemon`.
- Prefer `State directory` over `state.dir`.
- Prefer `DNS resolver` over `dnsmasq check`.
- Prefer `Port 443` over `https_port_available`.

If a label needs extra explanation, keep the label short and put the explanation in the description line.

## Descriptions

Descriptions explain why a check or step matters. They are not instructions.

Good:

```text
The Docker daemon must be running before any container can start.
Port 443 must be free for the local HTTPS gateway to bind to it.
```

Bad:

```text
You must start Docker.
Note: port 443 is required.
This is used by the runtime subsystem.
```

## State-Specific Guidance

Success should be quiet. Give the verdict, concise evidence when useful, and the next command if there is a natural handoff.

Failure should be useful. Lead with the blocker, keep passing details secondary, and show an exact next action.

Empty states should teach one next step. Do not print empty tables.

Warnings should state the consequence. If an action is destructive, say what changes and what does not change.

Long-running work should name the current step and set expectation only when waiting is normal.

## Editing Checklist

Before shipping terminal copy, check:

- The verdict is human-readable.
- The first action is exact and copy-pasteable.
- Internal status names, IDs, counters, and enum values are absent.
- The copy still works without colour or icons.
- Success is quieter than failure.
- Descriptions explain why, not what command to run.
- Each line changes the user's understanding or next action.
- New reusable wording is reflected in the pattern catalog.
