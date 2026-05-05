# Data Model: Guided TUI And Next-Action Planning

## Guided Session

Represents one no-args `stage` invocation.

Fields:

- `cwd`: directory where the command started.
- `interactive`: whether stdin/stdout can support TUI interaction.
- `tui_disabled`: whether shell environment disables TUI.
- `color_disabled`: whether color should be disabled.
- `plan`: the current `NextActionPlan`.
- `selected_action`: the current action, if any.
- `confirmed`: whether the user confirmed a mutating action.
- `result`: action result, when an action has run.

Rules:

- A session must not mutate state before confirmation.
- A session may be represented in TUI or text fallback.
- A session must preserve direct command equivalents for every action.

## TUI Capability

Represents terminal suitability.

Fields:

- `stdin_tty`
- `stdout_tty`
- `stderr_tty`
- `notui_flag`
- `cli_flag`
- `no_tui_shell_env`
- `no_color`
- `term`
- `reason`

Rules:

- TUI is allowed only when stdin and stdout are interactive, `notui_flag` is false, `cli_flag` is false, and `no_tui_shell_env` is false.
- Color is disabled when `NO_COLOR` is set or terminal capability is insufficient.
- Non-TTY fallback must not block for input.
- `STAGESERVE_NO_TUI` is shell-env-only and must not be read from project or stack `.env.stageserve`.

## Guided Context

A cheap snapshot of StageServe's current situation.

Fields:

- `cwd`
- `project_root`
- `project_env_path`
- `project_env_exists`
- `project_env_valid`
- `stack_home`
- `stack_id`: normalized `STAGESERVE_STACK` value, currently `20i`.
- `state_dir`
- `machine_readiness_summary`
- `project_state`
- `runtime_summary`
- `warnings`

Rules:

- Context collection should avoid expensive checks before first render.
- Expensive checks can be run when the user selects setup, doctor, status, or refresh.
- Context collection must not create `.env.stageserve`.

## Next Action Plan

Planner output consumed by TUI and text fallback.

Fields:

- `situation`: one of `machine_not_ready`, `project_missing_config`, `project_ready_to_run`, `project_running`, `project_down`, `drift_detected`, `not_project`, `unknown_error`.
- `title`
- `summary`
- `status_header`: one plain-language sentence that says what is true now.
- `decision_items`: zero to three user choices for the decision bar. Empty when StageServe is running a tool-owned checklist or recovery step.
- `work_items`: setup, progress, blocker, or recovery rows owned by StageServe rather than selected as menu actions.
- `footer_actions`: persistent affordances such as help, details, show direct commands, plain text output, advanced troubleshooting, and quit.
- `warnings`
- `direct_commands`
- `visible_defaults`: values and default actions the UI must show before commitment.
- `plain_language_terms`: copy decisions used by TUI and text fallback for the current plan.

Rules:

- A known situation must always have a status header and at least one safe next step, but that next step may be a tool-owned work item rather than a decision item.
- The decision bar must contain only real user goals, not diagnostic or implementation mechanics.
- Every action must include a direct command equivalent or an explicit reason why none exists.
- Every value in `visible_defaults` must render on the same screen where the user can commit to it.
- Warnings must be actionable.
- Titles, summaries, warnings, and action labels must describe user goals before implementation mechanics.
- Implementation words such as attach, detach, daemon, gateway, compose, container, registry, and runtime must stay out of first-level copy unless no StageServe-level phrase can explain the required action.

Canonical situation semantics:

| Situation | Meaning |
|---|---|
| `machine_not_ready` | Setup-level prerequisites block normal operation. |
| `project_missing_config` | Current directory can be a project, but `.env.stageserve` is absent. |
| `project_ready_to_run` | Project config exists and no retained down state or active runtime needs special handling. |
| `project_running` | Project is recorded and runtime/status checks indicate it is active. |
| `project_down` | Project has retained StageServe state marked down. |
| `drift_detected` | State, runtime, DNS, gateway, or config disagree and diagnosis is safer than normal action. |
| `not_project` | Current directory is not usable as a StageServe project root for guided project actions. |
| `unknown_error` | Context collection or planning failed without a safe classification. |

