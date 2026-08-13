# Claude Memory for Intergral Grafana Fork

## Project Context
This is Intergral's fork of Grafana (FusionReactor Cloud / OpsPilot) with custom features
for OpsPilot integration, enterprise auth, and HA alerting. Base: upstream tag `v12.4.8`.

## Branch Strategy
- **12.4.x-intergral**: current branch — `v12.4.8` + the Intergral customizations below.
- **12.3.x-intergral**: previous fork line (`v12.3.0` + 104 commits). Source of this port.
- **Remotes**: origin = intergral/grafana, grafana = grafana/grafana (upstream).
- Porting process and build/test tooling live in the sibling `grafana-port-tools` repo
  (see `CLAUDE-for-Grafana-Project.md` there — a CLAUDE.md must not be committed here).

## Key Intergral Customizations

### 1. OpsPilot Integration
- **Files:** `public/app/intergral/*` (all new), hook points in `AppChrome.tsx`,
  `SpanDetail/index.tsx`, `TracePageHeader.tsx`, `SpanDetailLinkButtons.tsx`,
  `LogDetailsRow.tsx`, `LogLineDetailsLinks.tsx`
- OpsPilotSpanButton / OpsPilotTraceButton / OpspilotDataLinkButton post
  `{type:'opspilot-host.integration', ...}` over `BroadcastChannel('opspilot')`.
- `OpsPilotBroadcastProvider` wraps the main view in `AppChrome.tsx`.
- `useIframeNavigation`: parent window sends `{type:'navigate', path}` → SPA navigation.
- `useOpspilotMetadata`: answers `opspilot-host.getMetadata` with URL/time range/timezone.
- 12.4.8 note: `SpanDetail/index.tsx` renders the span button inside the header's
  `serviceNameAndLinks` div (the old separate linkList div is gone upstream);
  `TracePageHeader.tsx` renders the trace button inside the `!hideHeaderDetails` action area.

### 2. Auth Proxy OrgName Support
- **Files:** `pkg/services/authn/clients/{proxy,grafana}.go`,
  `pkg/services/authn/authnimpl/registration.go`, `authnimpl/sync/user_sync.go`,
  `conf/defaults.ini` `[auth.proxy]`
- Maps org names to IDs via `X-WEBAUTH-ORG` header. Priority in `FetchSyncedUserHook`:
  `r.OrgID` (X-Grafana-Org-Id) > `id.OrgID` (X-WEBAUTH-ORG) > user default.
- 12.4.8 note: `GetSignedInUser` is positional `(ctx, userID, orgID)` via the new
  `UserProxy` interface; tests wrap fakes in `NewLegacyUserProxy`.
- **Testing:** `debug/docker-compose.yml` (Traefik + nginx header injection).

### 3. HA Alerting Scheduler Partitioning — REDESIGNED for 12.4.8
- **Files:** `pkg/services/ngalert/schedule/partitioner.go`,
  `schedule/partition_stop_reason.go` (both new), `schedule/{schedule,fetcher}.go`,
  `ngalert.go`, `pkg/setting/setting_unified_alerting.go`, `conf/defaults.ini`
- FNV-1a hash of `orgID:ruleUID` mod cluster size == peer position decides ownership.
- Config: `ha_scheduler_partitioning_enabled` (default false),
  `ha_scheduler_min_cluster_size` (default 2, clamped ≥2). Flat keys in `[unified_alerting]`
  next to `ha_peers` — NOT a `[unified_alerting.ha_scheduler]` subsection (env vars are
  identical either way: `GF_UNIFIED_ALERTING_HA_SCHEDULER_*`).
