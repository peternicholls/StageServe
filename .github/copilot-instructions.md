# StageServe Copilot Instructions

StageServe is an Apple-container-only Go CLI/TUI targeting local PHP/MariaDB development. Spec 012 and constitution 3.1.0 define the target; live runtime acceptance remains pending. Treat the Go implementation as the active product; Bash wrappers and older workflow material are reference-only unless the task explicitly targets archive cleanup.

Prefer the current StageServe contract over legacy compatibility behavior. Use `stage` as the canonical CLI entrypoint, keep `STAGESERVE_STACK=20i` explicit, use the Apple manifests under `stacks/20i/` as the runtime starting point; validate their behaviour before claiming support. Docker/Compose fallback is unsupported.

Preserve the current config ownership model. Project-local overrides belong in `<project>/.env.stageserve`, stack-wide defaults belong in `<stack-home>/.env.stageserve`, and machine-generated runtime env files under `.stageserve-state` are not user-owned inputs.

When changing config, lifecycle, naming, or runtime behavior, keep implementation, operator docs, and active spec-012 artifacts aligned. Update the relevant files together rather than letting README, docs, and workflow-contract/spec text drift apart.

Treat `archive/` and `previous-version-archive/` as immutable historical reference only. Do not restore legacy `20i-*` wrappers, archived TUI plans, or removed migration fallbacks unless the user explicitly asks for archival or compatibility work.

Prefer focused validation over broad test runs. Use the narrowest relevant checks first, especially `go test ./cmd/stage/commands`, `go test ./core/config`, and `go test ./core/lifecycle` when those areas change. Use `make test`, `make vet`, or `make lint` only when the change scope justifies it.

Keep changes minimal and contract-driven. Avoid inventing new config surfaces when an existing stack-home, project-local, or runtime-owned boundary already exists.

For TUI work, prefer the project-local Charm skill pack under `.github/skills/` as the implementation reference. Start with `charm-tui-builder`, then use `charm-lipgloss-layout`, `charm-bubbletea-components`, `charm-huh-forms`, `charm-tui-motion-observability`, and `charm-tui-qa` as the task demands.
Detailed contracts: `specs/012-apple-only-experience/contracts/`. Research gates are not implementation evidence. Follow `ux-validation.md` before full dashboard work.
