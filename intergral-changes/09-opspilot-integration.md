# 9. OpsPilot Integration

**Problem:** FusionReactor includes an AI assistant called OpsPilot. It needs to communicate with the embedded Grafana instance to receive trace/span/log data for analysis, and Grafana needs buttons that let users send data to OpsPilot.

**Solution:** A suite of components providing cross-tab and cross-frame communication.

## Broadcast Channel

- A React context providing a `BroadcastChannel('opspilot')` for cross-tab communication
- Wraps the main app chrome so it's available everywhere
- Also mounts the iframe navigation and metadata hooks

## Analyze Buttons -- Traces

- **Trace button:** On the trace page header action bar (alongside Share/Feedback buttons), sends full trace data via broadcast channel
- **Span button:** "Ask OpsPilot" dropdown on span detail header row (right-aligned, next to Share and Links buttons) with options: "Analyze Span", "Analyze Error", "Analyze Query" -- sends relevant span data via broadcast channel
- The span button uses a custom OpsPilot SVG icon (`public/app/intergral/opspilot-icon.svg`) instead of Grafana's built-in icon set. It uses Grafana's `Button` component for consistent sizing with adjacent buttons.
- A stringify utility formats trace/span objects into human-readable text for the AI

## Analyze Buttons -- Logs

- **Log link button:** Custom button for data links titled "OpsPilot AI" on log lines -- sends log content via broadcast channel
- This button is **datasource-driven**: it only appears when the Loki datasource has a derived field configured with `urlDisplayLabel: "OpsPilot AI"`. The code in `LogDetailsRow.tsx` and `LogLineDetailsLinks.tsx` checks for `link.title === 'OpsPilot AI'` and swaps in the `OpspilotDataLinkButton` instead of a normal link.
- The derived field needs: a catch-all regex like `([\s\S]*)`, a non-empty URL (e.g. `#opspilot` since the button doesn't navigate), and `urlDisplayLabel` set to exactly `"OpsPilot AI"`.
- Example Loki provisioning config:
  ```yaml
  derivedFields:
    - name: OpsPilot
      matcherRegex: "([\\s\\S]*)"
      url: "#opspilot"
      urlDisplayLabel: "OpsPilot AI"
  ```

## Iframe Communication

- **Metadata hook:** Responds to `opspilot-host.getMetadata` postMessage events from the parent frame with current URL, time range, and timezone
- **Navigation hook:** Listens for `navigate` postMessage events to enable SPA navigation when Grafana is embedded in an iframe (so the parent can drive navigation without full page reloads)

## Span Detail Layout (v12.4 changes)

- In v12.4, upstream moved span links from a dedicated row into the header alongside the operation name. The Intergral customization places OpsPilot, Links, and Share buttons in the header row, right-aligned via `marginLeft: auto`.
- The span links dropdown is forced to always show (via `MAX_LINKS = 0` in `SpanDetailLinkButtons.tsx`) rather than v12.4's default of showing individual buttons when there are <=3 links. This matches the original "Links" dropdown behavior.
- The `TracePageActions.tsx` file from v12.3 was removed as dead code -- v12.4's `TracePageHeader` handles the trace-level OpsPilot button directly.

## Where to implement

- **All OpsPilot components** live in `public/app/intergral/`
- **AppChrome:** wrap in the broadcast provider
- **Trace page header:** add trace analyze button in the `{!hideHeaderDetails && (...)}` action area
- **Span detail component:** add span analyze dropdown in the header `serviceNameAndLinks` div
- **Span detail link buttons:** set `MAX_LINKS = 0` to force dropdown
- **Log line detail links:** intercept "OpsPilot AI" titled links and route to custom button
- **Loki datasource config:** add derived field with `urlDisplayLabel: "OpsPilot AI"`