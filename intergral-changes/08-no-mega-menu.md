# 8. No Mega-Menu (Forced Undocked)

**Problem:** Grafana's sidebar mega menu auto-docks based on viewport width. For FusionReactor's embedded/iframe use case, the docked mega menu wastes space and isn't desired.

**Solution:** Force the mega menu to always start in undocked (closed/overlay) state:
- Replace the responsive docking hook with a no-op
- Set the initial `megaMenuDocked` state to `false`

## Where to implement

- **AppChrome component:** disable the responsive mega menu hook
- **AppChromeService:** hardcode initial docked state to false