package state

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	ngModels "github.com/grafana/grafana/pkg/services/ngalert/models"
)

// mockOrgReader implements OrgReader for testing.
type mockOrgReader struct {
	orgIDs []int64
	err    error
}

func (m *mockOrgReader) FetchOrgIds(_ context.Context) ([]int64, error) {
	return m.orgIDs, m.err
}

// mockRuleReader implements RuleReader for testing.
type mockRuleReader struct {
	rules ngModels.RulesGroup
	err   error
}

func (m *mockRuleReader) ListAlertRules(_ context.Context, _ *ngModels.ListAlertRulesQuery) (ngModels.RulesGroup, error) {
	return m.rules, m.err
}

// fakeInstanceReader implements InstanceReader for testing.
type fakeInstanceReader struct {
	instances []*ngModels.AlertInstance
	err       error
}

func (m *fakeInstanceReader) ListAlertInstances(_ context.Context, _ *ngModels.ListAlertInstancesQuery) ([]*ngModels.AlertInstance, error) {
	return m.instances, m.err
}

// mockRuleFilter implements RuleFilter. It returns only rules whose UIDs are in localUIDs.
type mockRuleFilter struct {
	localUIDs map[string]struct{}
}

func (f *mockRuleFilter) Filter(rules []*ngModels.AlertRule) []*ngModels.AlertRule {
	var result []*ngModels.AlertRule
	for _, r := range rules {
		if _, ok := f.localUIDs[r.UID]; ok {
			result = append(result, r)
		}
	}
	return result
}

func newTestManager(ruleFilter RuleFilter) *Manager {
	c := newCache()
	return &Manager{
		cache:                   c,
		log:                     log.New("test"),
		clock:                   clock.New(),
		ruleFilter:              ruleFilter,
		remoteStateSyncInterval: 30 * time.Second,
	}
}

func TestRefreshRemoteStates_LoadsOnlyRemoteRules(t *testing.T) {
	// Setup: 4 rules, filter claims rule1 and rule2 as local
	localFilter := &mockRuleFilter{
		localUIDs: map[string]struct{}{
			"rule1": {},
			"rule2": {},
		},
	}

	mgr := newTestManager(localFilter)

	rules := ngModels.RulesGroup{
		{UID: "rule1", OrgID: 1, Annotations: map[string]string{"a": "1"}},
		{UID: "rule2", OrgID: 1, Annotations: map[string]string{"a": "2"}},
		{UID: "rule3", OrgID: 1, Annotations: map[string]string{"a": "3"}},
		{UID: "rule4", OrgID: 1, Annotations: map[string]string{"a": "4"}},
	}

	labels3 := ngModels.InstanceLabels{"instance": "remote3"}
	labels4 := ngModels.InstanceLabels{"instance": "remote4"}
	labels1 := ngModels.InstanceLabels{"instance": "local1"}

	instances := []*ngModels.AlertInstance{
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule1"},
			Labels:           labels1,
			CurrentState:     ngModels.InstanceStateFiring,
			LastEvalTime:     time.Now(),
		},
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule3"},
			Labels:           labels3,
			CurrentState:     ngModels.InstanceStateFiring,
			LastEvalTime:     time.Now(),
		},
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule4"},
			Labels:           labels4,
			CurrentState:     ngModels.InstanceStateNormal,
			LastEvalTime:     time.Now(),
		},
	}

	mgr.orgReader = &mockOrgReader{orgIDs: []int64{1}}
	mgr.rulesReader = &mockRuleReader{rules: rules}
	mgr.instanceReader = &fakeInstanceReader{instances: instances}

	// Pre-populate cache with a local rule state to verify it's NOT overwritten
	localState := &State{
		AlertRuleUID: "rule1",
		OrgID:        1,
		CacheID:      labels1.Fingerprint(),
		Labels:       map[string]string{"instance": "local1"},
		State:        eval.Alerting,
	}
	mgr.cache.set(localState)

	// Run refresh
	mgr.refreshRemoteStates(context.Background())

	// Verify remote rules were loaded
	rule3States := mgr.cache.getStatesForRuleUID(1, "rule3")
	require.Len(t, rule3States, 1)
	assert.Equal(t, eval.Alerting, rule3States[0].State)

	rule4States := mgr.cache.getStatesForRuleUID(1, "rule4")
	require.Len(t, rule4States, 1)
	assert.Equal(t, eval.Normal, rule4States[0].State)

	// Verify local rule state still exists (not removed)
	rule1States := mgr.cache.getStatesForRuleUID(1, "rule1")
	require.Len(t, rule1States, 1)
	// The local state should NOT have been overwritten by DB refresh
	// (the refresh skips local rules entirely)
	assert.Equal(t, eval.Alerting, rule1States[0].State)
}

