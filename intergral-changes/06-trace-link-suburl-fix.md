# 6. Trace Link Sub-URL Fix

> **Note:** Check if upstream has fixed this in the target version before reimplementing.

**Problem:** When Grafana is served under a sub-path (e.g. `/grafana/`), clicking links in span detail breaks because `locationService.push()` already operates relative to the app base URL, but the link href includes the full sub-path prefix, resulting in a double prefix.

**Solution:** Before calling `locationService.push()` with a link href, strip the `config.appSubUrl` prefix if present.

## Where to implement

- The span detail link buttons component in the trace/explore view