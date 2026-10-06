# Historical StageServe planning — archived 2026-09-11

Specs 002–011 are preserved byte-for-byte under `specs/`. Their original checkboxes, technical assumptions and relative links are historical evidence, not current instructions. Relative links in source documents may describe the old repository layout; use the manifest's original/archived mapping to locate sources. No missing prototype has been reconstructed.

Current authority: [product direction](../../docs/product-direction.md), [roadmap](../../docs/roadmap.md), [spec 012](../../specs/012-apple-only-experience/spec.md).
Lessons: [review matrix](../../specs/012-apple-only-experience/legacy-lessons.md).
Provenance: `manifest.json` records every original path, archived path and SHA-256.

Restore an individual file to a temporary location for comparison; do not make this archive an active backlog or overwrite current work. Git history also retains original paths. No source/runtime data was deleted by this archival operation.

## Reproducible inventory (2026-09-13 correction)

`manifest.json` verifies 88 repository-content files and is the clean-checkout
acceptance inventory. `filesystem-manifest.json` preserves the original 90-file
snapshot. Its two `.DS_Store` entries are listed in `local-metadata-manifest.json`:
Finder metadata remains locally preserved but is deliberately not required from
Git. No historical source bytes or hashes changed. The former statement "90
files verified" was a local snapshot result, not a clean-checkout guarantee.

Run `python3 scripts/verify-planning.py` from the repository root. Metadata files
are neither deleted nor force-added. Restoring a document uses its original path
mapping; never overwrite current implementation or reactivate the archive.
