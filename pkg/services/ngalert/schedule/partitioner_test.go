package schedule

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/alerting/cluster"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// mockPeer implements notify.ClusterPeer for testing
type mockPeer struct {
	position int
	members  []string
}

func (m *mockPeer) Position() int {
	return m.position
}

func (m *mockPeer) WaitReady(ctx context.Context) error {
	return nil
}

func (m *mockPeer) AddState(key string, state cluster.State, reg prometheus.Registerer) cluster.ClusterChannel {
	return nil
}

// Members implements the Members() interface for Redis-style peers
func (m *mockPeer) Members() []string {
	return m.members
}

func TestNewPartitionFilter_NilPeer(t *testing.T) {
	filter := NewPartitionFilter(nil, 2)
	assert.Nil(t, filter, "should return nil filter when peer is nil")
}

func TestPartitioner_Filter_UnhealthyCluster(t *testing.T) {
	// Cluster with 1 member but minClusterSize is 2
	peer := &mockPeer{
		position: 0,
		members:  []string{"peer1"},
	}

	filter := NewPartitionFilter(peer, 2)
	require.NotNil(t, filter)

	rules := []*models.AlertRule{
		{UID: "rule1", OrgID: 1},
		{UID: "rule2", OrgID: 1},
		{UID: "rule3", OrgID: 1},
	}

	filtered := filter.Filter(rules)
	assert.Equal(t, len(rules), len(filtered), "unhealthy cluster should return all rules")
}

func TestPartitioner_Filter_HealthyCluster(t *testing.T) {
	// Create 3 peers and verify distribution
	members := []string{"peer0", "peer1", "peer2"}
	minClusterSize := 2

	// Generate test rules
	numRules := 100
	rules := make([]*models.AlertRule, numRules)
	for i := 0; i < numRules; i++ {
		rules[i] = &models.AlertRule{
			UID:   generateRuleUID(i),
			OrgID: 1,
		}
	}

	// Collect rules from all peers
	allAssigned := make(map[string]int) // rule UID -> peer position
	totalAssigned := 0

	for pos := 0; pos < 3; pos++ {
		peer := &mockPeer{
			position: pos,
			members:  members,
		}
		filter := NewPartitionFilter(peer, minClusterSize)
		require.NotNil(t, filter)

		filtered := filter.Filter(rules)
		totalAssigned += len(filtered)

		for _, rule := range filtered {
			// Verify no rule is assigned to multiple peers
			if existingPos, exists := allAssigned[rule.UID]; exists {
				t.Errorf("rule %s assigned to both peer %d and peer %d", rule.UID, existingPos, pos)
			}
			allAssigned[rule.UID] = pos
		}
	}

	// All rules should be assigned exactly once
	assert.Equal(t, numRules, totalAssigned, "all rules should be assigned across peers")
	assert.Equal(t, numRules, len(allAssigned), "each rule should be assigned to exactly one peer")
}

func TestPartitioner_ConsistentAssignment(t *testing.T) {
	members := []string{"peer0", "peer1", "peer2"}
	minClusterSize := 2

	rules := []*models.AlertRule{
		{UID: "consistent-rule-1", OrgID: 1},
		{UID: "consistent-rule-2", OrgID: 1},
		{UID: "consistent-rule-3", OrgID: 2},
	}

	// Run the same filter multiple times and verify consistent results
	for iteration := 0; iteration < 10; iteration++ {
		for pos := 0; pos < 3; pos++ {
			peer := &mockPeer{
				position: pos,
				members:  members,
			}
			filter := NewPartitionFilter(peer, minClusterSize)
			filtered1 := filter.Filter(rules)
			filtered2 := filter.Filter(rules)

			// Same peer should always get the same rules
			assert.Equal(t, len(filtered1), len(filtered2), "consistent assignment check failed")
			for i := range filtered1 {
				assert.Equal(t, filtered1[i].UID, filtered2[i].UID, "rules should be consistently assigned")
			}
		}
	}
}

func TestPartitioner_EvenDistribution(t *testing.T) {
	members := []string{"peer0", "peer1", "peer2"}
	minClusterSize := 2

	// Generate many rules to test distribution
	numRules := 300
	rules := make([]*models.AlertRule, numRules)
	for i := 0; i < numRules; i++ {
		rules[i] = &models.AlertRule{
			UID:   generateRuleUID(i),
			OrgID: 1,
		}
	}

	counts := make([]int, 3)
	for pos := 0; pos < 3; pos++ {
		peer := &mockPeer{
			position: pos,
			members:  members,
		}
		filter := NewPartitionFilter(peer, minClusterSize)
		filtered := filter.Filter(rules)
		counts[pos] = len(filtered)
	}

	// With 300 rules and 3 peers, expect ~100 each
	// Allow 20% variance for hash distribution
	expected := numRules / 3
	tolerance := expected / 5 // 20%

	for pos, count := range counts {
		assert.InDelta(t, expected, count, float64(tolerance),
			"peer %d got %d rules, expected ~%d (±%d)", pos, count, expected, tolerance)
	}
}

func TestPartitioner_OrgIsolation(t *testing.T) {
	members := []string{"peer0", "peer1"}
	minClusterSize := 1

	// Same rule UID in different orgs should potentially go to different peers
	rule1Org1 := &models.AlertRule{UID: "same-uid", OrgID: 1}
	rule1Org2 := &models.AlertRule{UID: "same-uid", OrgID: 2}

	peer0 := &mockPeer{position: 0, members: members}
	peer1 := &mockPeer{position: 1, members: members}

	filter0 := NewPartitionFilter(peer0, minClusterSize)
	filter1 := NewPartitionFilter(peer1, minClusterSize)

	// The key point is that orgID is included in the hash,
	// so same UID but different org should be treated as different rules
	rules := []*models.AlertRule{rule1Org1, rule1Org2}

	result0 := filter0.Filter(rules)
	result1 := filter1.Filter(rules)

	// Both rules should be assigned somewhere
	assert.Equal(t, 2, len(result0)+len(result1), "both rules should be assigned")
}

func generateRuleUID(i int) string {
	return "rule-" + string(rune('a'+i%26)) + string(rune('0'+i/26%10))
}
