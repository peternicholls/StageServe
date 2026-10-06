---
applyTo: "README.md,docs/**,specs/012-apple-only-experience/**,.env.stageserve.example"
---

Constitution 3.1.0 and specs/012-apple-only-experience define the Apple-only target.
Product direction and roadmap summarize it; docs/design defines presentation.
Historical implementation descriptions are not proof of target behaviour.

Keep affected operator documentation, command help, active spec contracts and
acceptance cases aligned in the same change. Distinguish observed implementation
from planned requirements; do not weaken a target merely to match incomplete code.

Canonical configuration is project and stack .env.stageserve. The documented
project application .env DB fallback is read-only and subordinate to explicit
StageServe values. Generated env files are outputs. Apple manifests are
stacks/20i/apple-container.20i.json and apple-container.shared.json.

archive/** and previous-version-archive/** are historical evidence only. Never
apply active-contract updates to archived files. Run the planning verifier when
changing the active spec or archive inventory.
