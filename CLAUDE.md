# Intergral Grafana Fork

This is the **Intergral fork** of Grafana, used by **FusionReactor Cloud**. It tracks upstream Grafana releases and adds custom features on top.

## Branch conventions

- `main` — upstream Grafana main (sync target)
- `12.4.x-intergral` — current active branch, based on Grafana v12.4.2 with Intergral customizations
- Pattern: `<upstream-version>-intergral`

## Intergral customizations

All custom changes are documented in `intergral-changes/`. Each numbered markdown file describes **what** and **why** for a single feature, written so the feature can be reimplemented against any future upstream version. Always read the relevant change doc before modifying Intergral-specific code.

Key custom features:
- **HA alerting partitioning** — distributes alert rule evaluation across cluster peers
- **OpsPilot integration** — cross-tab/iframe communication with the OpsPilot AI assistant (`public/app/intergral/`)
- **API-provisioned read-only datasources** — admin bypass for read-only guards
- **Alert notification external URL override** — for iframe embedding scenarios
- **Auth proxy org context** — org assignment via HTTP header
- **Dashboard permissions refresh** — fixes 403 after saving new dashboards
- **Trace link sub-URL fix** — prevents double-prefix in iframe deployments

See `intergral-changes/README.md` for the full index and porting status.

## Project structure (Intergral-specific)

```
intergral-changes/          # Change documentation (implementation guides)
public/app/intergral/       # OpsPilot frontend components and hooks
.github/workflows/          # intergral-* CI workflows (backend tests, frontend tests, docker)
```

## Testing

```bash
# Frontend — all Intergral tests
npx jest --no-cache public/app/intergral/

# Backend — Intergral settings tests
go test ./pkg/setting/ -run "TestIntergral" -v

# Backend — HA partitioner tests
go test ./pkg/services/ngalert/schedule/ -run "TestPartitioner" -v

# Full frontend test suite (sharded in CI)
yarn test

# Full backend test suite
go test ./...
```

Intergral changes can break upstream tests in subtle ways (e.g. shifting React children indices, changing constants that affect rendering branches). Always run relevant upstream tests after modifying Intergral code.

## Conventions

- When adding a new Intergral feature, create a numbered doc in `intergral-changes/` following the existing format
- Prefix commit messages with `fix(intergral/...)`, `feat(intergral/...)`, or `docs(intergral/...)` to distinguish from upstream
- Keep Intergral code isolated where possible (e.g. `public/app/intergral/` for frontend, config keys in `[unified_alerting]` for backend) to minimize merge conflicts on upstream upgrades
- The `Dockerfile` has a custom `GO_BUILD_DEV` ARG — preserve this when upgrading
- CI workflows are in `.github/workflows/intergral-*.yml` — these are independent of upstream CI