func TestRefreshRemoteStates_CleansUpDeletedRemoteInstances(t *testing.T) {
	localFilter := &mockRuleFilter{
		localUIDs: map[string]struct{}{
			"local-rule": {},
		},
	}

	mgr := newTestManager(localFilter)

	rules := ngModels.RulesGroup{
		{UID: "local-rule", OrgID: 1},
		{UID: "remote-rule", OrgID: 1, Annotations: map[string]string{}},
	}

	// Pre-populate cache with a state for remote-rule
	staleState := &State{
		AlertRuleUID: "remote-rule",
		OrgID:        1,
		CacheID:      data.Fingerprint(12345),
		Labels:       map[string]string{"gone": "true"},
		State:        eval.Alerting,
	}
	mgr.cache.set(staleState)

	// DB has NO instances for remote-rule anymore (rule resolved/deleted its instances)
	mgr.orgReader = &mockOrgReader{orgIDs: []int64{1}}
	mgr.rulesReader = &mockRuleReader{rules: rules}
	mgr.instanceReader = &fakeInstanceReader{instances: []*ngModels.AlertInstance{}}

	mgr.refreshRemoteStates(context.Background())

	// The stale cache entry should be removed
	states := mgr.cache.getStatesForRuleUID(1, "remote-rule")
	assert.Empty(t, states)
}

func TestRefreshRemoteStates_TopologyChange(t *testing.T) {
	// Phase 1: rule3 is remote, rule1 is local
	filter := &mockRuleFilter{
		localUIDs: map[string]struct{}{
			"rule1": {},
		},
	}

	mgr := newTestManager(filter)

	rules := ngModels.RulesGroup{
		{UID: "rule1", OrgID: 1, Annotations: map[string]string{}},
		{UID: "rule3", OrgID: 1, Annotations: map[string]string{}},
	}

	labels3 := ngModels.InstanceLabels{"instance": "r3"}
	instances := []*ngModels.AlertInstance{
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule3"},
			Labels:           labels3,
			CurrentState:     ngModels.InstanceStateFiring,
			LastEvalTime:     time.Now(),
		},
	}

	mgr.orgReader = &mockOrgReader{orgIDs: []int64{1}}
	mgr.rulesReader = &mockRuleReader{rules: rules}
	mgr.instanceReader = &fakeInstanceReader{instances: instances}

	mgr.refreshRemoteStates(context.Background())

	// rule3 should be in cache from DB
	rule3States := mgr.cache.getStatesForRuleUID(1, "rule3")
	require.Len(t, rule3States, 1)
	assert.Equal(t, eval.Alerting, rule3States[0].State)

	// Phase 2: Topology change — rule3 becomes local, rule1 becomes remote
	filter.localUIDs = map[string]struct{}{
		"rule3": {},
	}

	labels1 := ngModels.InstanceLabels{"instance": "r1"}
	mgr.instanceReader = &fakeInstanceReader{instances: []*ngModels.AlertInstance{
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule1"},
			Labels:           labels1,
			CurrentState:     ngModels.InstanceStateNormal,
			LastEvalTime:     time.Now(),
		},
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule3"},
			Labels:           labels3,
			CurrentState:     ngModels.InstanceStateFiring,
			LastEvalTime:     time.Now(),
		},
	}}

	mgr.refreshRemoteStates(context.Background())

	// rule1 is now remote and should be loaded from DB
	rule1States := mgr.cache.getStatesForRuleUID(1, "rule1")
	require.Len(t, rule1States, 1)
	assert.Equal(t, eval.Normal, rule1States[0].State)

	// rule3 is now local — its cache state should NOT have been touched by refresh
	// (the old entry from Phase 1 persists; ProcessEvalResults would handle it)
	rule3States = mgr.cache.getStatesForRuleUID(1, "rule3")
	require.Len(t, rule3States, 1)
}

func TestRefreshRemoteStates_NoopWhenAllRulesLocal(t *testing.T) {
	// When cluster is below min size, the filter returns all rules as local
	filter := &mockRuleFilter{
		localUIDs: map[string]struct{}{
			"rule1": {},
			"rule2": {},
		},
	}

	mgr := newTestManager(filter)

	rules := ngModels.RulesGroup{
		{UID: "rule1", OrgID: 1, Annotations: map[string]string{}},
		{UID: "rule2", OrgID: 1, Annotations: map[string]string{}},
	}

	labels1 := ngModels.InstanceLabels{"instance": "i1"}
	instances := []*ngModels.AlertInstance{
		{
			AlertInstanceKey: ngModels.AlertInstanceKey{RuleOrgID: 1, RuleUID: "rule1"},
			Labels:           labels1,
			CurrentState:     ngModels.InstanceStateFiring,
			LastEvalTime:     time.Now(),
		},
	}

	mgr.orgReader = &mockOrgReader{orgIDs: []int64{1}}
	mgr.rulesReader = &mockRuleReader{rules: rules}
	mgr.instanceReader = &fakeInstanceReader{instances: instances}

	// Should be a no-op: no remote rules to load
	mgr.refreshRemoteStates(context.Background())

	// rule1 should NOT be in cache (we didn't pre-populate, and it's local so refresh skips it)
	rule1States := mgr.cache.getStatesForRuleUID(1, "rule1")
	assert.Empty(t, rule1States)
}

func TestRunRemoteStateSync_RespectsContextCancellation(t *testing.T) {
	filter := &mockRuleFilter{localUIDs: map[string]struct{}{}}
	mgr := newTestManager(filter)
	mgr.remoteStateSyncInterval = 50 * time.Millisecond
	mgr.orgReader = &mockOrgReader{orgIDs: []int64{}}
	mgr.rulesReader = &mockRuleReader{rules: nil}
	mgr.instanceReader = &fakeInstanceReader{instances: nil}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		mgr.runRemoteStateSync(ctx)
		close(done)
	}()

	// Let it tick at least once
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good — goroutine exited
	case <-time.After(2 * time.Second):
		t.Fatal("runRemoteStateSync did not exit after context cancellation")
	}
}
