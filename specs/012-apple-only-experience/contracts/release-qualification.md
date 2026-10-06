# Release qualification, compatibility and host lifecycle

Revised 2026-09-13. Normative target for FR-004, FR-012 and FR-019–021.
Qualification is open until actual evidence exists; no current release is certified.

## Compatibility record and supported candidate

Initial validation candidate: Apple container 1.4.1, darwin/arm64, macOS 26 and 27
on real Apple silicon hardware. Each OS version is a separate test row, not an
inferred support promise. Only rows that pass can be advertised; unsupported rows
must fail readiness before mutation. Go remains 1.26/toolchain 1.26.2 until a
separate qualified toolchain change. Minimum hardware target: 16 GiB RAM and
30 GiB free disk before initial images; workload is two representative PHP sites.
This minimum is a proposed qualification target, not measured compatibility.

A release compatibility record contains: candidate revision; binary and installer
SHA-256; asset tarball SHA-256; manifest schema; read/write state-schema ranges;
Apple CLI/OS/arch rows; every OCI image index and linux/arm64 platform digest;
PHP/MariaDB versions; probe version; prior compatible releases; signing status;
and acceptance evidence digests. Never use floating image tags to recreate a
known project. Record each volume's engine digest/version and last writer release
in the retained ledger. PHP image cache keys include build inputs, PHP version,
platform and base-image digest, not one shared unversioned stageserve-apache tag.

## Update and rollback

V1 supports N -> N+1 and rollback to N only when explicitly qualified together.
State writes during that pair must remain readable by N. Breaking schema changes
are not ordinary updates: refuse until an explicit migration with backup/restore
and user-visible downtime exists. Never restore an old state snapshot over new
application writes silently. Safe rollback retains latest compatible ledger/state
and database data, switches binary/assets to N, and reobserves actual services.

Existing volumes remain on their recorded engine digest across StageServe updates.
Any DB engine digest/version change is an explicit migration, even a patch: first
prove compatibility on a copied/exported fixture, and execute into a new named
volume with the source preserved. No automatic in-place database engine update.
The v1 routine update UI refuses such requests and presents the documented logical
export/import procedure. The release must not keep an insecure engine silently:
flag its status, provide the migration path, and qualify replacements before use.
New projects use the release's approved digest.

Migration: quiesce source writes explicitly; export with documented credentials;
record schema/row counts and application sentinel; import into a new target volume;
verify counts, integrity and app acceptance; switch only after explicit preview.
Keep original export/source until operator disposes of them. If target has received
new writes, rollback requires a new export/reconciliation or explicit acknowledgement
of the exact data-loss window; refuse a blind restore. Interrupted migration leaves
both sources identifiable in the retained ledger. Source remains read-only during
cutover. Docker-to-Apple migration uses a user-supplied verified logical export;
StageServe never invokes Docker as a supported backend or globally prunes resources.

Ordinary update stages a complete binary/assets pair in a versioned directory and
atomically switches a single active pointer after verification. A journal covers
any additional state transitions. Installer/update cancellation never replaces just
one member of the pair. Project settings/source are never overwritten. Update with
running projects either postpones or performs a clearly previewed maintenance stop;
no background recreation. All projects and retained detached data are inventoried.

## Candidate -> qualify -> promote

T045 replaces the current tag-immediately-publishes workflow:

1. Build candidate artifacts for darwin/arm64 only. Include real binary, matching
   runtime bundle (all referenced templates/probes/build inputs), installer and
   checksums. Build once from a recorded revision. Store candidate privately or as
   a draft release; creating a tag must not publish an unqualified release.
2. Produce immutable digest inventory and scan bundle entries for absolute/traversal
   paths and missing assets. Download/checksum must fail closed; a missing checksum
   tool or missing digest is not a reason to skip verification.
3. Install the exact candidate through the production download/verification/extract
   code path on a clean test account, using an explicit candidate repository/tag.
   No STAGESERVE_TEST_ASSET_PATH/BUNDLE_PATH bypass, `/usr/bin/true` stand-in or
   manually preinstalled runtime assets count. Test source override only selects
   the artifact endpoint; it must not bypass checks, runtime or asset validation.
4. Outside the checkout, run real installed `stage` through create/up/status/logs/
   stop/reattach and PHP/DB/browser checks. Exercise first download, missing bundle,
   wrong binary/bundle combination, checksum mismatch, interrupted install, and
   package paths containing spaces. Ensure unsupported platforms refuse clearly.
5. Qualify N -> N+1 -> write new DB sentinel/change settings -> N rollback, proving
   new writes and compatible state survive. Also test deliberate newer schema and
   DB-engine incompatibility: refusal with recovery, no volume mutation.
