package schedule

import (
	"fmt"
	"hash/fnv"

	"github.com/grafana/alerting/cluster"
	"github.com/grafana/alerting/notify"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// RuleFilterFunc filters rules for this instance to evaluate.
// Returns the subset of rules this peer should evaluate.
type RuleFilterFunc func(rules []*models.AlertRule) []*models.AlertRule

// RulePartitioner provides rule filtering and cluster size tracking for HA partitioning.
type RulePartitioner interface {
	Filter(rules []*models.AlertRule) []*models.AlertRule
	ClusterSize() int
}

// NewPartitionFilter creates a RulePartitioner that partitions rules across cluster members.
// Returns nil if partitioning should not be enabled (no peer or unhealthy cluster).
func NewPartitionFilter(peer notify.ClusterPeer, minClusterSize int) RulePartitioner {
	if peer == nil {
		return nil
	}

	p := &partitioner{
		peer:           peer,
		minClusterSize: minClusterSize,
		logger:         log.New("ngalert.scheduler.partitioner"),
	}

	return p
}

type partitioner struct {
	peer           notify.ClusterPeer
	minClusterSize int
	logger         log.Logger
}

func (p *partitioner) position() int {
	return p.peer.Position()
}

// ClusterSize returns the current cluster size.
// This is used to detect topology changes that require rule redistribution.
func (p *partitioner) ClusterSize() int {
	return p.memberCount()
}

func (p *partitioner) memberCount() int {
	// Try memberlist Peer first
	if mp, ok := p.peer.(*cluster.Peer); ok {
		size := mp.ClusterSize()
		p.logger.Debug("Memberlist peer detected", "clusterSize", size)
		return size
	}
	// Try Redis peer (has Members() method)
	if rp, ok := p.peer.(interface{ Members() []string }); ok {
		members := rp.Members()
		p.logger.Debug("Redis peer detected", "members", len(members), "memberList", members)
		return len(members)
	}
	p.logger.Warn("Unknown peer type, defaulting to cluster size 1")
	return 1
}

func (p *partitioner) isHealthy() bool {
	return p.memberCount() >= p.minClusterSize
}

func (p *partitioner) shouldEvaluate(ruleUID string, orgID int64) bool {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%d:%s", orgID, ruleUID)
	return int(h.Sum64()%uint64(p.memberCount())) == p.position()
}

// Filter returns the subset of rules this peer should evaluate.
// If the cluster is unhealthy (below minimum size), all rules are returned
// as a safety fallback to ensure all rules are evaluated.
func (p *partitioner) Filter(rules []*models.AlertRule) []*models.AlertRule {
	memberCount := p.memberCount()
	position := p.position()
	healthy := p.isHealthy()

	p.logger.Debug("Filtering rules for HA partitioning",
		"totalRules", len(rules),
		"memberCount", memberCount,
		"position", position,
		"minClusterSize", p.minClusterSize,
		"healthy", healthy,
	)

	if !healthy {
		p.logger.Info("Cluster below minimum size, evaluating all rules",
			"memberCount", memberCount,
			"minClusterSize", p.minClusterSize,
		)
		return rules // fallback: evaluate all when cluster too small
	}

	filtered := make([]*models.AlertRule, 0, len(rules)/memberCount+1)
	for _, rule := range rules {
		if p.shouldEvaluate(rule.UID, rule.OrgID) {
			filtered = append(filtered, rule)
		}
	}

	p.logger.Info("HA partitioning applied",
		"totalRules", len(rules),
		"assignedRules", len(filtered),
		"memberCount", memberCount,
		"position", position,
	)

	return filtered
}
