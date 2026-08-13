package schedule

import (
	"context"
	"errors"

	"github.com/grafana/grafana/pkg/infra/log"
	ngmodels "github.com/grafana/grafana/pkg/services/ngalert/models"
)

// errRuleReassigned is the stop reason for a rule that left this peer's partition after a
// cluster topology change. It is deliberately NOT errRuleDeleted: the rule routine only
// calls DeleteStateByRuleUID (which deletes the rule's alert instances from the database
// and sends resolved notifications) for errRuleDeleted. For a reassigned rule we want
// ForgetStateByRuleUID — drop the local cache and leave the database state to the new
// owner, so firing alerts keep their identity across the handoff.
var errRuleReassigned = errors.New("rule reassigned to another peer by HA partitioning")

// PartitionStopReasonProvider implements AlertRuleStopReasonProvider for HA partitioning.
type PartitionStopReasonProvider struct {
	partitioner RulePartitioner
}

func NewPartitionStopReasonProvider(partitioner RulePartitioner) *PartitionStopReasonProvider {
	return &PartitionStopReasonProvider{partitioner: partitioner}
}

// FindReason returns errRuleReassigned when the rule is no longer owned locally. Returning
// (nil, nil) makes the scheduler fall back to its default reason (errRuleDeleted), which is
// correct for rules that really were deleted.
func (s *PartitionStopReasonProvider) FindReason(_ context.Context, logger log.Logger, key ngmodels.AlertRuleKeyWithGroup) (error, error) {
	if s.partitioner == nil || s.partitioner.Owns(key.OrgID, key.UID) {
		return nil, nil
	}
	logger.New(key.LogContext()...).Info("Rule stopped because HA partitioning reassigned it to another peer; keeping database state")
	return errRuleReassigned, nil
}