6. Gate promotion on successful automated checks and a protected evidence review
   recording exact digests and completed G0/GUX/G1/G2/G3/G4 cases. Apple hardware
   jobs use an available dedicated/self-hosted runner or a recorded supervised
   run; Ubuntu cross-build is not live Apple acceptance. Do not expose runner
   secrets to untrusted pull-request execution.
7. Promote the same qualified artifacts without rebuilding. Verify public download
   digests equal the qualified inventory. Candidate expiry, changed digest or
   missing evidence blocks promotion. Release notes publish actual tested rows,
   upgrade restrictions and limitations. Emergency fixes still require candidate
   qualification; never silently bypass the gate.

Signing decision: checksum integrity is required; it is not a code-signature claim.
Use signed/notarized distribution if signing credentials are available and qualified;
otherwise label binary signing status accurately and test the real download's macOS
launch behaviour. If the intended install route is blocked, G4 stays open; never
instruct bypassing Gatekeeper/security warnings or falsely claim signed binaries.

## Offline and capacity policy

Offline operation includes a fresh StageServe process, stopped start, and recreation
of containers from already cached exact images with unchanged qualified build inputs.
No image rebuild or registry contact is required on that path. Cached binary/assets,
probe, image digests and compatible volume state are prerequisites. A missing image
or changed PHP build input fails before stopping a working project and lists exactly
what needs an online preparation step. Offline initial provisioning is not promised.
DNS health checks are local and must not depend on public internet reachability.

Proposed budgets to measure and either meet or explicitly amend before G1:

| Metric | Target and method |
|---|---|
| Cold cached two-project start | <=120 s total startup deadline; report host, image digests, dataset and per-service timing |
| Runtime incremental memory | <=8 GiB sustained RSS/VM memory over host idle for two idle sites; measure 10 min idle and PHP/DB workload, peak separately |
| Idle CPU | <=10% of one core averaged over 5 min after warmup; no busy polling |
| Initial free disk | >=30 GiB; estimate images/build/volume growth and reject start if available headroom below estimate + 5 GiB |
| Disk failure | ENOSPC during write/reload leaves previous valid state/route and named recovery; no destructive retry |
| Managed log growth | 10 MiB per service, 3 retained rotations; log loss/truncation observable; no whole-runtime global cleanup |

Per-service memory/CPU limits are selected from measured profiles during T042;
record them in the released manifest, not untested guesses in runtime code. No
resource claim can pass on a larger Mac alone. If targets fail, optimize or amend
minimum hardware with evidence and adjust estimates before G1; do not waive a gate.

## Reboot, sleep and uninstall

No StageServe auto-start/auto-restart-at-login in v1. After reboot or Apple service
restart, observe desired and actual state, refresh addresses and offer explicit
resume. Never claim a project remains running solely from saved intent. Sleep/wake
rechecks addresses, DNS/TLS and health before enabling mutations; loss of runtime
produces unknown/unavailable state. Reconnect logs with visible boundary, no silent
replay duplication. Test sleep during start and at steady state.

Uninstall is a documented supervised process in v1, not an invented command. Show
inventory and stop only owned projects/gateway. Remove the installed binary and
versioned assets after inventory/export. Preserve config, retained ledger and DB
volumes by default. Never stop/delete other Apple workloads, uninstall Apple itself,
remove unrelated dnsmasq configuration, or revoke a shared mkcert CA. Owned resolver
files may be removed only if current hash matches the recorded managed version,
with explicit host-change authorization; otherwise leave and explain. Provide
reinstall/recovery steps and a separately scoped retained-data removal procedure.
Host changes keep original bytes, ownership marker and expected hash in the ledger.

## Qualification cases

| Cases | Required evidence | Tasks |
|---|---|---|
| REL-01–04 | exact candidate install; checksum/missing-bundle failure; unsupported platform; same digest promotion gating | T032, T045 |
| UPD-01–04 | interrupted pair switch; rollback after new writes; incompatible schema refusal; DB-engine change/export restore on separate volume | T031, T032, T033 |
| OFF-01–03 | disconnected cached fresh start/recreation; changed input/missing cache refusal before downtime; local DNS offline | T042 |
| CAP-01–03 | minimum host CPU/RAM/startup/disk measurements; ENOSPC rollback; bounded logs | T042, T026 |
| HOST-01–03 | sleep/wake; reboot/runtime restart with explicit resume; uninstall/reinstall preserving retained data and unrelated host resources | T044 |

Every case must have expected/observed result, host/versions and artifact digests.
A missing host, signing credential, participant or runtime is an open qualification
gate; it is not evidence that the planning contract is ambiguous or the product works.
