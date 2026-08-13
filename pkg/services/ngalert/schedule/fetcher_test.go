package schedule

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// mockPartitioner implements RulePartitioner for testing the fetcher integration.
type mockPartitioner struct {
	mu              sync.Mutex
	clusterSize     int
	filterCallCount int
	filterFunc      func(rules []*models.AlertRule) []*models.AlertRule
	ownsFunc        func(orgID int64, ruleUID string) bool
}

func (m *mockPartitioner) ClusterSize() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clusterSize
}

func (m *mockPartitioner) Filter(rules []*models.AlertRule) []*models.AlertRule {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filterCallCount++
	if m.filterFunc != nil {
		return m.filterFunc(rules)
	}
	return rules
}

func (m *mockPartitioner) Owns(orgID int64, ruleUID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ownsFunc != nil {
		return m.ownsFunc(orgID, ruleUID)
	}
	return true
}

func (m *mockPartitioner) setClusterSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clusterSize = size
}

func (m *mockPartitioner) getFilterCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.filterCallCount
}

func (m *mockPartitioner) resetFilterCallCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filterCallCount = 0
}

func TestFetcher_TopologyChangeTriggersRefetch(t *testing.T) {
	rs := newFakeRulesStore()
	for i := 0; i < 10; i++ {
		rs.PutRule(context.Background(), &models.AlertRule{
			UID:          generateRuleUID(i),
			OrgID:        1,
			NamespaceUID: "ns1",
			Version:      1,
		})
	}

	mp := &mockPartitioner{clusterSize: 2}

	sched := setupScheduler(t, rs, nil, nil, nil, nil, nil,
		withPartitioner(mp))

	// First call — populates rules (lastClusterSize=0 → clusterSize=2, forces update)
	_, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "filter should be called on first fetch")

	// Second call — stable topology, same rules → skip fetch
	mp.resetFilterCallCount()
	d, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.True(t, d.IsEmpty(), "no changes should be detected when topology is stable")
	assert.Equal(t, 0, mp.getFilterCallCount(), "filter should NOT be called when skipping")

	// Change cluster size to simulate a member joining
	mp.setClusterSize(3)
	mp.resetFilterCallCount()

	// Third call — topology changed → force re-fetch
	_, err = sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "filter should be called after topology change")
}

func TestFetcher_TopologyStableSkipsRefetch(t *testing.T) {
	rs := newFakeRulesStore()
	for i := 0; i < 5; i++ {
		rs.PutRule(context.Background(), &models.AlertRule{
			UID:          generateRuleUID(i),
			OrgID:        1,
			NamespaceUID: "ns1",
			Version:      1,
		})
	}

	mp := &mockPartitioner{clusterSize: 2}

	sched := setupScheduler(t, rs, nil, nil, nil, nil, nil,
		withPartitioner(mp))

	// First call populates
	_, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)

	// Multiple calls with stable topology should all skip
	for i := 0; i < 5; i++ {
		mp.resetFilterCallCount()
		d, err := sched.updateSchedulableAlertRules(context.Background())
		require.NoError(t, err)
		assert.True(t, d.IsEmpty(), "call %d should return empty diff", i)
		assert.Equal(t, 0, mp.getFilterCallCount(), "call %d should not invoke filter", i)
	}
}

func TestFetcher_PartitionedKeysFastPath(t *testing.T) {
	// The registry only holds this peer's share of rules, so the cheap-keys check must
	// narrow the key list with Owns before comparing. Without that, the length mismatch
	// forces a full re-fetch on every tick and the fast path never fires.
	rs := newFakeRulesStore()
	ownedUIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		uid := generateRuleUID(i)
		rs.PutRule(context.Background(), &models.AlertRule{
			UID:          uid,
			OrgID:        1,
			NamespaceUID: "ns1",
			Version:      1,
		})
		if i%2 == 0 {
			ownedUIDs[uid] = true
		}
	}

	// This peer owns only half of the rules.
	owns := func(orgID int64, ruleUID string) bool {
		return ownedUIDs[ruleUID]
	}
	mp := &mockPartitioner{
		clusterSize: 2,
		ownsFunc:    owns,
		filterFunc: func(rules []*models.AlertRule) []*models.AlertRule {
			filtered := make([]*models.AlertRule, 0, len(rules))
			for _, r := range rules {
				if owns(r.OrgID, r.UID) {
					filtered = append(filtered, r)
				}
			}
			return filtered
		},
	}

	sched := setupScheduler(t, rs, nil, nil, nil, nil, nil,
		withPartitioner(mp))

	// First call — populates the registry with this peer's share only.
	_, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "filter should be called on first fetch")

	// Second call — nothing changed. The fast path must detect this from the narrowed
	// key set and skip the full fetch entirely.
	mp.resetFilterCallCount()
	d, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.True(t, d.IsEmpty(), "stable partitioned registry should return an empty diff")
	assert.Equal(t, 0, mp.getFilterCallCount(), "stable partitioned registry should not re-fetch or re-filter")
}

func TestFetcher_TopologyChangeWithRuleChanges(t *testing.T) {
	rs := newFakeRulesStore()
	for i := 0; i < 5; i++ {
		rs.PutRule(context.Background(), &models.AlertRule{
			UID:          generateRuleUID(i),
			OrgID:        1,
			NamespaceUID: "ns1",
			Version:      1,
		})
	}

	mp := &mockPartitioner{clusterSize: 2}

	sched := setupScheduler(t, rs, nil, nil, nil, nil, nil,
		withPartitioner(mp))

	// First call — populates
	_, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)

	// Simultaneously change topology AND add a new rule
	mp.setClusterSize(3)
	rs.PutRule(context.Background(), &models.AlertRule{
		UID:          "new-rule",
		OrgID:        1,
		NamespaceUID: "ns1",
		Version:      1,
	})
	mp.resetFilterCallCount()

	_, err = sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "filter should be called after topology+rule change")
}

func TestFetcher_MemberLeaveThenJoin(t *testing.T) {
	rs := newFakeRulesStore()
	for i := 0; i < 10; i++ {
		rs.PutRule(context.Background(), &models.AlertRule{
			UID:          generateRuleUID(i),
			OrgID:        1,
			NamespaceUID: "ns1",
			Version:      1,
		})
	}

	mp := &mockPartitioner{clusterSize: 3}

	sched := setupScheduler(t, rs, nil, nil, nil, nil, nil,
		withPartitioner(mp))

	// Populate
	_, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)

	// Member leaves: 3 → 2
	mp.setClusterSize(2)
	mp.resetFilterCallCount()
	_, err = sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "should re-filter after member leaves")

	// Stable at 2
	mp.resetFilterCallCount()
	d, err := sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.True(t, d.IsEmpty(), "should skip when stable at 2")
	assert.Equal(t, 0, mp.getFilterCallCount())

	// Member joins: 2 → 3
	mp.setClusterSize(3)
	mp.resetFilterCallCount()
	_, err = sched.updateSchedulableAlertRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, mp.getFilterCallCount(), "should re-filter after member joins")
}
