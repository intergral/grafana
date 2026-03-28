# 16. Test Coverage for Intergral Additions

**Problem:** The Intergral-specific code additions had no dedicated test coverage. If upstream merges or refactors break these customizations, there would be no automated way to detect regressions.

**Solution:** Add unit tests for all Intergral-specific frontend components and backend configuration fields.

## Frontend tests

All tests live in `public/app/intergral/` alongside the source files.

### `opspilotStringify.test.ts` (18 tests)

Tests the pure `opsPilotStringify()` formatting function which converts trace/span objects into human-readable text for the OpsPilot AI assistant.

Coverage:
- Primitive and edge-case inputs (null, undefined, numbers, strings, empty objects)
- Flat key-value formatting with camelCase-to-label conversion
- Nested objects rendered as labeled sections with blank-line separators
- Arrays: simple (comma-separated), key-value pairs, general objects
- Skipping of null/undefined/empty values
- Realistic trace data structure

### `OpsPilotSpanButton.test.tsx` (12 tests)

Tests the span-level OpsPilot dropdown menu component.

Coverage:
- Button renders with "Ask OpsPilot" label
- "Analyze Span" menu item always present
- "Analyze Error" menu item conditionally shown (requires `statusCode === 2` AND a `reason` tag)
- "Analyze Query" menu item conditionally shown (requires `flv === 'JDBCRequest'` tag)
- Broadcast messages sent with correct `content_type` for each action (span/error/jdbc)
- Tag filtering: generic tags limited to `flv`, `app_name`, `tann` plus action-specific tags
- `references` array cleared before broadcasting

### `OpsPilotTraceButton.test.tsx` (5 tests)

Tests the trace-level "Analyze Trace" action button.

Coverage:
- Returns null when trace is undefined
- Renders button when trace is provided
- Broadcasts trace content with `content_type: 'trace'`
- Clears span references before broadcasting
- Handles postMessage errors gracefully (logs to console)

### `OpspilotDataLinkButton.test.tsx` (6 tests)

Tests the log data link button that sends log content to OpsPilot.

Coverage:
- Renders with link title text
- Broadcasts log content via BroadcastChannel when in an iframe (`window.parent !== window`)
- Does NOT broadcast when not in an iframe
- Handles broadcast errors gracefully
- Passes through `buttonProps` to underlying Button

### `useIframeNavigation.test.ts` (5 tests)

Tests the hook that enables SPA navigation from parent window postMessages.

Coverage:
- Navigates via `locationService.push()` on valid `{ type: 'navigate', path: string }` messages
- Ignores messages with wrong type, non-string path, or null data
- Cleans up event listener on unmount

## Backend tests

### `pkg/setting/setting_unified_alerting_test.go` (5 new tests)

Tests the Intergral-specific configuration fields in `[unified_alerting]`.

**ExternalURL (2 tests):**
- Defaults to empty string when not configured
- Reads custom `external_url` value correctly

**HA Scheduler Partitioning (3 tests):**
- Defaults: `ha_scheduler_partitioning_enabled=false`, `ha_scheduler_min_cluster_size=2`, `ha_scheduler_remote_state_sync_interval=30s`
- Reads custom values for all three fields
- Returns error on invalid duration for `ha_scheduler_remote_state_sync_interval`

Note: The partitioner logic itself (`pkg/services/ngalert/schedule/partitioner.go`) already had thorough test coverage in `partitioner_test.go` covering distribution, consistency, org isolation, and topology changes. No additional tests were needed there.

## How to recreate

The mocking patterns used:

- **BroadcastChannel context:** Mocked via `jest.mock('./OpsPilotBroadcastContext')` returning a mock `postMessage` function
- **opsPilotStringify:** Mocked to return `JSON.stringify(obj)` so broadcast content can be parsed and asserted
- **ActionButton:** Mocked to a simple `<button>` element for trace button tests
- **@grafana/runtime:** `locationService.push` mocked for navigation tests
- **Go ini config:** Uses `ini.Empty()` with `NewSection`/`NewKey` to construct test configs, then calls `cfg.ReadUnifiedAlertingSettings(f)`

To run all Intergral tests:

```bash
# Frontend
npx jest --no-cache public/app/intergral/

# Backend (settings)
go test ./pkg/setting/ -run "TestIntergral" -v

# Backend (partitioner - pre-existing)
go test ./pkg/services/ngalert/schedule/ -run "TestPartitioner" -v
```
