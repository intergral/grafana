# 1. HA Alerting Partitioning

**Problem:** In a multi-instance Grafana HA deployment, every instance evaluates every alert rule. This wastes resources and can cause duplicate notifications.

**Solution:** Partition alert rules across cluster members so each instance only evaluates a subset (1/N) of rules.

## How it works

- A partitioner assigns each rule to an instance using a deterministic hash: `FNV32(orgID:ruleUID) % clusterSize`. This ensures stable assignment -- the same rule always goes to the same instance unless the cluster topology changes.
- The partitioner gets the list of peers from Grafana's existing HA mechanism (memberlist or Redis-based peer discovery, whatever the alerting scheduler already uses).
- When cluster size is below a configurable minimum, partitioning is disabled and all instances evaluate all rules (safety fallback).
- When a topology change is detected (peer joins/leaves), all rules are re-fetched to recompute partition assignments.

## Remote state sync

**Problem:** If instance A evaluates rule X, instance B still needs to know rule X's current state to serve accurate API responses (e.g. the Prometheus-compatible `/api/v1/rules` endpoint). Without this, the Grafana UI shows stale/missing alert states on instances that don't own those rules.

**Solution:**
- A background goroutine periodically reads alert rule states from the database for rules NOT owned by the local instance.
- These remote states are loaded into the in-memory state cache so the API can return them.
- Stale entries for rules that are no longer active are cleaned up.
- The Prometheus API endpoint falls back to the state cache when the scheduler has no local status for a rule.

## Configuration

3 new keys in `[unified_alerting]` section of `defaults.ini`:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ha_scheduler_partitioning_enabled` | bool | `false` | Enable HA partitioning |
| `ha_scheduler_min_cluster_size` | int | `2` | Partitioning disabled below this |
| `ha_scheduler_remote_state_sync_interval` | duration | `30s` | How often to sync remote rule states |

## Where to implement

- **Partitioner logic:** new file alongside the alert scheduler
- **Fetcher:** filter rules after fetch, detect topology changes to force re-fetch
- **Scheduler config:** accept and wire the partitioner
- **State manager:** add background remote state refresh loop
- **Prometheus API handler:** fall back to state cache for non-local rules
- **Settings:** parse the 3 new config keys
- **ngalert init:** create partitioner and wire everything together