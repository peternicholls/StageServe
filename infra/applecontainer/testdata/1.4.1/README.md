# Apple container 1.4.1 CLI evidence

Captured on 2026-10-06 from the signed arm64 installer published at
https://github.com/apple/container/releases/tag/1.4.1. The package SHA-256 is
`c0d2716afefbb194c93fae662e9cae7cc186bcbcf746816608ec673dd648a6a4`;
`pkgutil --check-signature` verified Apple's developer certificate and trusted
notarization. `pkgutil --expand-full` extracted the binary without installation.

`capture.json` records exact arguments, exit codes and output hashes. Help and
version commands passed. System status returned exit 1 and `unregistered`.
Inventory commands failed because no runtime service was registered; their
machine-specific errors are kept outside the repository. No runtime was started.

The adapter's status discriminator follows the pinned implementation:
https://github.com/apple/container/blob/1.4.1/Sources/ContainerCommands/System/SystemStatus.swift.
Only `running` means ready. Synthetic running/invalid payloads in unit tests are
parser coverage, not captured runtime evidence.

T002 remains incomplete: real inspect/run/network/volume inventory fixtures and
the tested host matrix are still required. T003 requires live connectivity tests.
These files certify neither the parser for running containers nor the candidate
network topology.

Synthetic parser coverage also follows the pinned `ManagedContainer` encoder,
which emits `id`, `configuration` (including raw labels) and nested `status.state`:
https://github.com/apple/container/blob/1.4.1/Sources/ContainerResource/Container/ManagedContainer.swift.
Project inventory matches exact observed `io.stageserve.project` labels rather
than name prefixes. This preserves labels needed for future UUID ownership
validation; it does not yet establish installation/project UUID ownership.
An arbitrary network interface is never selected as the routing endpoint by
these synthetic tests.
