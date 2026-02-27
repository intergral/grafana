package schedule

import (
	"context"
	"fmt"
	"sync"
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

// dynamicMockPeer implements notify.ClusterPeer with mutable membership.
// Models real gossip ring behavior where members join/leave dynamically.
type dynamicMockPeer struct {
	mu       sync.Mutex
	position int
	members  []string
}

func (m *dynamicMockPeer) Position() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.position
}

func (m *dynamicMockPeer) WaitReady(ctx context.Context) error {
	return nil
}

func (m *dynamicMockPeer) AddState(key string, state cluster.State, reg prometheus.Registerer) cluster.ClusterChannel {
	return nil
}

func (m *dynamicMockPeer) Members() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.members...)
}

func (m *dynamicMockPeer) setTopology(position int, members []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.position = position
	m.members = members
}

// generateRules creates n test rules with unique UIDs.
func generateRules(n int) []*models.AlertRule {
	rules := make([]*models.AlertRule, n)
	for i := 0; i < n; i++ {
		rules[i] = &models.AlertRule{
			UID:   fmt.Sprintf("rule-%04d", i),
			OrgID: 1,
		}
	}
	return rules
}

// memberNames returns a slice of member name strings for n members.
func memberNames(n int) []string {
	members := make([]string, n)
	for i := range members {
		members[i] = fmt.Sprintf("peer%d", i)
	}
	return members
}

// assertFullCoverageWithPeers filters rules through all partitioners and asserts every
// rule is assigned to exactly one peer.
func assertFullCoverageWithPeers(t *testing.T, rules []*models.AlertRule, partitioners []RulePartitioner, description string) {
	t.Helper()
	assigned := make(map[string]int)
	totalAssigned := 0

	for i, p := range partitioners {
		filtered := p.Filter(rules)
		totalAssigned += len(filtered)
		for _, rule := range filtered {
			key := fmt.Sprintf("%d:%s", rule.OrgID, rule.UID)
			if existingPeer, exists := assigned[key]; exists {
				t.Errorf("[%s] rule %s (org %d) assigned to both peer %d and peer %d",
					description, rule.UID, rule.OrgID, existingPeer, i)
			}
			assigned[key] = i
		}
	}

	assert.Equal(t, len(rules), totalAssigned, "[%s] all rules should be assigned", description)
	assert.Equal(t, len(rules), len(assigned), "[%s] each rule should be assigned exactly once", description)
}

func TestPartitioner_MemberJoins(t *testing.T) {
	rules := generateRules(100)
	minClusterSize := 1

	// Phase 1: 2-member cluster
	members2 := memberNames(2)
	peers := make([]*dynamicMockPeer, 2)
	partitioners := make([]RulePartitioner, 2)
	for i := 0; i < 2; i++ {
		peers[i] = &dynamicMockPeer{position: i, members: members2}
		partitioners[i] = NewPartitionFilter(peers[i], minClusterSize)
	}
	assertFullCoverageWithPeers(t, rules, partitioners, "2-member cluster")

	// Phase 2: Third member joins — update existing peers and add new one
	members3 := memberNames(3)
	for i, p := range peers {
		p.setTopology(i, members3)
	}
	newPeer := &dynamicMockPeer{position: 2, members: members3}
	peers = append(peers, newPeer)
	partitioners = append(partitioners, NewPartitionFilter(newPeer, minClusterSize))

	assertFullCoverageWithPeers(t, rules, partitioners, "3-member cluster after join")
}

func TestPartitioner_MemberLeaves(t *testing.T) {
	rules := generateRules(100)
	minClusterSize := 1

	// Phase 1: 3-member cluster
	members3 := memberNames(3)
	peers := make([]*dynamicMockPeer, 3)
	partitioners := make([]RulePartitioner, 3)
	for i := 0; i < 3; i++ {
		peers[i] = &dynamicMockPeer{position: i, members: members3}
		partitioners[i] = NewPartitionFilter(peers[i], minClusterSize)
	}
	assertFullCoverageWithPeers(t, rules, partitioners, "3-member cluster")

	// Phase 2: Last member leaves — update remaining peers
	members2 := memberNames(2)
	remainingPartitioners := partitioners[:2]
	for i := 0; i < 2; i++ {
		peers[i].setTopology(i, members2)
	}

	assertFullCoverageWithPeers(t, rules, remainingPartitioners, "2-member cluster after leave")
}

func TestPartitioner_HealthyToUnhealthy(t *testing.T) {
	rules := generateRules(50)
	minClusterSize := 3

	// Phase 1: Healthy 3-member cluster
	members3 := memberNames(3)
	peers := make([]*dynamicMockPeer, 3)
	partitioners := make([]RulePartitioner, 3)
	for i := 0; i < 3; i++ {
		peers[i] = &dynamicMockPeer{position: i, members: members3}
		partitioners[i] = NewPartitionFilter(peers[i], minClusterSize)
	}
	assertFullCoverageWithPeers(t, rules, partitioners, "healthy 3-member cluster")

	// Phase 2: Member leaves — cluster drops below minClusterSize (unhealthy)
	members2 := memberNames(2)
	for i := 0; i < 2; i++ {
		peers[i].setTopology(i, members2)
	}

	// Each remaining peer should now return ALL rules (safety fallback)
	for i := 0; i < 2; i++ {
		filtered := partitioners[i].Filter(rules)
		assert.Equal(t, len(rules), len(filtered),
			"peer %d should return all rules when cluster is unhealthy", i)
	}
}

func TestPartitioner_UnhealthyToHealthy(t *testing.T) {
	rules := generateRules(50)
	minClusterSize := 2

	// Phase 1: Unhealthy 1-member cluster
	members1 := memberNames(1)
	peer0 := &dynamicMockPeer{position: 0, members: members1}
	p0 := NewPartitionFilter(peer0, minClusterSize)

	filtered := p0.Filter(rules)
	assert.Equal(t, len(rules), len(filtered), "unhealthy cluster should return all rules")

	// Phase 2: Second member joins — cluster becomes healthy
	members2 := memberNames(2)
	peer0.setTopology(0, members2)
	peer1 := &dynamicMockPeer{position: 1, members: members2}
	p1 := NewPartitionFilter(peer1, minClusterSize)

	partitioners := []RulePartitioner{p0, p1}
	assertFullCoverageWithPeers(t, rules, partitioners, "healthy 2-member cluster after recovery")

	// Each peer should get a subset (not all rules)
	for i, p := range partitioners {
		filtered := p.Filter(rules)
		assert.Less(t, len(filtered), len(rules),
			"peer %d should get a subset of rules when cluster is healthy", i)
	}
}

func TestPartitioner_RapidTopologyChanges(t *testing.T) {
	rules := generateRules(200)
	minClusterSize := 1

	// Cycle through various cluster sizes
	sizes := []int{2, 3, 5, 2, 4, 1}

	for _, size := range sizes {
		members := memberNames(size)
		partitioners := make([]RulePartitioner, size)
		for i := 0; i < size; i++ {
			peer := &mockPeer{position: i, members: members}
			partitioners[i] = NewPartitionFilter(peer, minClusterSize)
		}
		assertFullCoverageWithPeers(t, rules, partitioners,
			fmt.Sprintf("%d-member cluster", size))
	}
}
