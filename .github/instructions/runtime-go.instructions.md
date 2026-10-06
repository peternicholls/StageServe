---
applyTo: "cmd/**,core/**,infra/**,observability/**,platform/**"
---

Follow constitution 3.1.0 and specs/012-apple-only-experience/contracts/.
Apple container is the only supported target runtime; Docker/Compose fallback
and Engine sockets are unsupported. Existing legacy adapter code is retirement
work, not the architecture to preserve. OCI Dockerfiles may remain build inputs.

Ownership:
- core/config: effective values and source provenance
- core/state: project identity, retained-data ledger and atomic operation journal
- core/lifecycle: locks, transactional apply, recovery and shared routing
- infra/applecontainer: version-qualified CLI commands and observations
- infra/gateway: per-route TLS and validated configuration
- core/guidance and command adapters: shared action semantics and presentation

Use exact project/resource ownership, never name prefixes, to authorize changes.
Never implicitly adopt legacy records as Apple resources. Preserve data on stop
and detach. Test current contracts, including failure and recovery, before
claiming support. Unit success does not prove live Apple compatibility.

Use focused tests first, then required broad gates. Do not add dependencies or
reintroduce legacy configuration aliases without an explicit requirement.
