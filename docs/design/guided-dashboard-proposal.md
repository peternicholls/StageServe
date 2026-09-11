# Guided Dashboard Proposal

This proposal updates the guided shell dashboard area used in the StageServe mockups. It treats the top area as useful product chrome, not as a dense status strip.

## Goal

The dashboard should help a user answer four questions quickly:

1. Where am I?
2. What matters now?
3. What evidence supports that state?
4. What can I press next?

Older mockups showed four equal-weight status facts. That made the area compact, but it did not make the most important information obvious. The revised design gives the dashboard a stronger hierarchy.

## Proposed Structure

```text
◆ StageServe                                      Running

This project is running at https://demo.test.
● Site online https://demo.test    ● DNS ready *.test routes locally    ● Apache + PHP up Up 3 hours
────────────────────────────────────────────────────────────────────────
[o] open browser   [l] logs   [s] status   [r] restart   [x] stop
```

Use this structure for guided and utility surfaces:

- **Product and state:** the first line establishes StageServe and the current surface state. The gradient background on this line spans only the text width, matching the footer rule width — not the full terminal width.
- **Headline:** one human sentence names the verdict, next step, or utility context.
- **Evidence facts:** three or four compact facts support the headline without becoming a table.
- **Command strip:** relevant keys sit directly below the facts.

## State Examples

Ready:

```text
◆ StageServe                                      Ready

Next: run this project.
● Machine ready Docker and ports passed    ● DNS ready *.test points here    ● Project stopped Configured, not running
────────────────────────────────────────────────────────────────────────
[↵] run project   [e] edit settings   [d] diagnostics   [m] more
```

Blocked:

```text
◆ StageServe                                      Needs attention

Next: open Docker Desktop.
● Docker needs action Open Docker Desktop first    ● DNS not checked Waiting for Docker    ● Ports not checked Waiting for setup
────────────────────────────────────────────────────────────────────────
[↵] open Docker   [d] diagnostics   [m] setup commands   [q] quit
```

Error:

```text
◆ StageServe                                      Action error

Next: run diagnostics.
● Start failed Project did not come online    ● Files safe No project files changed    ● Recovery guided actions available
────────────────────────────────────────────────────────────────────────
[↵] run diagnostics   [t] try again   [m] more tools   [q] quit
```

## Design Rules

- Do not make all facts equal. The headline owns the main message.
- Prefer a next-step headline when the user is blocked or ready to act.
- Prefer a verdict headline when the project is running or a utility report is open.
- Keep evidence facts short enough to scan. Put details in the body or More panel.
- Use semantic colour only: green for ready, yellow for attention, red for errors, cyan for focus or command-like affordances.
- Keep the command strip local to the current state. Do not advertise keys that are unavailable on the screen.
- In narrow terminals, wrap facts onto additional lines before truncating important values.

## Relationship To The Body

The dashboard should not replace the main content. It gives orientation and fast controls. The body still owns deeper explanation, project settings, checklists, action descriptions, confirmations, details, and reports.

For guided screens, the body can repeat the headline only when repetition improves confidence. If the body already starts with the same verdict, keep the dashboard headline shorter and more operational.
