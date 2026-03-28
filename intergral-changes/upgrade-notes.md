# Upgrade Notes: v12.3 to v12.4

The following conflicts and issues were encountered when cherry-picking Intergral commits from `12.3.x-intergral` onto Grafana v12.4.2. These notes should help with future upgrades.

## Auth Proxy (Section 2)
- **user_sync.go**: `userService.GetSignedInUser()` changed from accepting a `GetSignedInUserQuery` struct to taking `(ctx, userID, orgID)` directly. The org priority logic had to be adapted to the new signature.

## Ad-hoc Filter UID Resolution (Section 3)
- **transformSaveModelSchemaV2ToScene.ts**: v12.4 changed from directly returning `new AdHocFiltersVariable({...})` to building an `adhocVariableState` object first (to conditionally add `allowCustomValue`). The UID resolution code had to be placed before the state object rather than inline in the constructor.
- **loki/datasource.ts**: Had minor whitespace conflicts and an unused `exprAfterTemplateVars` variable from the v12.3 version -- dropped it since `exprWithAdHoc` duplicates the same logic.

## OpsPilot Integration (Section 9)
- **SpanDetail/index.tsx**: v12.4 moved `linksComponent` into the header section (previously in a separate div). The `OpsPilotSpanButton` was placed directly after the header instead of in the old link list div.
- **TracePageHeader.tsx**: v12.4 completely restructured the trace page header with new filter components (`TracePageSearchBar`, `SpanGraph`, `TraceFilterPills`, `useTraceAdHocFiltersController`). The `OpsPilotTraceButton` had to be added into the new `{!hideHeaderDetails && (...)}` action area rather than the old flat layout.
- **AppChrome.tsx**: This file required conflict resolution across 4 separate cherry-picks due to accumulated changes. The v12.4 version added `floatingUtils.BOUNDARY_ELEMENT_ID`, changed header heights, and added `ExtensionSidebar` support. Each OpsPilot-related commit (BroadcastProvider wrapper, iframe nav, metadata hook, mega menu) touched this file.

## No Mega-Menu (Section 8)
- **AppChrome.tsx**: The `useMediaQueryMinWidth` import had to be removed since the responsive hook was replaced with a no-op. The `DOCKED_LOCAL_STORAGE_KEY` export moved around in v12.4.

## Alerts from Graphs (Section 7)
- **PanelMenuBehavior.tsx**: v12.4 added several new imports (`LocalValueVariable`, `DataQuery`, `OptionsWithLegend`, `appEvents`). The conflict was import-only -- the menu item logic applied cleanly.

## Dashboard Permissions Refresh (Section 4)
- **Most conflict-prone feature.** v12.4 had 91 commits touching the dashboard scene area.
- **apiserver/client.ts**: `BackendSrvRequest` import added, and `contextSrv` import path changed from `app/core/core` (v12.3) to `app/core/services/context_srv` (v12.4). The old path no longer exists.
- **dashboard/api/v2.ts**: The `subresource()` call gained additional parameters in v12.4. The `{ showErrorAlert: false }` option needed to be passed as the 4th argument.
- **DashboardPageProxy.test.tsx**: Test structure changed -- the Intergral version added a test case that conflicted with v12.4's restructured test suite. Kept v12.4's test structure.
- **.github/workflows/docker_build.yml**: Deleted in v12.4 base (CI workflows will be rebuilt from scratch).

## HA Alerting Partitioning (Section 1)
- **ngalert.go**: v12.4 uses `ng.schedCfg` (struct field) instead of local `schedCfg` variable. The partitioner creation code was placed before the struct assignment, and references updated accordingly.
- **setting_unified_alerting.go**: v12.4 added `AlertmanagerMaxTemplateOutputSize`, `BacktestingMaxEvaluations`, and `IgnorePendingForNoDataAndError` fields. The HA partitioning fields were appended after these.
- **state/manager.go**: Same pattern -- v12.4 added `ignorePendingForNoDataAndError` field. HA fields appended after it. The Manager constructor needed both sets of fields.
- **api/prometheus/api_prometheus.go**: `RuleStatusMutatorGenerator` gained a `ctx context.Context` parameter in v12.4's closure signature and `statusReader.Status()` now takes `(ctx, key)`. The Intergral variadic `stateManager` parameter had to be combined with the new context-aware signature.