# Intergral Fork - Custom Changes Implementation Guide

Last reviewed: 2026-03-27 against intergral `12.4.x-intergral` branch (ported from `12.3.x-intergral` to Grafana v12.4.2)

This directory describes every custom feature in the intergral Grafana fork (FusionReactor Cloud). Each file describes *what* and *why*, not specific patches, so they can be reimplemented against any future upstream version.

## Changes

| # | Feature | File |
|---|---------|------|
| - | [Upgrade Notes: v12.3 to v12.4](upgrade-notes.md) | `upgrade-notes.md` |
| 1 | [HA Alerting Partitioning](01-ha-alerting-partitioning.md) | `01-ha-alerting-partitioning.md` |
| 2 | [Auth Proxy Org Context](02-auth-proxy-org-context.md) | `02-auth-proxy-org-context.md` |
| 3 | [Ad-hoc Filter UID Resolution](03-adhoc-filter-uid-resolution.md) | `03-adhoc-filter-uid-resolution.md` |
| 4 | [Dashboard Permissions Refresh](04-dashboard-permissions-refresh.md) | `04-dashboard-permissions-refresh.md` |
| 5 | [Alert Email Branding](05-alert-email-branding.md) | `05-alert-email-branding.md` |
| 6 | [Trace Link Sub-URL Fix](06-trace-link-suburl-fix.md) | `06-trace-link-suburl-fix.md` |
| 7 | [Alerts from Graphs](07-alerts-from-graphs.md) | `07-alerts-from-graphs.md` |
| 8 | [No Mega-Menu](08-no-mega-menu.md) | `08-no-mega-menu.md` |
| 9 | [OpsPilot Integration](09-opspilot-integration.md) | `09-opspilot-integration.md` |
| 10 | [Navigation Tree Trimming](10-navigation-tree-trimming.md) | `10-navigation-tree-trimming.md` |
| 11 | [Tempo Search Nil Panic Fix](11-tempo-search-nil-panic.md) | `11-tempo-search-nil-panic.md` |
| 12 | [CI/Build Customizations](12-ci-build-customizations.md) | `12-ci-build-customizations.md` |
| 13 | [API-Provisioned Read-Only Datasources](13-api-readonly-datasources.md) | `13-api-readonly-datasources.md` |
| 14 | [Alert Notification External URL Override](14-alert-external-url-override.md) | `14-alert-external-url-override.md` |
| 15 | [Miscellaneous](15-miscellaneous.md) | `15-miscellaneous.md` |

## Testing

After making changes, always run the existing tests for any modified files. Most modified files have corresponding test files (`*_test.go` for Go, `*.test.tsx`/`*.test.ts` for frontend). Intergral changes can break upstream tests in subtle ways -- for example, adding a `{shareButton}` element to a React component's JSX shifts the `children` indices that tests rely on, or changing a constant like `MAX_LINKS` can alter which rendering branch is exercised. Run the relevant tests before committing to catch these issues early.

## Items NOT Ported (from v12.3 to v12.4)

- **Docker plugin extraction** (Section 12 area): The revert commit (#54) indicated this approach was abandoned. Skipped.
- **Navigation tree trimming** (Section 10): Only the alerting visibility fix was ported. The broader nav trimming (`&& false` guards) was not -- upstream v12.4 may have already removed some of these sections or made them configurable.
- **Tempo search nil panic** (Section 11): Not checked/ported -- likely fixed upstream.
- **Command palette simplification** (Section 13): Not ported -- needs to be re-evaluated against v12.4's command palette.
- **CI/Build workflows** (Section 12): Rebuilt for v12.4 -- backend tests, frontend tests, and Docker build/push workflows are in place. Release workflow still needed.