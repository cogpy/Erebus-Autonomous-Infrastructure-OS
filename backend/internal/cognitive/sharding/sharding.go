package sharding

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
)

// Shard represents a partition of the AtomSpace
type Shard struct {
	ID        int
	AtomSpace *atomspace.AtomSpace
	Load      int64 // Current number of atoms in this shard
	LastUsed  time.Time
	mu        sync.RWMutex
}

// ShardManager manages dynamic sharding of atoms across multiple AtomSpaces
type ShardManager struct {
	shards       []*Shard
	numShards    int
	rebalanceThreshold int64 // Rebalance when difference exceeds this
	mu           sync.RWMutex
	
	// Channels for concurrent shard operations
	routeChan    chan routeRequest
	rebalanceChan chan struct{}
	done         chan struct{}
}

type routeRequest struct {
	atomID   string
	tenantID string
	response chan int // Returns shard ID
}

// NewShardManager creates a new shard manager with dynamic sharding
func NewShardManager(numShards int, workers int) *ShardManager {
	sm := &ShardManager{
		shards:             make([]*Shard, numShards),
		numShards:          numShards,
		rebalanceThreshold: 1000,
		routeChan:          make(chan routeRequest, 1000),
		rebalanceChan:      make(chan struct{}, 1),
		done:               make(chan struct{}),
	}
	
	// Initialize shards
	for i := 0; i < numShards; i++ {
		sm.shards[i] = &Shard{
			ID:        i,
			AtomSpace: atomspace.NewAtomSpace(workers / numShards),
			Load:      0,
			LastUsed:  time.Now(),
		}
	}
	
	// Start router workers
	for i := 0; i < workers; i++ {
		go sm.routerWorker()
	}
	
	// Start rebalancing monitor
	go sm.rebalanceMonitor()
	
	return sm
}

// routerWorker handles routing requests concurrently
func (sm *ShardManager) routerWorker() {
	for {
		select {
		case req := <-sm.routeChan:
			shardID := sm.getShardIDInternal(req.atomID, req.tenantID)
			req.response <- shardID
		case <-sm.done:
			return
		}
	}
}

// rebalanceMonitor periodically checks for rebalancing needs
func (sm *ShardManager) rebalanceMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if sm.needsRebalance() {
				select {
				case sm.rebalanceChan <- struct{}{}:
					go sm.rebalance()
				default:
					// Rebalance already in progress
				}
			}
		case <-sm.done:
			return
		}
	}
}

// GetShardID returns the shard ID for a given atom (consistent hashing)
func (sm *ShardManager) GetShardID(atomID, tenantID string) int {
	response := make(chan int, 1)
	sm.routeChan <- routeRequest{atomID: atomID, tenantID: tenantID, response: response}
	return <-response
}

// getShardIDInternal is the internal implementation
func (sm *ShardManager) getShardIDInternal(atomID, tenantID string) int {
	// Consistent hashing with tenant isolation
	h := fnv.New64a()
	h.Write([]byte(tenantID + ":" + atomID))
	hash := h.Sum64()
	return int(hash % uint64(sm.numShards))
}

// GetShard returns the shard for a given atom
func (sm *ShardManager) GetShard(atomID, tenantID string) *Shard {
	shardID := sm.GetShardID(atomID, tenantID)
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.shards[shardID]
}

// GetShardByID returns a shard by its ID
func (sm *ShardManager) GetShardByID(shardID int) (*Shard, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if shardID < 0 || shardID >= sm.numShards {
		return nil, fmt.Errorf("invalid shard ID: %d", shardID)
	}
	
	return sm.shards[shardID], nil
}

// AddAtom adds an atom to the appropriate shard
func (sm *ShardManager) AddAtom(atom atomspace.Atom) error {
	shard := sm.GetShard(atom.GetID(), atom.GetTenantID())
	
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	err := shard.AtomSpace.AddAtom(atom)
	if err == nil {
		shard.Load++
		shard.LastUsed = time.Now()
	}
	
	return err
}

// GetAtom retrieves an atom from the appropriate shard
func (sm *ShardManager) GetAtom(atomID, tenantID string) (atomspace.Atom, error) {
	shard := sm.GetShard(atomID, tenantID)
	return shard.AtomSpace.GetAtom(atomID, tenantID)
}

// QueryAtoms queries atoms across all shards for a tenant
func (sm *ShardManager) QueryAtoms(tenantID string, filter func(atomspace.Atom) bool) []atomspace.Atom {
	sm.mu.RLock()
	numShards := len(sm.shards)
	sm.mu.RUnlock()
	
	// Parallel query across all shards
	type shardResult struct {
		atoms []atomspace.Atom
	}
	
	resultChan := make(chan shardResult, numShards)
	
	for i := 0; i < numShards; i++ {
		go func(shardID int) {
			shard, _ := sm.GetShardByID(shardID)
			atoms := shard.AtomSpace.QueryAtoms(tenantID, filter)
			resultChan <- shardResult{atoms: atoms}
		}(i)
	}
	
	// Collect results
	var allAtoms []atomspace.Atom
	for i := 0; i < numShards; i++ {
		result := <-resultChan
		allAtoms = append(allAtoms, result.atoms...)
	}
	
	return allAtoms
}

