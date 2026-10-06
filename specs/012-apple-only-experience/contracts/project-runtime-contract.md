# Project identity, settings and runtime contract

Revised 2026-09-13. Normative target for FR-002–009, FR-013–014 and FR-016–018.
Existing code is the implementation starting point, not evidence of compliance.
Case IDs below are acceptance obligations, not passing test results.

## Durable identity and retained data

Choose an installation-owned retained-data ledger independent of registration.
Generate a cryptographically random UUID when first confirming project setup;
persist it in `<state-dir>/identities/<uuid>.json`. This record contains canonical
path, display slug, registration status, volume names/IDs, owner labels, database
engine digest/version, applied configuration fingerprint and schema version.
The ordinary project registry is a projection of registered identities, not the
only copy of identity. Detach sets `registered=false`; it does not delete this
ledger. Source settings need not contain or share a machine-local identifier.
Resource names use UUID-derived tokens; labels include exact installation UUID,
project UUID and service role. A matching display name alone never authorizes use.

Before any mutation, lock and reconcile ledger plus observed labels/IDs. Atomic
journal intent is written before creating resources; observed IDs are then saved.
If a crash occurs between creation and recording, discover only resources with
exact installation/project labels matching a pending journal; report ambiguous
matches without adoption/deletion. Never recover from prefix matching. Networks
without suitable label support require recorded exact IDs and creation provenance;
missing ledger requires manual inspection, not deletion. Volume creation must
supply installation and project labels supported by the pinned CLI.

| Event | Required behaviour |
|---|---|
| Detach then attach same canonical path | Reuse same UUID/owned data; revalidate engine and settings before use |
| Edit slug/hostname | UUID and volume unchanged; names in UI may change |
| Copy folder to another path | New identity and fresh data after setup preview; never attach original volumes |
| Move folder | Detect missing old path and potential collision; refuse implicit relocation. Recovery previews old/new path and explicit rebind only after operator verifies same project; until a supported rebind exists use documented export/import into a fresh identity |
| Settings removed | Existing path maps to retained identity; preview settings restoration/new values, never silently reset DB |
| Ledger missing/corrupt or duplicate path ownership | Read-only error with recover-from-backup instructions; no adoption or destructive cleanup |
| Explicit volume deletion | UUID-scoped preview and current owner check; remove exact recorded volumes, record tombstones; preserve source/settings |
| Uninstall | Retain identity ledger/data unless separately exported and explicitly removed |

Backups include ledger, registry, generated secret/config provenance and schema;
DB data backup is a separately verified logical export. Paths alone do not prove
that a replacement folder is the former project: mismatched settings require
explicit recovery preview. Legacy records missing runtime identity remain legacy.

## Settings apply is a transaction

Use one typed action `PreviewSettings(projectID, proposedValues)` returning base
revision, effective values with origins, changed fields, restart/route effects,
retained data and redacted secrets. `ApplySettings(previewRevision)` rechecks scope
and revision under the project lock; stale previews fail without changes.
CLI `init --force` on an existing project uses the same transaction; manual config
edits are detected before `up`/`attach`. A stopped project's edit is saved as desired
configuration and applied on next start; it never claims the runtime was updated.

Fingerprint canonical resolved runtime settings, image digests, manifest version,
mounts, probe configuration and route policy. Track desired and applied hashes.
Use an installation-keyed digest for secret-bearing inputs, not a public password
hash. Reports never reveal secrets or the keyed digest. Operational command flags
and transient addresses are excluded; addresses have separate observation revision.

| Field | Apply policy |
|---|---|
| UI display name | Metadata-only update; no resource rename |
| Slug, hostname, suffix, TLS choice | Validate uniqueness; prepare new route/cert; replace old route only after success; retain UUID |
| Web folder, source mount, PHP version, web image, PHP env | Recreate affected web/PHP services after explicit downtime preview; reuse exact compatible DB volume |
| Host DB/debug port | Reserve new port before change; recreate affected service if needed; release old reservation only after commit |
| DB user/name/password | No silent rewriting of an existing database. Reject generic apply and give explicit export/configuration/migration procedure; new empty volumes may use previewed values |
| DB engine version/digest | Refuse change against an existing volume in ordinary apply/update; use separately rehearsed migration into a new volume |
| Canonical project path/identity | Not editable in settings; use explicit recovery described above |

