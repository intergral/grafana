# Claude Memory for Intergral Grafana Fork

## Project Context
This is Intergral's fork of Grafana with custom features for OpsPilot integration and enterprise auth.

## Key Intergral Customizations

### 1. OpsPilot Integration
- **Files:** `public/app/intergral/*`
- **Features:**
  - OpsPilotSpanButton: AI-powered span analysis
  - OpsPilotTraceButton: Trace-level AI analysis
  - OpsPilotBroadcastContext: Message passing with OpsPilot
  - opspilotStringify: Custom JSON serialization
  - OpspilotDataLinkButton: AI button for log details
- **Usage:** Broadcast buttons appear in trace view for AI assistance

### 2. Auth Proxy OrgName Support
- **Files:** `pkg/services/authn/clients/{grafana.go,proxy.go,grafana_test.go}`, `pkg/services/authn/authnimpl/registration.go`
- **Feature:** Map organization names to IDs via X-WEBAUTH-ORG header
- **Config:** `conf/defaults.ini` [auth.proxy] section
- **Testing:** `debug/docker-compose.yml` provides local test environment with Traefik and nginx

### 3. UI Customizations
- **Alerting Tab Removed:** Panel menu no longer shows alerting tab (already in upstream v12.3.x)
- **Profile Removed:** Top bar simplified (profile features removed or never existed in v12.3.x)
- **Mega Menu Docked:** Menu starts docked on screens >1200px width
- **Navtree Alerting Enabled:** Unified alerting visible in navigation (GFN-45)

### 4. Old FusionReactor Support
- **Files:** `public/app/features/explore/utils/links.ts`
- **Feature:** Handle profiles from older FR agents without span info (isOldFusionReactorSpan parameter)

### 5. Iframe Navigation Support
- **Files:** `public/app/intergral/useIframeNavigation.ts`, `public/app/intergral/useOpspilotMetadata.ts`, `public/app/intergral/OpsPilotBroadcastContext.tsx`
- **Feature:** Enable navigation via postMessage when Grafana is embedded in an iframe
- **Metadata Hook:** useOpspilotMetadata sends URL, time range, and timezone to OpsPilot host
- **Usage:** Parent window can send `{type: 'navigate', path: '/some/path'}` to navigate Grafana
- **Commit:** 63626370fa4 (from origin/iframe-nav branch) + 94ffd52fbfa (added missing useOpspilotMetadata)

## Branch Strategy
- **v12.0.x:** Old stable fork (Intergral customizations on Grafana 12.0)
- **exp_12.3.x-intergral-migration:** Current development (Grafana 12.3 base + Intergral features)
- **main:** Fork's main branch (tracks intergral/main)
- **Remotes:** origin = intergral/grafana, grafana = grafana/grafana (upstream)

## Recent Forward-Port (Jan 2026)
Cherry-picked from v12.0.x and origin/iframe-nav to exp_12.3.x-intergral-migration:
- ✅ 94ffd52fbfa: Add missing useOpspilotMetadata hook (required by iframe-nav)
- ✅ ea3d6cc5538: Add iframe navigation support via postMessage (from iframe-nav branch)
- ✅ e5a7f6af536: Enable alerting in navtree (GFN-45)
- ✅ f1f93818a29: Add Docker Compose debug infrastructure for auth proxy testing
- ✅ 554a5fce5a7: OrgName header support to auth proxy
- ✅ 35936a60301: Simplify OpsPilotSpanButton layout (removed unnecessary div wrapper)
- ✅ a78287fc60b: Update OpsPilot button handling

## Commits Already Present (Skipped During Cherry-Pick)
These were empty during cherry-pick because changes already exist in exp_12.3.x:
- 35fddb555ba: apiserver runner error handling
- cb17c733fd9: Remove navtree items (theme preference actions already removed in upstream)
- 438fc70f4be: Adhoc UID vars logic
- 65ca4ddc1c6: Broadcast integration structure
- 944708d71e0: Initialize references array

