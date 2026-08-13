package schedule

import (
	"context"
	"math/rand"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/services/ngalert/state"
)

func TestPartitionStopReasonProvider_FindReason(t *testing.T) {
	logger := log.New("test")
	key := models.AlertRuleKeyWithGroup{
		AlertRuleKey: models.AlertRuleKey{OrgID: 1, UID: "rule-1"},
		RuleGroup:    "group",
	}

	t.Run("returns no reason when the rule is still owned locally", func(t *testing.T) {
		provider := NewPartitionStopReasonProvider(&mockPartitioner{clusterSize: 2})
		reason, err := provider.FindReason(context.Background(), logger, key)
		require.NoError(t, err)
		assert.Nil(t, reason, "owned rule must fall back to the scheduler's default stop reason")
	})

	t.Run("returns errRuleReassigned when the rule left this peer's partition", func(t *testing.T) {
		provider := NewPartitionStopReasonProvider(&mockPartitioner{
			clusterSize: 2,
			ownsFunc:    func(orgID int64, ruleUID string) bool { return false },
		})
		reason, err := provider.FindReason(context.Background(), logger, key)
		require.NoError(t, err)
		assert.ErrorIs(t, reason, errRuleReassigned)
	})

	t.Run("returns no reason for a nil partitioner", func(t *testing.T) {
		provider := NewPartitionStopReasonProvider(nil)
		reason, err := provider.FindReason(context.Background(), logger, key)
		require.NoError(t, err)
		assert.Nil(t, reason)
	})
}

func TestSchedule_deleteAlertRule_Reassigned(t *testing.T) {
	// When a rule is dropped from this peer's schedule because HA partitioning moved it
	// to another peer, the rule routine must stop with errRuleReassigned — NOT
	// errRuleDeleted — so that it forgets the local cache instead of deleting database
	// state and sending resolved notifications.
	ctx := context.Background()

	mp := &mockPartitioner{
		clusterSize: 2,
		ownsFunc:    func(orgID int64, ruleUID string) bool { return false },
	}

	ruleStore := newFakeRulesStore()
	sch := setupScheduler(t, ruleStore, nil, nil, nil, nil, NewPartitionStopReasonProvider(mp), withPartitioner(mp))
	ruleFactory := ruleFactoryFromScheduler(sch)
	rule := models.RuleGen.GenerateRef()
	ruleStore.PutRule(ctx, rule)
	key := rule.GetKey()
	info, _ := sch.registry.getOrCreate(ctx, ruleWithFolder{rule: rule, folderTitle: ""}, ruleFactory)

	sch.deleteAlertRule(ctx, key)

	require.ErrorIs(t, info.(*alertRule).ctx.Err(), errRuleReassigned)
	require.False(t, sch.registry.exists(key))
}

func TestRuleRoutine_ReassignedKeepsDatabaseState(t *testing.T) {
	// The regression this guards against: on every cluster topology change, reassigned
	// rules were stopped with errRuleDeleted, which deleted their alert instances from
	// the database and sent spurious "resolved" notifications for firing alerts. With
	// errRuleReassigned only the local cache may be dropped.
	gen := models.RuleGen
	rule := gen.With(withQueryForState(t, eval.Alerting)).GenerateRef()

	sender := NewSyncAlertsSenderMock()
	ruleStore := newFakeRulesStore()
	instanceStore := &state.FakeInstanceStore{}
	registry := prometheus.NewPedanticRegistry()
	sch := setupScheduler(t, ruleStore, instanceStore, registry, sender, nil, nil, withSchedulerClock(clock.NewMock()))

	results := eval.GenerateResults(
		rand.Intn(5)+1,
		eval.ResultGen(
			eval.WithEvaluatedAt(sch.clock.Now()),
			eval.WithState(eval.Alerting),
		),
	)
	_ = sch.stateManager.ProcessEvalResults(context.Background(), sch.clock.Now(), rule, results, nil, nil)
	require.NotEmpty(t, sch.stateManager.GetStatesForRuleUID(context.Background(), rule.OrgID, rule.UID))

	factory := ruleFactoryFromScheduler(sch)
	ruleInfo := factory.new(context.Background(), ruleWithFolder{rule: rule, folderTitle: ""})
	stoppedChan := make(chan error)
	go func() {
		stoppedChan <- ruleInfo.Run()
	}()

	ruleInfo.Stop(errRuleReassigned)
	err := waitForErrChannel(t, stoppedChan)
	require.NoError(t, err)

	// Local cache is dropped so the new owner's evaluations are authoritative here on.
	require.Empty(t, sch.stateManager.GetStatesForRuleUID(context.Background(), rule.OrgID, rule.UID))

	// No resolved notifications and no database deletions.
	sender.AssertNotCalled(t, "Send")
	for _, op := range instanceStore.RecordedOps() {
		if recorded, ok := op.(state.FakeInstanceStoreOp); ok {
			assert.NotEqual(t, "DeleteAlertInstances", recorded.Name,
				"reassignment must not delete alert instances from the database")
		}
	}
}
