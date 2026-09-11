# StageServe product direction

Decision date: 2026-09-11. Target product, not a claim that the current checkout is release-ready.
Authority: [constitution 3.0.0](../.specify/memory/constitution.md), [spec 012](../specs/012-apple-only-experience/spec.md). Delivery: [roadmap](roadmap.md).

## The product we are building

StageServe is a local PHP/web development environment for a developer on an Apple silicon Mac. Open a project folder, run `stage`, review the detected settings, and run the site at a predictable local address. StageServe owns service ordering, local routing, health, persistence and recovery so the operator does not have to assemble container commands.

Apple `container` is the only runtime. The initial supported stack is the existing 20i-style web/PHP/MariaDB stack. This is local development with useful production resemblance, not a promise of exact 20i hosting parity or a generic container manager.

## Interface decision

Ship a polished guided TUI as the primary v1 experience, alongside stable direct CLI commands, plain text and JSON. Retain Go, Cobra, Bubble Tea, Bubbles, Huh and Lip Gloss already present in the repository. Do not rewrite the backend in Swift merely because Apple's runtime uses Swift.

A native GUI is outside v1. Revisit only after the TUI release and observed user trials show that terminal entry itself prevents intended users completing the workflow. A future GUI must use the same application services and versioned contracts; choosing AppKit, SwiftUI, a webview or a transport is deferred until that gate. The HTML mockups are a design viewer, not an application frontend to ship.

Why: there is already a planner, production guided renderer, report vocabulary, terminal style guide and broad mockup coverage. Two interface implementations would increase testing and packaging work before Apple networking and persistence have live proof. The strongest alternative is a native app with folder selection and a menu-bar overview; its discovery benefits do not yet justify a second product surface.

## A normal session

1. `stage` in an unconfigured folder shows the folder, proposed name, web folder and local URL. It explains any missing machine dependency before offering changes.
2. Confirm settings once. Project file creation is previewed; DNS and certificate trust are separate, explicit host changes.
3. Run the project. Progress names the current step, elapsed time, cancellation behaviour and useful recovery. Success requires working application traffic, not just started processes.
4. The running dashboard shows the project name, URL, observed health and relevant keys. Enter views logs (safe default); Open browser is explicit. Status, restart, stop, settings and More remain discoverable.
5. Stop retains source and database data. Starting again retains identity, URL and data. Removing a project record and deleting its data are separate actions.
6. With several projects, show an explicit selected project. Operations never silently switch scope; shared routing repairs must preserve all other healthy projects.

## Screen contract

Use the current visual/copy guides for semantic colour, spacing and terminology. The selected dashboard is product/state header -> one verdict -> three or four short evidence facts -> local command strip -> focused body. Avoid repeating the same verdict and facts in the body. No essential information depends on colour, glyphs, gradients or a wide terminal.

| Situation | Primary action | Evidence visible before action |
|---|---|---|
| Machine unavailable | Show next actionable setup step | Missing tool/service, supported version, exact effect |
| Not a project / missing settings | Preview project setup | Folder, web folder, name, suffix, URL, file path |
| Ready or stopped | Run this project | Selected project and resolved settings |
| Starting/stopping | Observe progress / cancel safely | Step, elapsed time, retained resources |
| Running | View logs | URL, application health, selected project |
| State mismatch | Preview safe recovery | Desired/observed differences and affected resources |
| Error or unknown state | Read-only diagnostics | Failed step, what changed, next action |
| Destructive confirmation | Cancel by default | Exact project, volume/data scope and backup advice |

52-column and 80-column layouts, light/dark terminals, NO_COLOR, non-TTY, resize and cancellation are release acceptance cases. The browser viewer's width controls are illustrative; real terminal checks are required.

## Scope boundaries

Included in v1: Apple host readiness, one stack, guided project settings, reliable lifecycle, local HTTP `.test` routing, explicit opt-in local TLS, logs/status, multi-project selection, safe recovery, database preservation, offline-friendly cached operation, checked installer/update/rollback documentation and migration guidance.

Excluded: Docker/Compose fallback or compatibility, Docker socket consumers, Linux/Intel host support, arbitrary Compose import, general stack marketplace, cloud deployment, remote/multi-user management, public network hosting, automatic Docker volume conversion, separate DNS product, native GUI and unattended destructive repair.

Retain `.test` as the new-project default. Existing explicit `.develop` and `.dev` settings must never be silently rewritten; migration validates them and reports unsupported combinations. Do not describe `.test` as zero-setup until host DNS acceptance proves it.

## Current truth and authority

The working tree already defaults to Apple only and wires the Apple adapter into CLI/TUI paths. It still contains Docker-era tests, assets and documentation. Live Apple acceptance has not run because this host has no `container` executable. This direction does not certify those implementation details.

Current authority order: constitution -> spec 012 and its contracts -> product direction and roadmap -> design guides. Historical specs and older implementation narratives are evidence only. The immutable archive preserves former decisions, including rejected ones; it is not a second backlog.
