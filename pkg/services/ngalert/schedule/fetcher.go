package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// updateSchedulableAlertRules updates the alert rules for the scheduler.
// It returns diff that contains rule keys that were updated since the last poll,
// and an error if the database query encountered problems.
func (sch *schedule) updateSchedulableAlertRules(ctx context.Context) (diff, error) {
	start := time.Now()
	defer func() {
		sch.metrics.UpdateSchedulableAlertRulesDuration.Observe(
			time.Since(start).Seconds())
	}()

	// Cluster topology changes redistribute rule ownership even when no rule changed,
	// so force a re-fetch when the member count moves.
	forceUpdate := false
	if sch.partitioner != nil {
		if currentSize := sch.partitioner.ClusterSize(); sch.lastClusterSize != currentSize {
			sch.log.Info("Cluster size changed, forcing rule re-fetch",
				"previousSize", sch.lastClusterSize, "currentSize", currentSize)
			sch.lastClusterSize = currentSize
			forceUpdate = true
		}
	}

	if !forceUpdate && !sch.schedulableAlertRules.isEmpty() {
		keys, err := sch.ruleStore.GetAlertRulesKeysForScheduling(ctx)
		if err != nil {
			return diff{}, err
		}
		// schedulableAlertRules only holds this peer's share, so the key set must be
		// narrowed the same way or needsUpdate() always reports a length mismatch and
		// the cheap-keys fast path never fires.
		if sch.partitioner != nil {
			local := make([]models.AlertRuleKeyWithVersion, 0, len(keys))
			for _, k := range keys {
				if sch.partitioner.Owns(k.OrgID, k.UID) {
					local = append(local, k)
				}
			}
			keys = local
		}
		if !sch.schedulableAlertRules.needsUpdate(keys) {
			sch.log.Debug("No changes detected. Skip updating")
			return diff{}, nil
		}
	}
	// At this point, we know we need to re-fetch rules as there are changes.
	q := models.GetAlertRulesForSchedulingQuery{
		PopulateFolders: !sch.disableGrafanaFolder,
	}
	if err := sch.ruleStore.GetAlertRulesForScheduling(ctx, &q); err != nil {
		return diff{}, fmt.Errorf("failed to get alert rules: %w", err)
	}
	if sch.partitioner != nil {
		q.ResultRules = sch.partitioner.Filter(q.ResultRules)
	}
	d := sch.schedulableAlertRules.set(q.ResultRules, q.ResultFoldersTitles)
	sch.log.Debug("Alert rules fetched", "rulesCount", len(q.ResultRules), "foldersCount", len(q.ResultFoldersTitles), "updatedRules", len(d.updated))
	return d, nil
}
