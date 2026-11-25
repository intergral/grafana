# Claude Memory for Intergral Grafana Fork

## Project Context
This is Intergral's fork of Grafana (FusionReactor Cloud / OpsPilot) with custom features
for OpsPilot integration, enterprise auth, and HA alerting. Base: upstream tag `v12.4.16`
(merged onto the original `v12.4.8` port on 2026-11-25).

## Branch Strategy
- **12.4.x-intergral**: trunk for the whole 12.4 stream — `v12.4.16` + the Intergral
  customizations below. Feature branches fork from it and merge back via PRs targeting it.
  Note `origin/HEAD` points at `main`, which is a stale upstream-merge branch — never use it.
- **12.3.x-intergral**: previous fork line (`v12.3.0` + 104 commits). Source of the
  12.4.8 port. **12.4.x-intergral-old** on origin is a BAD v12.4.2-based port — never
  source from it.
- **v12.4.x** on origin: pristine `v12.4.8` snapshot used as the PR #66 review base
  (protected). Not moved by upstream patch merges.
- **Remotes**: origin = intergral/grafana, grafana = grafana/grafana (upstream).
- Porting process and build/test tooling live in the sibling `grafana-port-tools` repo
  (see `CLAUDE-for-Grafana-Project.md` there — a CLAUDE.md must not be committed here).

## Taking Upstream Releases
- **Patch release on the same minor (e.g. 12.4.8 → 12.4.16): MERGE the upstream tag.**
  The upstream tag is an ancestor-compatible line; the fork commits are already curated.
  1. Branch from `12.4.x-intergral`; `git fetch grafana --tags`;
     `git merge --no-ff --no-commit vX.Y.Z`.
  2. Every conflict so far is modify/delete on upstream `.github/workflows/*` — resolve by
     deleting. Also delete any workflow upstream ADDED (they merge in silently): only
     Intergral's 4 workflows may remain.
  3. Verify the fork delta survived byte-for-byte: for each file in
     `git diff --name-only <old tag> HEAD`, the `+/-` lines of `git diff <old tag> HEAD`
     must equal those of `git diff <new tag> $(git write-tree)`; and the merged tree must
     differ from the new tag ONLY in fork files.
  4. Read upstream's diff for every overlapping file for semantic interaction.
  5. `make gen-go`, `yarn install --immutable`, build, targeted tests (below), typecheck.
  6. **The PR must land as a merge commit — never squash or rebase.** Squashing drops the
     upstream tag as a parent and the next upstream merge re-conflicts everything.
- **New minor/major (e.g. → 13.x): curated re-port** onto a pristine snapshot of the tag
  (the 12.3.x → 12.4.8 process). Re-derive each customization against the new code; do
  not cherry-pick blindly.
- Either way: test skips are never inherited without checking their removal condition.

## Releases
- Tags `v<upstream>-intergral[-<letter>]-{RC<n>,REL}`; letter suffix = increment on the
  same base (`v12.4.8-intergral-REL`, `-B-REL` 2026-08-27, `-C-REL` 2026-09-15).
- `on_release.yml` builds the multi-arch image for every tag; Docker `latest` moves only
  on `-REL` tags. `BUILD_BRANCH` = the tag, so an RC and its REL have different digests —
  equivalence rests on the commit SHA.