## Guided Surface

Represents one stable region of the guided TUI. Surfaces keep setup work, user decisions, and advanced material separate.

Fields:

- `surface`: `status_header`, `decision_bar`, `tool_work_panel`, `details_panel`, or `footer`.
- `visible`: whether the surface renders for the current situation.
- `items`: rows, actions, values, or messages rendered by that surface.
- `default_item`: the row or action triggered by `enter`, if any.
- `direct_command_refs`: commands exposed through footer/detail surfaces, not first-level labels.

Rules:

- The status header is always visible.
- The footer is always visible except during raw log follow mode, where log exit keys replace it.
- The decision bar is hidden while StageServe owns the next step, such as setup checklist execution or ordered recovery.
- Tool work panels may run checks, progress, or blocker instructions, but must not require the user to understand command names.

## Guided Action

Represents one user-selectable operation.

Fields:

- `id`: stable action id such as `setup`, `init`, `up`, `attach`, `status`, `logs`, `down`, `detach`, `doctor`, `diagnose`, `setup_help`, `recovery_help`, `edit_config`, `show_commands`, `advanced`.
- `kind`: `confirm`, `choose`, `tool_step`, `inline_form`, `footer`, or `advanced`.
- `label`: easy-mode label shown in the TUI, such as "Run this project" or "Remove this project from StageServe".
- `description`: plain-language explanation of what will happen and why it helps.
- `internal_name`: command or domain name used by implementation and "show commands", when different from the label.
- `mutates_state`
- `requires_confirmation`
- `direct_command`
- `expected_result`

Rules:

- Mutating actions require confirmation.
- Actions must route through existing command/domain behavior.
- Advanced actions may reveal implementation details, but decision-bar actions should not.
- Direct command names are command equivalents, not first-level labels.
- The TUI and text fallback must use the same `label` and `description` for the same action.
- Planner-only ids such as `diagnose`, `setup_help`, `recovery_help`, and `init_here` may alias existing command/domain behavior through `internal_name` instead of introducing new command surfaces.
- `doctor` and `diagnose` actions are footer, advanced, or tool-owned work items. They are not peer first-level choices next to "Run this project" or "Use these settings".

Recommended easy-mode labels:

| Internal Action | Easy-Mode Label |
|---|---|
| `setup` | Set up this computer |
| `init` | Create project settings |
| `up` | Run this project |
| `attach` | Add this project to StageServe |
| `status` | Check project status |
| `logs` | View project logs |
| `down` | Stop this project |
| `detach` | Remove this project from StageServe |
| `doctor` | Troubleshoot this problem |
| `diagnose` | Troubleshoot this problem |
| `init_here` | Set up this directory as a project |
| `setup_help` | Get setup help |
| `recovery_help` | Show recovery help |
| `edit_config` | Edit project settings |
| `show_commands` | Show commands |
| `advanced` | Advanced troubleshooting |

## Config Preview

Represents a pending `.env.stageserve` write.

Fields:

- `path`
- `values`
- `site_name`
- `web_folder`
- `domain_suffix`
- `url_scheme`
- `url_port`
- `local_url`
- `advanced_summary`
- `comments`
- `overwrite`
- `source`

Rules:

- Show before write.
- Preserve existing overwrite protection unless user explicitly confirms force behavior.
- Write only `.env.stageserve`.
- Values must be sourced from the same config/defaulting path the direct commands use. The preview must not invent a UI-only default.
- The URL preview must render the configured suffix, scheme, and port rather than hard-coding `.develop`, `.test`, `.dev`, or HTTPS.

## Recovery Path

Represents guidance for a non-ready or failed state.

Fields:

- `problem`
- `why_it_matters`
- `primary_fix`
- `direct_command`
- `advanced_detail`

Rules:

- The primary fix should be a StageServe command or config edit.
- Advanced detail can include Docker only when needed.
- Recovery steps must be ordered from least invasive to most invasive and pause after each step for re-planning.
