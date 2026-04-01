# Runtime Contract

This document locks the implemented Phase 1 and Phase 2 command semantics and state model.

## Goals

- Preserve the existing shell-first workflow.
- Standardize project identity and future hostname derivation.
- Make `up`, `attach`, `down`, `detach`, and status behavior explicit.
- Provide a stable state model before the shared gateway and DNS work begins.

## Command Behavior

### `20i-up`

- Works from a project root.
- Resolves config from `.env`, `.20i-local`, environment variables, and CLI flags.
- Derives a project slug from the folder name by default.
- Plans a hostname of `<slug>.test` unless overridden.
- Ensures the shared gateway and shared Docker network exist.
- Starts the current Docker compose project.
- Writes a project state file and marks the project as `attached`.
- Updates the shared gateway to route its current default target to this project.

### `20i-attach`

- Uses the same runtime resolution as `20i-up`.
- Is the explicit command for bringing an additional repo into the managed set.
- Starts a second isolated per-project runtime and repoints the current shared gateway default route to it.

### `20i-down`

- Stops only the current project runtime by default.
- Retains the project record and marks it `down`.
- Repoints the shared gateway to another attached project when one exists, otherwise leaves the gateway on a no-route response.
- Supports `20i-down --all` for global teardown.

### `20i-detach`

- Stops only the current project runtime.
- Removes the project record entirely.
- Repoints the shared gateway to another attached project when one exists.

### `20i-status`

- Reports tracked projects from the state directory.
- Shows shared gateway health, planned hostname, shared localhost URL, document root, container docroot, project path, and Docker status.

## Config Precedence

1. CLI flags
2. `.20i-local`
3. Shell environment
4. Stack `.env`
5. Built-in defaults

## Hostname Derivation

- Default source: project folder name
- Override source: `SITE_NAME`
- Full override: `SITE_HOSTNAME`
- Default suffix: `.test`
- `.dev` remains deferred until the stack gains local HTTPS support

## Document Root Contract

- `DOCROOT` is the preferred project override.
- `CODE_DIR` remains as a legacy alias.
- Default behavior is `public_html` when present, otherwise the project root.
- The current 20i-style container layout mounts the whole repo at `/home/sites/<project-slug>`.
- The effective container docroot becomes `/home/sites/<project-slug>/<docroot-relative-path>`.
- For the current implementation, `DOCROOT` must resolve inside the project directory.

## Project State Model

- `attached`: runtime is intended to be active and recorded in state
- `down`: runtime has been stopped but the record is retained
- `detached`: record removed from active state storage
- global teardown: all records removed and all known runtimes stopped

## State Storage

- Location: `<stack-home>/.20i-state/projects/<slug>.env`
- One file per project
- Stores resolved runtime identity, container path layout, and published per-project service ports

## Shared Infrastructure

- Shared gateway compose file: `docker-compose.shared.yml`
- Shared network: `twentyi-shared` by default
- Shared gateway host ports: `80/443` by default, overrideable via `SHARED_GATEWAY_HTTP_PORT` and `SHARED_GATEWAY_HTTPS_PORT`
- Per-project web containers no longer publish host ports directly for normal site access
- phpMyAdmin remains per project in the current milestone and still publishes its own host port

## Phase Boundary

This contract now includes the shared gateway split, but it still uses one default localhost route rather than hostname-aware routing. The `.test` hostname remains the future routing target, and local DNS plus host-based gateway rules still land in later phases.