## Not Cherry-Picked (Intentional)
- 5147517b756: Synthetic OpsPilot links (reverted by commit 9bf43de21f6)
- 5297daf67ad, 4394129f29b: Adhoc logging (debug console.logs for staging only)
- Various test skips/fixes: Test infrastructure only, not functional changes
- 9d039d0a96b: OpsPilot integration commit (features already present from earlier commits)

## Testing Intergral Features

### Auth Proxy with OrgName
```bash
cd debug
docker-compose up -d
# Start Grafana: make run
# Access: http://grafana.localhost
# Creates user with org "admin@localhost" via X-WEBAUTH-ORG header
```

### OpsPilot Integration
- Load trace view in Explore
- Look for AI buttons on spans/traces
- Click "Ask OpsPilot" for span analysis options
- Check broadcast integration works with OpsPilot host

### UI Customizations
- Verify mega menu docks automatically on wide screens
- Check alerting appears in navigation tree
- Verify panel menus don't show removed items

## Important Gotchas

### Forward-Porting
- Always check v12.0.x for Intergral features before assuming they're missing
- Many commits in v12.0.x are test fixes or upstream changes - filter carefully
- Use `git log --author="dan\|Glen\|glen\|intergral"` to find Intergral commits
- Debug/staging commits (console.logs) should NOT be forward-ported
- Some features may already exist due to upstream Grafana v12.3.x improvements

### Cherry-Pick Conflicts
- OpsPilotSpanButton: HEAD has i18n support, v12.0.x doesn't - keep i18n version
- staticActions.ts: Upstream v12.3.x already removed theme actions
- Most conflicts are due to upstream architectural improvements in v12.3.x

### Commit Identification
- Look for commits by Intergral authors: dan, Glen Dovey, John Hawksley
- Functional commits: feat(), fix(), refactor()
- Skip: test fixes, lint fixes, workflow changes, debug logging
- Reverted commits (check for "Revert" in log)

## Commit Convention (Going Forward)

### Recommended Format
Use `(intergral/scope)` prefix for fork-specific commits:
```
feat(intergral/opspilot): add span button integration
fix(intergral/auth): add OrgName header support
chore(intergral/ui): remove alerting tab from panel menu
refactor(intergral/navtree): enable alerting visibility
```

### Benefits
- Easy to filter: `git log --grep="(intergral/"`
- Clear ownership and purpose
- Automated scripts can identify fork-specific commits
- Makes forward-porting 10x easier

## File Locations

### Intergral Custom Code
- `public/app/intergral/*` - All OpsPilot components
- `pkg/services/authn/clients/*` - Auth proxy customizations
- `debug/*` - Testing infrastructure
- `conf/defaults.ini` - Default config with auth proxy enabled

### Key Config Files
- `WORKFLOW.md` - Trunk-based development workflow
- `.github/CODEOWNERS` - Subsystem ownership
- `.github/PULL_REQUEST_TEMPLATE.md` - PR requirements

## Commands Reference

### Find Intergral Commits
```bash
# Between branches
git log v12.0.x --not exp_12.3.x-intergral-migration --oneline --author="dan\|Glen\|intergral"

# With grep filter
git log --grep="(intergral/" --oneline
```

### Cherry-Pick Helper
```bash
# Cherry-pick with conflict handling
git cherry-pick <commit>
# If empty: git cherry-pick --skip
# If conflict: resolve, then git cherry-pick --continue
```

### Testing
```bash
# Backend tests
make test

# Frontend tests
yarn test

# Run Grafana locally
make run

# Docker test environment
cd debug && docker-compose up -d
```

## Team
- **John Hawksley** - Intergral
- **Glen Dovey** - Intergral (glen.dovey@intergral.com)
- **Dan Hodgson** - Intergral (danhodgson@hotmail.co.uk)
- **Ben Donnelly** - Intergral (b.w.donnelly1@googlemail.com) - Build and explore features
- **Sam Donnelly** - Intergral (samdonnelly123@gmail.com) - Co-author on various commits

## Last Updated
2026-01-09 - After v12.0.x to exp_12.3.x-intergral-migration forward-port
