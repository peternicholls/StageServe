# Supported JSON output, schema version 1

Onboarding JSON (`stage setup --json`, `stage doctor --json`, and
`stage init --json`) and the hidden `stage guidance-plan` inspection
command emit one JSON document followed by a newline. They do not add terminal
formatting or interactive prompts. Each top-level object includes the integer
`schema_version: 1`. The onboarding projector sets this version even when a
caller constructs a result directly.

This contract does not add JSON support to `stage status` or `stage logs`.

## Onboarding envelope

Required keys are `schema_version`, `overall_status`, `exit_code`, and `steps`.
Resolved project commands add `project_scope`, with `dir` and `slug`, and
`project_id` only when registered ownership validates through the state store.
Fresh projects omit `project_id`; unresolved configuration omits `project_scope`.
Setup and doctor report the project context used for their configuration-driven
machine checks. Configuration failures are error envelopes with code
`project-config-invalid`; damaged ownership or journals produce an explicit
`project-state-invalid` step and omit unvalidated UUIDs. Init does not write
settings when scope validation fails, including dry runs.
`overall_status` is `ready`, `needs_action`, or `error`; the existing numeric
exit code remains unchanged. `BuildResult` represents no steps as `[]`.

Each step has `id`, `label`, `status`, `message`, and `remediation`. An absent
remediation is explicitly `null`. Empty `code` and `meta` are omitted. Empty
`result` and `next_steps` are also omitted. These omission meanings remain
unchanged. Step order is execution/report order.

`result` and `meta` are command-owned payloads; the version addition does not
rename, redact, or reinterpret their values. Producers must continue to avoid
putting credentials into user-facing fields. The projector is a serializer,
not a credential scrubber.

## Guidance inspection envelope

Required keys are `schema_version`, `situation`, and `status_header`. Empty
`summary`, `decision_items`, `work_items`, `visible_defaults`, `direct_commands`,
and `warnings` remain omitted.

`project_scope` is present when configuration resolves, with `dir` and `slug`.
Its optional `project_id` comes only from a successfully read, validated state
record; unregistered projects do not infer a UUID. Configuration errors omit
unresolved scope.

Nested objects preserve the existing Go field-name casing:

- Actions: `ID`, `Kind`, `Label`, `Description`, `InternalName`, `MutatesState`,
  `RequiresConfirmation`, `DirectCommand`, `ExpectedResult`, `Inputs`.
- Work items: `Label`, `Status`, `Description`, `DirectCommand`.
- Visible defaults: `Label`, `Value`, `Note`.

Nested empty values are still serialized (for example `Inputs: null`). The
inspection view selects plan fields and does not serialize the full collected
configuration or environment. Password environment values are excluded.

## Compatibility

Consumers should inspect `schema_version` and ignore unfamiliar additive keys.
Changes to existing key meanings, casing, or omission/null behavior require an
explicit contract revision and corresponding regression coverage.

Scope validation compares canonical absolute project paths, so a symlink alias
retains the same validated UUID. Existing installation ownership is validated
even when no identity records remain; damaged or newer installation schemas
produce an explicit scope error. Any scope validation failure clears the UUID.
