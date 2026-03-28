# 12. CI/Build Customizations

**Approach:** Don't try to port CI configs -- rebuild them for the new version based on what's needed.

## What was implemented (v12.4)

Three custom GitHub Actions workflows and a Dockerfile modification.

### Workflows

All workflows trigger on pushes to `*-intergral` branches and PRs targeting them. They use `concurrency` groups with `cancel-in-progress` to avoid duplicate runs.

#### 1. Backend Tests (`.github/workflows/intergral-backend-tests.yml`)

- **4 sharded Go test jobs** running in parallel on `ubuntu-latest`
- Uses the existing `scripts/ci/backend-tests/shard.sh` utility to distribute packages across shards
- Each shard runs: `CGO_ENABLED=0 go test -short -timeout=30m`
- Uses `actions/setup-go` with built-in Go module caching (keyed on `go.sum`)
- A `required-backend-tests` rollup job checks all shards passed (same pattern as upstream)
- 4 shards is sufficient for OSS-only (upstream uses 8 because enterprise adds more packages)

#### 2. Frontend Tests (`.github/workflows/intergral-frontend-tests.yml`)

- **4 sharded JS test jobs** running in parallel on `ubuntu-latest`
- Uses the existing `.github/actions/setup-node` composite action for Node.js + yarn caching
- Runs `yarn install --immutable` with `PUPPETEER_SKIP_DOWNLOAD` and `CYPRESS_INSTALL_BINARY=0`
- Each shard runs: `yarn run test:ci` with `TEST_SHARD` / `TEST_SHARD_TOTAL` env vars
- `TEST_MAX_WORKERS=2` (tuned for standard GHA runners; upstream uses 4 on large runners with 16 shards)
- Same rollup job pattern as backend tests

#### 3. Docker Build & Push (`.github/workflows/intergral-docker.yml`)

- Builds and pushes to `intergral/grafana` Docker registry
- Uses `docker/build-push-action` with `docker/setup-buildx-action`
- **BuildKit automatically parallelizes** the independent `js-builder` and `go-builder` Dockerfile stages -- no need to split into separate GHA jobs
- **GHA layer caching** (`cache-from: type=gha` / `cache-to: type=gha,mode=max`) caches the `go mod download` and `yarn install` layers across builds
- Supports **dev builds** via `workflow_dispatch` with a `go_build_dev` input -- when set to `dev`, passes `JS_NODE_ENV=dev`, `JS_YARN_BUILD_FLAG=dev`, and `GO_BUILD_DEV=dev` as build args
- Tags images with branch name and SHA (e.g., `12.4.x-intergral`, `12.4.x-intergral-abc1234`)
- Requires `DOCKER_USERNAME` and `DOCKER_PASSWORD` repository secrets

### Dockerfile modification

- Added `ARG GO_BUILD_DEV=""` to the `go-builder` stage (after `WIRE_TAGS`)
- Passes it through to the build: `make build-go GO_BUILD_TAGS=${GO_BUILD_TAGS} WIRE_TAGS=${WIRE_TAGS} GO_BUILD_DEV=${GO_BUILD_DEV}`
- This wires the dev build flag from the workflow through to the Makefile, which translates `GO_BUILD_DEV=dev` into `GO_BUILD_FLAGS += -dev`

### Upstream workflow suppression

Most upstream workflows already skip on forks because they check `github.repository == 'grafana/grafana'`. Push triggers only match `main` and `release-*.*.*`, which don't match `*-intergral` branches. Some bare `pull_request:` triggers may still fire for PRs but they detect no relevant changes and exit quickly. A future improvement would be adding `branches-ignore: ['*-intergral']` to those triggers.

## What's still needed

- Release workflow (tag-based trigger, build production artifacts, push final images)
- Flaky tests may need to be skipped (these change with each version)
- Test infrastructure may need startup/shutdown timeouts adjusted
- Consider increasing shard counts if CI runners are upgraded