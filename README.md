# Patches

This repository is a monorepo of mirrored Terraform providers, automatically
synchronized from upstream sources (e.g. HashiCorp) and augmented with local
patch stacks. Each mirrored provider exposes a consumable Go module and tracks
upstream releases on a per-major-version basis to enable multiple LTS branches.

High-level behavior:
* For each provider and major version:
    * Poll upstream repositories daily for new tags via GitHub actions.
    * Mirror the upstream source into this repository under `/mirrors/<provider>/<major>`.
    * Apply transformations and patches from `/providers/<provider>/<major>/patches` on top of the mirrored code.
    * Run go mod tidy.
    * Open a pull request with auto-merge enabled.
    * Tag the repository with a namespaced tag matching the upstream version after the PR merges.

If patches fail to apply cleanly, the process stops and requires human (or AI)
intervention.