Transaction: validate -> acquire locks -> revalidate revision -> write durable
intent/backup -> build/pull before downtime -> prepare replacement configuration
-> stop affected services -> create/start/probe -> validate candidate gateway
configuration -> reload -> atomically commit desired/applied state and settings.
A journal bridges files that cannot be atomically renamed together. Project files
other than `.env.stageserve` are never edited. On failure, restore the prior
compatible service configuration and route using a fresh cleanup context. If
rollback cannot complete, mark degraded with exact leftovers; do not claim the
old project is running. Database contents are never rewound during settings apply.

Acquire a project UUID lock, then short-lived shared-route lock; do not acquire
project locks while holding the route lock. Broad operations obtain project locks
in UUID order. Route reload failure retains the last valid config; busy locks have
a bounded wait and clear retry advice. No lock is held awaiting user confirmation.

## Application configuration and health

StageServe never writes application `.env`, PHP source, credentials into public
files, or database rows as a readiness check. For existing projects, read only
DB_DATABASE/DB_USERNAME/DB_PASSWORD from application `.env` as fallback below all
explicit StageServe sources and above built-in defaults. Determine explicitness
by source presence, not equality with default strings. StageServe precedence is
CLI -> project `.env.stageserve` -> shell -> stack `.env.stageserve` -> application
DB fallback (these three keys only) -> defaults. Other application keys are ignored.
Record each origin. Existing explicit StageServe DB values always win. Bootstrap
commands remain explicit project-owned actions, disabled unless configured; they
are not required for infrastructure readiness or silently retried on failure.

Provide DB connection details in project settings/detail: service-internal observed
host/address and port, database/user, and separately the optional host-loopback
connection. Inject existing DB_* environment keys into PHP as today; do not assume
frameworks prefer them to their own config. Surface a mismatch as application
configuration guidance, not permission to edit the app. Secret reveal requires an
explicit local interactive action and must not enter general JSON/support output.

Three distinct observations:

1. Infrastructure: required processes plus routed StageServe probe are ready.
   A fixed reserved exact path `/__stageserve_health` routes to a read-only
   StageServe-owned probe outside the project mount. A nonce response proves the
   actual PHP request path, not static nginx config. Responses expose no secrets.
2. Managed DB connectivity: execute a read-only `SELECT 1` inside the managed
   service using a protected credential channel (not shell argv/logs); required
   for this stack. No application schema or row writes. Container CLI exec status
   alone is not the HTTP/PHP proof.
3. Application health: optional configured relative HTTP path and accepted status
   set, default unset. Bounded GET, no cross-origin redirects, no credential logging.
   When unset report `not_checked`; when failing report application-unhealthy while
   infrastructure stays running. Never roll back healthy services for app code 500,
   login redirects, missing application DB tables or an empty web folder.

Each request timeout 5 s; combined infrastructure deadline 120 s default, as
already exposed by wait-timeout. Optional app probe uses a separate maximum 10 s
budget and cannot extend infrastructure deadline. Probe payload is mounted from
versioned assets, never created in docroot; teardown removes only generated probe
config. Test fixtures, not real app readiness, own sentinel read/write tests.

## Connectivity policy and candidate topology

Threat boundary: prevent direct cross-project private-service access and accidental
cross-project operations. This is not a sandbox for hostile code against the Mac,
the shared gateway, privileged operator, or intentionally accessible local websites.
Do not claim malicious-tenant isolation. Host access to container addresses is part
of Apple's runtime model. Local website access via gateway is intentionally shared;
private database/exec/admin endpoints must not be reachable from another project.

Candidate: private application network per UUID (PHP+DB), separate ingress network
per UUID (project nginx), with project nginx also attached to its own application
network for FastCGI. Shared gateway attaches to each ingress network, never private
application networks. Inspect explicit addresses per interface; do not use bare DNS
service names or the first arbitrary inspect address. Disable forwarding between
interfaces. Published gateway ports bind loopback only. DB host port and debug
publication are opt-in; neither may create a reachable cross-project bypass.
All wildcard host binds are rejected in v1. Do not attach every project to one
shared network. Candidate needs live multi-network proof before adoption.