// UpdateAtom updates an atom in the appropriate shard
func (sm *ShardManager) UpdateAtom(atomID, tenantID string, updater func(atomspace.Atom) error) error {
	shard := sm.GetShard(atomID, tenantID)
	return shard.AtomSpace.UpdateAtom(atomID, tenantID, updater)
}

// DeleteAtom deletes an atom from the appropriate shard
func (sm *ShardManager) DeleteAtom(atomID, tenantID string) error {
	shard := sm.GetShard(atomID, tenantID)
	
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	err := shard.AtomSpace.DeleteAtom(atomID, tenantID)
	if err == nil {
		shard.Load--
	}
	
	return err
}

// needsRebalance checks if shards need rebalancing
func (sm *ShardManager) needsRebalance() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if len(sm.shards) == 0 {
		return false
	}
	
	minLoad := sm.shards[0].Load
	maxLoad := sm.shards[0].Load
	
	for _, shard := range sm.shards {
		shard.mu.RLock()
		load := shard.Load
		shard.mu.RUnlock()
		
		if load < minLoad {
			minLoad = load
		}
		if load > maxLoad {
			maxLoad = load
		}
	}
	
	return (maxLoad - minLoad) > sm.rebalanceThreshold
}

// rebalance redistributes atoms across shards for better load balancing
func (sm *ShardManager) rebalance() {
	// This is a simplified rebalancing implementation
	// In a production system, this would involve more sophisticated algorithms
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Find overloaded and underloaded shards
	var overloaded, underloaded []*Shard
	avgLoad := int64(0)
	
	for _, shard := range sm.shards {
		shard.mu.RLock()
		avgLoad += shard.Load
		shard.mu.RUnlock()
	}
	avgLoad /= int64(len(sm.shards))
	
	for _, shard := range sm.shards {
		shard.mu.RLock()
		load := shard.Load
		shard.mu.RUnlock()
		
		if load > avgLoad+sm.rebalanceThreshold/2 {
			overloaded = append(overloaded, shard)
		} else if load < avgLoad-sm.rebalanceThreshold/2 {
			underloaded = append(underloaded, shard)
		}
	}
	
	// Note: Actual atom migration would happen here
	// For now, we just log that rebalancing would occur
	if len(overloaded) > 0 && len(underloaded) > 0 {
		// In production, migrate atoms from overloaded to underloaded shards
		// This requires careful handling to maintain consistency
	}
	
	// Drain the rebalance channel
	select {
	case <-sm.rebalanceChan:
	default:
	}
}

// GetShardStats returns statistics for all shards
func (sm *ShardManager) GetShardStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	shardStats := make([]map[string]interface{}, len(sm.shards))
	totalLoad := int64(0)
	
	for i, shard := range sm.shards {
		shard.mu.RLock()
		load := shard.Load
		lastUsed := shard.LastUsed
		shard.mu.RUnlock()
		
		shardStats[i] = map[string]interface{}{
			"shard_id":  shard.ID,
			"load":      load,
			"last_used": lastUsed,
		}
		totalLoad += load
	}
	
	avgLoad := int64(0)
	if len(sm.shards) > 0 {
		avgLoad = totalLoad / int64(len(sm.shards))
	}
	
	return map[string]interface{}{
		"num_shards":   sm.numShards,
		"total_load":   totalLoad,
		"average_load": avgLoad,
		"shards":       shardStats,
	}
}

// Close shuts down the shard manager and all shards
func (sm *ShardManager) Close() {
	close(sm.done)
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	for _, shard := range sm.shards {
		shard.AtomSpace.Close()
	}
}

// GetTenantStats returns statistics for a specific tenant across all shards
func (sm *ShardManager) GetTenantStats(tenantID string) map[string]interface{} {
	sm.mu.RLock()
	numShards := len(sm.shards)
	sm.mu.RUnlock()
	
	type shardTenantStats struct {
		shardID int
		stats   map[string]interface{}
	}
	
	resultChan := make(chan shardTenantStats, numShards)
	
	for i := 0; i < numShards; i++ {
		go func(shardID int) {
			shard, _ := sm.GetShardByID(shardID)
			stats := shard.AtomSpace.GetStats(tenantID)
			resultChan <- shardTenantStats{shardID: shardID, stats: stats}
		}(i)
	}
	
	// Aggregate results
	totalAtoms := 0
	atomsByType := make(map[atomspace.AtomType]int)
	shardDistribution := make(map[int]int)
	
	for i := 0; i < numShards; i++ {
		result := <-resultChan
		stats := result.stats
		
		if total, ok := stats["total_atoms"].(int); ok {
			totalAtoms += total
			shardDistribution[result.shardID] = total
		}
		
		if typeMap, ok := stats["atoms_by_type"].(map[atomspace.AtomType]int); ok {
			for atomType, count := range typeMap {
				atomsByType[atomType] += count
			}
		}
	}
	
	return map[string]interface{}{
		"tenant_id":          tenantID,
		"total_atoms":        totalAtoms,
		"atoms_by_type":      atomsByType,
		"shard_distribution": shardDistribution,
	}
}
