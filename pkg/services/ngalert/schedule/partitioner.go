package schedule

import (
	"fmt"
	"hash/fnv"

	"github.com/grafana/alerting/cluster"
	"github.com/grafana/alerting/notify"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// RulePartitioner provides rule filtering and cluster size tracking for HA partitioning.
type RulePartitioner interface {
	// Filter returns the subset of rules this peer should evaluate.
	Filter(rules []*models.AlertRule) []*models.AlertRule
	// Owns reports whether this peer is responsible for the given rule right now.
	// It returns true for every rule when partitioning is not currently applicable,
	// matching Filter's evaluate-all fallback.
	Owns(orgID int64, ruleUID string) bool
	// ClusterSize returns the number of peers currently participating in partitioning.
	// This is used to detect topology changes that require rule redistribution.
	ClusterSize() int
}

// memberLister is satisfied by the Redis peer, which has no ClusterSize but exposes
// its alive members.
type memberLister interface{ Members() []string }

// NewPartitionFilter creates a RulePartitioner that partitions rules across cluster members.
// It returns nil when the peer cannot report cluster membership — which is the case for
// notifier.NilPeer, i.e. when HA clustering is not configured. Returning nil here (rather
// than defaulting to a cluster size of 1) means a single-node deployment never enters the
// partitioning code paths at all.
func NewPartitionFilter(peer notify.ClusterPeer, minClusterSize int) RulePartitioner {
	if peer == nil {
		return nil
	}
	logger := log.New("ngalert.scheduler.partitioner")

	var count func() int
	switch p := peer.(type) {
	case *cluster.Peer:
		// Memberlist peer: ClusterSize() counts alive members, consistent with Position().
		count = p.ClusterSize
	case memberLister:
		// Redis peer: len(Members()) counts alive members only. Deliberately NOT the Redis
		// peer's own ClusterSize(), which also counts dead nodes that have not timed out
		// yet and would therefore be inconsistent with Position().
		count = func() int { return len(p.Members()) }
	default:
		logger.Warn("Peer type cannot report cluster membership; HA scheduler partitioning disabled",
			"peerType", fmt.Sprintf("%T", peer))
		return nil
	}

	if minClusterSize < 2 {
		logger.Warn("ha_scheduler_min_cluster_size below 2 is meaningless; clamping to 2",
			"configured", minClusterSize)
		minClusterSize = 2
	}

	return &partitioner{peer: peer, memberCount: count, minClusterSize: minClusterSize, logger: logger}
}

type partitioner struct {
	peer           notify.ClusterPeer
	memberCount    func() int
	minClusterSize int
	logger         log.Logger
}

func (p *partitioner) ClusterSize() int {
	return p.memberCount()
}

// snapshot reads cluster size and peer position once. ok is false when partitioning must
// not be applied: cluster below minimum size, or a position that does not fit the member
// count (which happens transiently during churn, and when the Redis peer's Position()
// falls back to 0 on lookup failure). Callers must then treat every rule as local.
func (p *partitioner) snapshot() (size, position int, ok bool) {
	size = p.memberCount()
	position = p.peer.Position()
	if size < p.minClusterSize {
		return size, position, false
	}
	if position < 0 || position >= size {
		p.logger.Warn("Peer position does not fit cluster size; evaluating all rules",
			"position", position, "clusterSize", size)
		return size, position, false
	}
	return size, position, true
}

func assignedTo(orgID int64, ruleUID string, size int) int {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%d:%s", orgID, ruleUID)
	return int(h.Sum64() % uint64(size))
}

// Owns reports whether the local peer is responsible for a rule right now.
func (p *partitioner) Owns(orgID int64, ruleUID string) bool {
	size, position, ok := p.snapshot()
	if !ok {
		return true
	}
	return assignedTo(orgID, ruleUID, size) == position
}

// Filter returns the subset of rules this peer should evaluate. Size and position are
// read once so that every rule in one pass is hashed against the same cluster view.
// If the cluster is unhealthy (below minimum size), all rules are returned as a safety
// fallback to ensure all rules are evaluated.
func (p *partitioner) Filter(rules []*models.AlertRule) []*models.AlertRule {
	size, position, ok := p.snapshot()
	if !ok {
		p.logger.Info("HA partitioning not applied, evaluating all rules",
			"clusterSize", size,
			"minClusterSize", p.minClusterSize,
			"position", position,
			"totalRules", len(rules),
		)
		return rules
	}

	filtered := make([]*models.AlertRule, 0, len(rules)/size+1)
	for _, rule := range rules {
		if assignedTo(rule.OrgID, rule.UID, size) == position {
			filtered = append(filtered, rule)
		}
	}

	p.logger.Debug("HA partitioning applied",
		"totalRules", len(rules),
		"assignedRules", len(filtered),
		"clusterSize", size,
		"position", position,
	)

	return filtered
}
