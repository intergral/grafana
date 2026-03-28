# 10. Navigation Tree Trimming

**Problem:** Several Grafana navigation sections (Profile, Data Connections, Admin, Help) aren't relevant for the FusionReactor deployment and clutter the UI.

**Solution:** Disable these sections in the navigation tree builder. Previously done with `&& false` guards on the condition checks.

**Better approach for future versions:** Rather than fragile `&& false` hacks, consider using Grafana's feature flags or RBAC to hide these sections, or find the nav tree registration point and skip registering unwanted sections entirely.

## Sections to remove

- Profile section
- Data Connections section
- Admin section
- Help/support links

## Where to implement

- The navigation tree builder service (server-side, in the navtree implementation)