- Docker publish needs manual "Docker Hub" environment approval per run.

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
- `proxyFieldOrgName` is part of `proxyFields`, so it is part of the proxy cache key.
  12.4.16 changed `getProxyCacheKey` to length-prefix every field (missing ones too), so
  cache keys change once on upgrade (one-time cache miss, harmless).
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
    Without it, topology changes deleted DB state (the 12.3.x behaviour — a bug).
  - `fetcher.go` narrows the cheap-keys list through `Owns()` — otherwise the fast path
    never fires under partitioning and every node full-fetches every tick.
  - `Filter()` snapshots cluster size+position once per pass and falls back to
    evaluate-all when position doesn't fit `[0, size)`.
  - **Alert broadcast (#67):** `ngalert.go` enables upstream's `BroadcastAlerts`
    (`alerts:broadcast` memberlist channel) when partitioning is on, not only for
    `ha_single_node_evaluation`. Grafana HA gossips silences + nflog but not alerts, so
    after a handoff the old owner's Alertmanager kept an unrefreshed entry and sent a
    false `resolved` when it expired (+120s). Broadcast cut this from 5 to ~1 per
    topology change (delivery-level test). Gossiped nflog keeps deliveries deduplicated.
- Hard startup errors: partitioning + `ha_single_node_evaluation`, and partitioning +
  `alertingSaveStatePeriodic` feature toggle (its full-sync persister would wipe other
  peers' state). Both guards verified live against the compose cluster (2026-08-13).
- **Open issues (live in production):**
  - GFN-131: residual ~1 false `resolved` per topology change (best-effort gossip).
    Test with a contact point that really delivers — a failing receiver hides it.
  - GFN-132: adopted rules start from a cold state cache, so `activeAt` resets and state
    history records a bogus Normal→Alerting. Fix sketch: factor `Manager.Warm`'s inner
    loop into a per-rule `WarmRule`, call it on first evaluation behind an empty-cache
    guard.
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
  ⚠️ **Latent since 12.3.x** (verified 2026-08-13): `MegaMenu` only mounts when the menu
  is already open/docked, and the top-bar toggle is removed — so the ≥1200px dock effect
  never runs and the menu is effectively unreachable. This is identical on the 12.3.x
  branch, i.e. current production behaviour (the FR shell provides navigation). If the
  docked menu is ever wanted, the dock-on-wide-screen decision must move out of
  `MegaMenu.tsx` into `AppChrome`/`AppChromeService`.
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

### 9. Editors Can Write Protected Contact Point Fields (#69, FRUI-599 / GFN-139)
- **Files:** `pkg/services/ngalert/accesscontrol.go` (+ `accesscontrol_test.go`),
  `pkg/tests/apis/alerting/notifications/receivers/receiver_test.go`
- `alert.notifications.receivers.protected:write` added to `receiversWriterRole`, which
  composes up to the Editor basic role. Upstream grants it only to `fixed:alerting:admin`;
  we run OSS and never make customers org Admins, so Editors could not change the URL of
  a contact point someone else created. `secrets:read` / `permissions:write` stay Admin.

### 10. Datasource Proxy 403 Log Removed (#68, GFN-130)
- **Files:** `pkg/api/pluginproxy/ds_proxy.go`
- The "falling back to 403" log fired on every Prometheus proxy GET while allowing the
  request — the bulk of the post-12.4.8 production log flood. Matches upstream 13.
  Loki gating and the DELETE/PUT/POST blocks are untouched (VUL-2026-0084).

### 11. CI / Build
- `.github/workflows/`: upstream workflows deleted wholesale (**delete by directory** on
  ports and upstream merges — catches workflows upstream adds); Intergral's 4 remain:
  `docker_build.yml`, `on_release.yml`, `on_push_go.yml` (8-shard, uses
  `go-version-file: go.mod`), `on_push_ui.yml` (4-way sharded `test:ci`, bounded jest
  workers + `workerIdleMemoryLimit`; `test:coverage` OOMs runners). `.github/actions/`
  kept and takes upstream changes.
- Push-only triggers with branch glob `**` (`*` does not match `/`, so `feat/…` branches
  got no CI). No `pull_request` triggers — they caused duplicate runs and racing
  same-tag docker publishes. GFN-137 (open): make the docker branch build
  `workflow_dispatch` only.
- `on_release.yml`: `latest` only on `-REL` tags (GFN-133); `BUILD_BRANCH` = the tag
  (GFN-129).
- `Dockerfile`: `ARG GO_BUILD_DEV` passed to `make build-go` (Delve-friendly builds).
- `scripts/ci/backend-tests/shard.sh`: excludes 6 non-test module roots.
- Legacy `// +build` lines removed (govet under Go 1.26).

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
Reproduced on a pristine v12.4.8 worktree on this machine (2026-08-13):
- `TestProcessTicks` in `pkg/services/ngalert/schedule` fails under `-race`
  (`models.ConvertToRecordingRule` mutating a rule while a routine reads it).
- 6 jest datetime/locale tests fail locally (machine ICU locale differs from CI):
  `formats.test.ts`, `rangeutil.test.ts` (grafana-data), `TimeRangePicker/mapper.test.ts`.
- `HelpWizard.test.tsx` `SupportSnapshot › Can render` is order-dependent — fails when
  batched with certain suites, passes in isolation.
- CI flakes skipped in the Makefile with run-ID evidence. Both re-checked against
  v12.4.16 and still needed:
  - `TestIntegrationProvisioning` family (SQLite "database is locked" in the
    secure-value store). No upstream fix in 12.4.16.
  - `TestSearchGetOrCreateIndexWithCancellation` (1ms-deadline race). Upstream main
    rewrote it, but `search_test.go` is unchanged in 12.4.16 — drop the skip only when
    a base contains the rewrite.
- CI flake NOT skipped: `TestIntegrationConvertPrometheusAlertmanagerEndpoints` hangs to
  the 5m suite timeout (run 33046589669); passes on re-run. Re-run, do not skip.
- `TestSmtpSend`'s 1ms-sleep race was fixed properly in smtp_test.go rather than skipped.
- Basic-role seeder lock race (`oss-ac-basic-role-seeder`): several nodes started at the
  same time against one DB killed all but one. Upstream #138362 (in 12.4.16) retries the
  lock — verify with a simultaneous multi-node start, then drop the "stagger node
  starts" workaround.

## Testing Intergral Features
- Backend: `go test ./pkg/services/ngalert/schedule/... ./pkg/services/ngalert/notifier/...
  ./pkg/services/ngalert/ ./pkg/services/authn/... ./pkg/services/accesscontrol/...
  ./pkg/api/pluginproxy/... ./pkg/setting/... ./pkg/services/navtree/...`
- Frontend: `yarn jest public/app/intergral public/app/core/components/AppChrome TraceView
  logs commandPalette public/app/features/dashboard-scene public/app/features/dashboard/api
  public/app/features/variables/adhoc public/app/features/templating`
- Cluster/e2e: `grafana-port-tools/local-compose-testenv` — `make build-docker-full` here,
  `make import-image && make up-cluster-5` there; `scripts/add-alert.sh`,
  `scripts/alert-status.py` against every node (all nodes must report identical rule
  health incl. non-owned rules — proves the StoreStateReader binding). On scale up/down,
  count false `resolved` deliveries at a working webhook receiver (expect ≤1, GFN-131).

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
2026-11-25 - Merged upstream v12.4.16 into 12.4.x-intergral (after v12.4.8-intergral-C-REL).