- **Differences from the 12.3.x implementation (do not "restore" these):**
  - `runRemoteStateSync` / `state.RuleFilter` / the `api_prometheus.go` variadic
    `AlertInstanceManager` are GONE. Non-local rule state is served by upstream's
    `state.StoreStateReader` bound as `apiStateManager`/`apiStatusReader` in `ngalert.go`
    (same substitution upstream makes for `ha_single_node_evaluation`).
    `ha_scheduler_remote_state_sync_interval` config was dropped as dead.
  - `multiorg_alertmanager.go` is untouched — upstream now has `Peer()`. The NilPeer
    check moved into `NewPartitionFilter`, which returns nil for any peer that cannot
    report membership.
  - `PartitionStopReasonProvider` (upstream `AlertRuleStopReasonProvider` hook) stops
    reassigned rules with `errRuleReassigned` → `ForgetStateByRuleUID` (cache only).
    Without it, topology changes deleted DB state and sent false "resolved"
    notifications (the 12.3.x behaviour — a bug, fixed in this port).
  - `fetcher.go` narrows the cheap-keys list through `Owns()` — otherwise the fast path
    never fires under partitioning and every node full-fetches every tick.
  - `Filter()` snapshots cluster size+position once per pass and falls back to
    evaluate-all when position doesn't fit `[0, size)`.
