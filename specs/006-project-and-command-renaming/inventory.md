# Rename Inventory: Active `stacklane` Surfaces

## Purpose

This file is the concrete Gate B disposition table for spec 006.

It records repository-owned active surfaces that still carry the old name and classifies each one as:

- rename now
- rename later in this spec
- historical only
- archive only

`rename later in this spec` is a staging classification, not a final-state exception.

## Disposition Table

| Surface | Examples observed | Disposition | Why | First implementation anchor |
|---|---|---|---|---|
| CLI root command and help text | `cmd/stacklane/commands/root.go`, `version.go`, `setup.go`, `doctor.go`, `logs.go`, `up.go`, `down.go`, `attach.go`, `status.go`, `dnssetup.go`, `init.go` | rename now | These files define the active operator-facing command surface and help text | `cmd/stacklane/commands/root.go` |
| Installer command/output surface | `install.sh` asset names, output messages, next-step commands, install destination | rename now | The installer is a direct public contract and must cut over early | `install.sh` |
| Build output names | `Makefile` `BINARY := stacklane-bin` and `go build -o $(BINARY) ./cmd/stacklane` | rename now | Build outputs and installer expectations need one canonical binary story | `Makefile` |
| README and install docs | `README.md`, `docs/installer-onboarding.md` | rename now | These are the primary entry points for new operators | `README.md` |
| Runtime and migration docs | `docs/runtime-contract.md`, `docs/migration.md`, `docs/architecture.md`, `docs/contributing.md` | rename now | These are maintained docs and must not keep the old active contract | `docs/runtime-contract.md` |
| Active normative specs and quickstarts | maintained examples under `specs/004-*`, `specs/005-*`, and active planning text under `specs/006-*` | rename now | They still teach the current command and naming model | `specs/005-installer-and-onboarding/quickstart.md` |
| Compatibility shim | temporary `stacklane` forwarding path if retained | rename now, then remove | It may exist only as a transition aid, not as a closeout surface | shim file/path to be decided in Gate A |
| Go module path | `go.mod` `module github.com/peternicholls/stacklane` | rename later in this spec | This is active code identity, but riskier than the command cutover and should follow the public command rename | `go.mod` |
| Go import paths | imports under `cmd/`, `core/`, `infra/`, `observability/`, `platform/` still point at `github.com/peternicholls/stacklane/...` | rename later in this spec | Must be migrated together with the module path to avoid broken builds | `cmd/stacklane/main.go` |
| Package and directory names | `cmd/stacklane/**` and any repository-owned paths still carrying `stacklane` | rename later in this spec | Path/layout rename is active code identity work and should be sliced carefully | `cmd/stacklane` |
| Internal env and state filenames | `.env.stacklane`, `.stacklane-state`, `.stacklane-local` references, generated text that still emits them | rename later in this spec | They are active surfaces today but must not remain by closeout | `core/config/loader.go` |
| Internal env variable names | `STACKLANE_*`, `STACKLANE_PROJECT_ENV_*`, `STACK_HOME` comments or active emitted config text where rename is required | rename later in this spec | Active code identity and config contract must eventually drop the old brand | `core/config/types.go` |
| Runtime Docker prefix | `stln-<slug>` compose project, `stln-<slug>-runtime` network, `stln-<slug>-db-data` volume, `stln-shared` gateway project/network — set in `core/config/loader.go` lines ~297-334 | rename later in this spec | Must rename to `stage-*` by closeout; no live migration needed (local dev only, `docker system prune` acceptable) | `core/config/loader.go` |
| Internal generated config text | `cmd/stacklane/commands/project_env.go`, `core/onboarding/project_env.go` comments and emitted file content | rename later in this spec | These are user-visible generated artifacts but depend on the internal rename decision | `cmd/stacklane/commands/project_env.go` |
| Lifecycle remediation messages | guidance strings like `Run \`stacklane up\` first` or `Check \`stacklane logs\`` in `core/lifecycle/orchestrator.go` | rename later in this spec | Active behavior messaging must align after the command cutover | `core/lifecycle/orchestrator.go` |
| Gateway and observability headers | `X-Stacklane-*`, `stacklane-no-route`, `__stacklane_gateway_health`, gateway error strings in `infra/gateway/**` | rename later in this spec | These are active runtime-visible surfaces and must be renamed deliberately with testdata updates | `infra/gateway/templates.go` |
| Docker label namespace | `io.stacklane.service` comments or label contracts in `infra/docker/types.go` | rename later in this spec | Active runtime metadata should not retain the old brand by closeout | `infra/docker/types.go` |
| Platform-specific filenames and remediation text | `stacklane-*.conf`, `stacklane setup`, `Stacklane managed include` under `platform/dns/**` | rename later in this spec | Active system integration artifacts and messages must be renamed carefully | `platform/dns/macos.go` |
| Test fixtures and assertions that model active behavior | `cmd/**_test.go`, `core/**_test.go`, `infra/gateway/testdata/**`, `observability/**_test.go`, `platform/**_test.go` | rename later in this spec | Test expectations must follow the active contract once implementation slices land | owning file for each touched slice |
| Historical CLI analysis | `docs/cli-naming-analysis.md` | historical only | It may retain the old name as analysis, but must be clearly historical | `docs/cli-naming-analysis.md` |
| Archived wrappers and legacy UI | `previous-version-archive/**` | archive only | Archive content must not drive the active rename | `previous-version-archive/` |

## Open Questions Still Requiring Gate A Decisions

1. What are the final replacement names for `.env.stacklane`, `.stacklane-state`, and `STACKLANE_*`?
2. Is the module path changing in the same spec closeout window, or does the spec need to define a repository-rename dependency?
3. Will a temporary `stacklane` shim be shipped at all, and if yes, what is the exact removal point before closeout?
4. Which runtime-visible headers and health endpoints under `infra/gateway/**` need exact replacement names versus compatibility aliases during transition?
5. ~~Does `stln-*` stay as the runtime prefix?~~ **Decided**: rename `stln-*` → `stage-*`. No live migration required; local dev only.

## First Execution Slices Suggested By Inventory

1. Public command and help cutover: `cmd/stacklane/commands/root.go`
2. Installer and asset naming cutover: `install.sh` and `Makefile`
3. Top-level docs cutover: `README.md` and `docs/installer-onboarding.md`
4. Internal naming contract slice: `core/config/loader.go` and `core/config/types.go`
5. Module/import path slice: `go.mod` and `cmd/stacklane/main.go`
6. Gateway header and testdata slice: `infra/gateway/templates.go` and `infra/gateway/testdata/**`