# Terratest template

Provisioned from [`Qode-Fleet-Control/fleet-template-v1`](https://github.com/Qode-Fleet-Control/fleet-template-v1) — the fleet
lifecycle contract (`bin/`, `fleet.conf`, `compose.yaml`, deploy workflows) with a Terraform module and its [Terratest](https://terratest.gruntwork.io) suite
laid on top.

**This repo is a job, not a service.** Its container runs the Go test suite, which
**really applies** the module with terraform, checks what it created, and destroys it —
then exits, 0 when every test passes. The module uses only credential-free providers
(`random`, `local`), so no cloud account is involved. Nothing listens on `$PORT`.

## What is in it

| path | |
|---|---|
| `modules/pet-file/` | the module under test: a `random_pet` name and a `local_file` greeting; `pet_length` is validated (1-5); lock file committed |
| `test/pet_file_test.go` | `TestPetFileApply`: init + apply, outputs checked, file content checked, a second plan must be empty (`-detailed-exitcode` 0), destroy, file gone. `TestPetFileRejectsBadLength`: plan with `pet_length = 9` must fail with the validation message |
| `test/go.mod`, `test/go.sum` | Terratest v2 (`modules/terraform/v2`, `modules/core/v2`), testify |

Each test copies the module to a temp dir (`files.CopyTerraformFolderToTemp`) so tests run
in parallel without sharing `.terraform` or state.

## Run it

**On the fleet:** `bin/run` builds the image (`docker compose build`) and stops there —
`DOCKER_START_CMD` is empty because there is no server. Run the job with
`docker compose run --rm app`.

**With docker:**

    docker compose build
    docker compose run --rm app        # exit 0 = PASS for every test

**Without docker** (needs Go >= 1.26 and `terraform` >= 1.10 on `PATH`):

    cd test
    go test -v -count=1 -timeout 10m ./...

`FLEET_RUNTIME=process bin/run` runs `INSTALL_CMD` (`go -C test mod download`) and
`BUILD_CMD` (`go -C test vet ./...`) and then stops at the start step, by design.

## Origin

    hand-written — Terratest ships no project generator
    cd test && go mod init github.com/Qode-Fleet-Control/qode-terratest-template-v1/test
    go get github.com/gruntwork-io/terratest/modules/terraform/v2@v2.0.0 github.com/gruntwork-io/terratest/modules/core/v2@v2.0.0
    go mod edit -go=1.26.0 && go mod tidy

Layout as the Terratest quick start teaches: Terraform code in `modules/`, Go tests in
`test/` with their own `go.mod`. Uses Terratest **v2** (released 2026-09-23): per-module
`/v2` import paths and the `...Context(t, ctx, ...)` helpers.

## Deviations, and why

- `Dockerfile` is two-stage: `golang:1.27-alpine` vets the suite and compiles it to a test
  binary (`go test -c`); the runtime stage is `hashicorp/terraform:1.16.5` (the tests shell
  out to `terraform`) running that binary with `-test.v`. No Go toolchain ships in the
  final image. Runs as non-root `app` (uid 10001).
- The providers are placed in a plugin cache (`TF_PLUGIN_CACHE_DIR`) at build time from the
  module's lock file, so each test's `terraform init` links them instead of downloading.
- `go 1.26.0` in `go.mod` (the minimum Terratest v2 needs), not the 1.27 toolchain that
  created it, so older Go installs can run the suite.

## Verified

**The docker job has NOT been verified yet.** On 2026-10-05 the build host's docker disk
stayed below the 6 GB floor (0-3 GB free) for over three hours, so `docker compose build`
was never run for this repo. Build and run it once before trusting it:

    docker compose build && docker compose run --rm app; docker compose down --rmi local -v

What WAS checked, with the real CLIs outside docker (same `scripts/check.sh` the image runs):

    go 1.27.1 + terraform 1.16.5:
    cd test && go vet ./... && go test -count=1 -timeout 10m ./...   # ok (both tests PASS, ~7s)

## Serving over HTTP

There is no HTTP surface. If you add one, listen on `0.0.0.0:$PORT`, serve at `/`, set
`PORT`, `HEALTH_PATH`, `START_CMD` and `DOCKER_START_CMD` in `fleet.conf`, and publish
`"${PORT}:${PORT}"` in `compose.yaml`. See `docs/fleet-lifecycle.md`.
