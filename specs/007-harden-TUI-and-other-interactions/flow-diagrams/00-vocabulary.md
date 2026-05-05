# Plain-Language Vocabulary

This file defines the words StageServe says to the user, and the words it never says.

## Hard No List (Never Used In First-Level User Text)

These words are accurate but useless to a normal user. They never appear in the bare `stage` TUI, `stage init`, `stage setup`, or `stage doctor` user-visible text. They are allowed only inside the optional `Advanced and troubleshooting` view, in `--cli` text output for power users, and in command help.

| Banned Word | Why It Is Banned |
|---|---|
| drift | Means nothing to a normal user. They do not know what should match what. |
| gateway | Implementation detail. The user wants a working URL, not a routing concept. |
| compose | Docker concept. The user did not ask for Docker. |
| container | Docker concept. |
| daemon | Operating-system concept. |
| runtime | Internal noun. |
| registry | Internal noun. |
| state record | Internal noun. |
| attach | Lifecycle term. The user wants to "add a project to StageServe". |
| detach | Lifecycle term. The user wants to "remove a project from StageServe". |
| bootstrap | Implementation step. The user wants their project to "start". |
| post-up hook | Implementation step. |
| docroot | Maybe-known to web devs, but say "web folder" first. |
| TUI | The user does not need to know they are in a TUI. |

## What We Say Instead

Examples use `.develop`, but the real interface says whatever suffix, scheme, and port StageServe resolved for the current project. Copy must never hard-code `.develop` when the active project is configured for `.test`, `.dev`, a full hostname, or a custom suffix.

| Concept | Plain-Language Phrase |
|---|---|
| `drift_detected` | "This project doesn't match what StageServe expects." |
| `machine_not_ready` | "Your computer isn't ready yet." |
| `project_missing_config` | "This folder doesn't have StageServe settings yet." |
| `project_ready_to_run` | "This project is ready to run." |
| `project_running` | "This project is running at \<URL\>." |
| `project_down` | "This project is stopped." |
| `not_project` | "This folder isn't a StageServe project yet." |
| `unknown_error` | "StageServe couldn't safely choose a next step." |
| Docker Desktop missing | "StageServe needs Docker Desktop to run your sites." |
| Docker daemon stopped | "Docker Desktop is installed but isn't running." |
| Port 80 in use | "Something else on your computer is using port 80." |
| DNS resolver missing | "Your computer can't yet open *.develop URLs." |
| mkcert root not installed | "Your browser doesn't trust StageServe's local HTTPS yet." |
| Compose project name | not shown |
| Container name | not shown |
| Network name | not shown |
| Volume name | not shown |
| `.stageserve-state` directory | "StageServe's hidden working folder" (only mentioned in advanced view) |
| `attach` action | "Add this project to StageServe" |
| `detach` action | "Remove this project from StageServe" |
| `up` action | "Run this project" |
| `down` action | "Stop this project" |
| `setup` action | "Set up your computer" |
| `init` action | "Set up this folder as a project" |
| `doctor` action | not shown as an action; the tool runs it itself when needed |
| `--cli` opt-out | "Plain text output" |

## How To Phrase A Blocker

A blocker has three parts, in this order:

1. What is wrong, in one sentence, in user terms. ("Docker Desktop isn't running.")
2. What to do, as a numbered list of physical actions. ("1. Open Docker Desktop. 2. Wait for the whale icon to settle.")
3. What StageServe will do next. ("When you've done this, press enter and StageServe will check again.")

A blocker never says the name of the failing internal check. The internal name appears only in the advanced view.

## How To Phrase A Recovery

When something has gone wrong and StageServe is recovering:

1. Say what does not match, in user terms. ("StageServe expected this project to be running, but it isn't responding.")
2. Say what the safe next step is, in one sentence. ("StageServe will treat this project as stopped. Nothing on disk will be deleted.")
3. Make that safe step the highlighted default. The user only has to press enter.

A recovery never says "drift", "diagnose", "doctor", "state mismatch", or "reconcile".

## How To Phrase A Confirmation

A confirmation has the file path or URL it is about to act on, the value it is about to use, and a `Yes/No` choice with `Yes` highlighted by default.

```
StageServe will create:
  /Users/pete/sites/pete-site/.env.stageserve

with these settings:
  Site name      pete-site
  Web folder     ./public_html
  Local URL      http://pete-site.develop

▶ Yes, create it    No, cancel

  ←/→ choose • enter confirm • esc cancel
```

The user can read what will happen before they commit. They never have to drill down to see the values.
