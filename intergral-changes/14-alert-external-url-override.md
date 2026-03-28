# 14. Alert Notification External URL Override

**Problem:** Alert notifications (emails, Slack, etc.) include links built from Grafana's `root_url` setting. When Grafana is embedded in an iframe (as in FusionReactor Cloud), `root_url` points to the internal Grafana path (e.g. `https://app.fusionreactor.io/g/`), but users should be directed to the parent app's URL (e.g. `https://app.fusionreactor.io/grafana/alerting/grafana/{uid}/view`) so the Next.js catch-all route handles navigation within the iframe.

**Solution:** Add a new `external_url` setting under `[unified_alerting]` that, when set, overrides `AppURL` for all alert notification link generation.

## How it works

When `external_url` is set, it replaces `cfg.AppURL` in the three places that build alert notification URLs:

1. **Alerts router and scheduler** (`appUrl` used by `StateToPostableAlert` to construct `GeneratorURL`)
2. **Local Grafana Alertmanager** (`ExternalURL` in `GrafanaAlertmanagerOpts`, available as `.ExternalURL` in notification templates)
3. **Remote Alertmanager config** (`ExternalURL` field)

When not set, behaviour is unchanged -- `AppURL` (from `root_url`) is used as before.

## Configuration

In `[unified_alerting]` section of `defaults.ini`:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `external_url` | string | empty | e.g. `https://app.fusionreactor.io/grafana` |

## Result

Alert notification links like `GeneratorURL` become `https://app.fusionreactor.io/grafana/alerting/grafana/{uid}/view`, which hits the Next.js `/grafana/[...slug]/` catch-all route and renders inside the iframe layout.

## Where to implement

- **`pkg/setting/setting_unified_alerting.go`:** Add `ExternalURL` field to `UnifiedAlertingSettings` struct, read it from the `external_url` ini key
- **`pkg/services/ngalert/ngalert.go`:** Use the override URL (if set) instead of `cfg.AppURL` when parsing `appUrl` for the alerts router/scheduler, and when setting `ExternalURL` on the remote alertmanager config
- **`pkg/services/ngalert/notifier/alertmanager.go`:** Use the override URL (if set) instead of `cfg.AppURL` for the local alertmanager's `ExternalURL`
- **`conf/defaults.ini`:** Document the new setting in the `[unified_alerting]` section