| Source -> destination | Policy |
|---|---|
| Host -> gateway HTTP/TLS loopback | Allow |
| Host -> own project container IP | Runtime permits; disclose local host trust boundary |
| Gateway -> project nginx ingress:80 | Allow |
| Gateway -> any DB/PHP private network | Deny direct connection |
| Project nginx -> same project PHP FastCGI | Allow |
| PHP/debug -> same project DB:3306 | Allow |
| Project A -> project B private IPs/DB/admin or host-published DB port | Deny |
| Project A -> B public website through gateway | Allow as ordinary local site traffic; never mistake this for DB isolation |
| LAN -> any published StageServe service | Deny |

The matrix defines cross-project/private-service isolation, not zero-trust separation
between trusted services within the same project.

IPv4 and IPv6 negative tests are mandatory; if a family cannot be safely isolated,
disable it explicitly for StageServe networks/endpoints and report that supported
limitation rather than leave an untested bypass. Also test host-address aliases,
not just container IPs. No guest route manipulation or host security weakening to
force a pass. If dynamic shared-gateway network attachment requires recreation,
G0 must prove preservation of existing sites; otherwise block and evaluate a
host-side gateway design in a separate recorded architecture decision. Do not
silently weaken the connectivity policy to keep the old gateway implementation.

DNS decision for candidate: retain existing dnsmasq adapter for host project names
pointing to loopback. Use inspected addresses internally. Apple-native DNS is an
explicit comparison experiment, not a simultaneous active provider or mandatory
migration; it must handle project aliases, configured suffixes, ownership and avoid
restarting unrelated Apple workloads before it can replace the host DNS path.

## Mixed HTTP/TLS routing

TLS is per route. Each route stores scheme, redirect flag, certificate reference
and project UUID. Shared listeners bind 127.0.0.1:80 and 127.0.0.1:443; an existing
installation may preserve a validated explicit alternate HTTPS port (e.g. 8443).
No per-project operation silently changes installation listener ports. Conflict
fails before mutation with an explicit installation-level preview as remediation.
`.test` defaults HTTP; `.dev` requires HTTPS; existing suffixes are retained.
HTTP-only routes do not redirect because a neighbour enables TLS. An HTTP-only
hostname on the TLS listener is rejected rather than accidentally serving another
project; TLS routes use exact SNI/host matching. HTTP routes always use their own
upstreams. Local trust is explicit; no certificate-error bypass instructions.

Generate one certificate per TLS hostname into protected StageServe-owned storage,
using the existing TLS provider; shared CA trust is installation scope. Pre-generate
and validate replacements before reload, including SAN/expiry/key match. Renewal
is checked on startup/status; warn within 30 days of expiry and offer explicit
renewal action. Expired/untrusted TLS is unavailable with recovery guidance, never
silently downgraded. Failure leaves other certificates/routes intact. Detach removes
only the route; cached cert may be retained with ledger ownership. Uninstall removes
only demonstrably owned leaf material; never removes a shared mkcert CA trust entry
automatically. Host-change rollback uses backups/hashes and does not overwrite
externally edited resolver or trust configuration.

## Required cases and task ownership

| Cases | Proof | Tasks |
|---|---|---|
| ID-01–04 | detach/fresh reattach; copied/renamed folder; lost ledger/settings; crash between create and save with owner mismatch | T040, T013, T019 |
| CFG-01–04 | running and stopped edit; changed origin equal to default; failed recreate/reload; concurrent/stale apply, secrets redacted | T004, T041, T025 |
| APP-01–04 | PHP probe + SELECT 1; empty docroot/redirect/500 stays running; .env precedence; no source/DB writes by readiness | T008, T009, T011 |
| NET-01–04 | complete allowed/denied matrix; IPv6/host-alias bypass; restart addresses; gateway network change preserves two sites | T003, T014, T016 |
| TLS-01–04 | mixed HTTP/HTTPS in both start/stop orders; differing suffixes; refused trust/renewal failure; port conflict and cert removal | T014, T016, T020 |