- Hard startup errors: partitioning + `ha_single_node_evaluation`, and partitioning +
  `alertingSaveStatePeriodic` feature toggle (its full-sync persister would wipe other
  peers' state). Both guards verified live against the compose cluster (2026-08-13).
- **Known trade-off vs the 12.3.x design:** when a rule is adopted by a new peer after a
  topology change, the new owner starts a cold state cache, so a firing alert's
  `activeAt` resets to the adoption time (verified in cluster testing; the alert never
  leaves firing state, no resolved notification is sent, and Alertmanager dedup prevents
  re-notification — only the UI "active since" jumps). The old remote-state sync
  preserved `activeAt` within its 30s staleness window. If this matters, a follow-up
  could warm adopted rules' state from the DB on first evaluation.
- Upstream omission fixed in this fork: `ha_single_node_evaluation` was read in code but
  missing from defaults.ini, making `GF_UNIFIED_ALERTING_HA_SINGLE_NODE_EVALUATION` a
  silent no-op. The key is now listed (default false).

### 4. Alert Notification external_url Override
- **Files:** `ngalert.go` (2 sites), `notifier/alertmanager.go`,
  `setting_unified_alerting.go`, `conf/defaults.ini`
- `[unified_alerting] external_url` replaces `AppURL` in notification links (for iframe
  embedding where root_url differs from the user-facing URL).

### 5. Alert Email Branding
- **Files:** `public/emails/ng_alert_notification.html`
- OpsPilot logo/footer (opspilot.com, Intergral GmbH imprint), 🚨 for firing.
- ⚠️ `emails/templates/ng_alert_notification.mjml` is NOT updated — regenerating emails
  from MJML would revert the branding. Patch the MJML if regeneration becomes routine.

### 6. UI Trimming
- **Top bar** (`SingleTopBar.tsx`): only breadcrumbs + command palette trigger +
  TopBarExtensionPoint + HomeLink remain. Mega-menu toggle, QuickAdd, Help, extensions
  toolbar, SignIn, Invite, Profile removed (the FR Cloud shell provides them).
- **Mega menu**: starts undocked (`AppChromeService.megaMenuDocked = false`), docks+opens
  itself via matchMedia ≥1200px in `MegaMenu.tsx`; `useResponsiveDockedMegaMenu` in
  `AppChrome.tsx` is a no-op.
- **Nav tree** (`navtree.go`): Profile / Data connections / Org admin gated off via
  `intergralShowTrimmedNavSections` const; help links + support bundles removed;
  alerting kept visible (GFN-45).
- **Command palette** (`staticActions.ts`): theme-switch, dev-tooling and invite-user
  actions removed. Re-derived against 12.4.8 — the 12.3.x version of this file had a
  broken import and was not carried.

### 7. Ad-hoc Filter Datasource UID Resolution
- **Files:** `dashboard-scene/utils/variables.ts`, `transformSaveModelSchemaV2ToScene.ts`,
  `template_srv.ts`, `variables/adhoc/actions.ts` + pickers
- Resolves datasource *names* stored in the `uid` field to real UIDs
  (`getInstanceSettings`), guarded by `resolvedDs?.uid` — do not drop the guard, an
  instance without uid must fall back to the original ref (see tracking.test.ts).

### 8. Dashboard Save Permission Refresh + afterSave Retry
- **Files:** `useDashboardSave.tsx`, `useSaveDashboard.ts`,
  `DashboardScenePageStateManager.ts`, `dashboard/api/{v1,v2}.ts`,
  `apiserver/{client,types}.ts`; backend `accesscontrol/{acimpl/service,resolvers}.go`
- Frontend refetches permissions before post-save redirect; `?afterSave=1` triggers up
  to 6×500ms retries on 404/403/500 (multi-instance propagation lag). Backend clears
  basic-role caches in `ClearUserPermissionCache`.

### 9. CI / Build
- `.github/workflows/`: upstream workflows deleted wholesale (**delete by directory on
  future ports** — catches workflows upstream adds); Intergral's 4 remain:
  `docker_build.yml`, `on_release.yml`, `on_push_go.yml` (8-shard, uses
  `go-version-file: go.mod`), `on_push_ui.yml`. `.github/actions/` kept.
- `Dockerfile`: `ARG GO_BUILD_DEV` passed to `make build-go` (Delve-friendly builds).
- `scripts/ci/backend-tests/shard.sh`: excludes 6 non-test module roots.

## Already Upstream in v12.4.8 — intentionally NOT ported
- Panel-menu "New alert rule" (`PanelMenuBehavior.tsx`) + `alertRuleFormSchema.ts` — the
  12.3.x commit b88f94fa5cc is native upstream, with tests.
- `pkg/services/auth/auth.go` + `oauth_token.go` nil-session guards (#115728).
- `pkg/tsdb/tempo/search.go` tableType default (#116313).

## Intentionally NOT ported from 12.3.x
- Docker plugin-zip extraction (was reverted on 12.3.x itself, #54).
- `isOldFusionReactorSpan` in `explore/utils/links.ts` — dead code, no caller ever
  passed it.
- `TracePageActions.tsx` — dead file, imported by nothing.
- `APIGroupRunner.GetRestConfig` — swallowed errors, satisfied no interface.
- ALL test skips / snapshot churn / `jest.config.js` `.skip.test.*` mechanism /
  Makefile `test-go-integration -skip` list / `pkg/tests/apis/*` reverts /
  `public_testdata.golden.jsonc` mutation. Policy: **re-validate flaky tests against the
  new base; never inherit skips.**
- Debug/staging console.log commits, docs revert, loki datasource churn.

## Known Pre-existing Upstream Issues (not ours)
All reproduced on a pristine v12.4.8 worktree on this machine (2026-08-13):
- `TestProcessTicks` in `pkg/services/ngalert/schedule` fails under `-race`
  (`models.ConvertToRecordingRule` mutating a rule while a routine reads it).
- 6 jest datetime/locale tests fail locally (machine ICU locale differs from CI):
  `formats.test.ts`, `rangeutil.test.ts` (grafana-data), `TimeRangePicker/mapper.test.ts`.
- `HelpWizard.test.tsx` `SupportSnapshot › Can render` is order-dependent — fails when
  batched with certain suites, passes in isolation.
Full-suite results for the port itself: Go unit suite (short mode) zero failures;
jest 18,538 passed with only the above pre-existing/environmental failures.

## Testing Intergral Features
- Backend: `go test ./pkg/services/ngalert/schedule/... ./pkg/services/authn/... ./pkg/services/accesscontrol/...`
- Frontend: `yarn jest public/app/intergral public/app/core/components/AppChrome TraceView logs commandPalette`
- Cluster/e2e: `grafana-port-tools/local-compose-testenv` — `make build-docker-full` here,
  `make import-image && make up-cluster-5` there; `scripts/add-alert.sh`,
  `scripts/alert-status.py` against every node (all nodes must report identical rule
  health incl. non-owned rules — proves the StoreStateReader binding); on scale up/down,
  firing alerts must keep `activeAt` with no spurious resolved notifications.

## Commit Convention
`type(intergral/scope): subject` — e.g. `feat(intergral/alerting): ...`. Filter fork
commits with `git log --grep="(intergral/"`.

## Team
- **John Hawksley** - Intergral
- **Glen Dovey** - Intergral (glen.dovey@intergral.com)
- **Dan Hodgson** - Intergral (danhodgson@hotmail.co.uk)
- **Ben Donnelly** - Intergral (b.w.donnelly1@googlemail.com) - Build and explore features
- **Sam Donnelly** - Intergral (samdonnelly123@gmail.com) - Co-author on various commits

## Last Updated
2026-08-13 - After 12.3.x-intergral → v12.4.8 port.
