package cognitive

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive/agents"
	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
	"github.com/Avik2024/erebus/backend/internal/cognitive/inference"
	"github.com/Avik2024/erebus/backend/internal/cognitive/pipeline"
	"github.com/Avik2024/erebus/backend/internal/cognitive/sharding"
)

// CognitiveEngine is the main orchestrator for the OpenCog-inspired cognitive architecture
type CognitiveEngine struct {
	shardManager  *sharding.ShardManager
	inferenceEngines map[string]*inference.InferenceEngine // tenantID -> engine
	agentScheduler   *agents.AgentScheduler
	pipelineOrch     *pipeline.PipelineOrchestrator
	
	// Configuration
	numShards     int
	workersPerShard int
	inferenceWorkers int
	agentWorkers     int
	pipelineWorkers  int
	
	mu sync.RWMutex
	done chan struct{}
}

// Config holds configuration for the cognitive engine
type Config struct {
	NumShards        int
	WorkersPerShard  int
	InferenceWorkers int
	AgentWorkers     int
	PipelineWorkers  int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		NumShards:        8,
		WorkersPerShard:  4,
		InferenceWorkers: 16,
		AgentWorkers:     8,
		PipelineWorkers:  8,
	}
}

// NewCognitiveEngine creates a new cognitive engine with the specified configuration
func NewCognitiveEngine(cfg *Config) *CognitiveEngine {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	
	ce := &CognitiveEngine{
		shardManager:     sharding.NewShardManager(cfg.NumShards, cfg.WorkersPerShard*cfg.NumShards),
		inferenceEngines: make(map[string]*inference.InferenceEngine),
		agentScheduler:   agents.NewAgentScheduler(cfg.AgentWorkers),
		pipelineOrch:     pipeline.NewPipelineOrchestrator(cfg.PipelineWorkers),
		numShards:        cfg.NumShards,
		workersPerShard:  cfg.WorkersPerShard,
		inferenceWorkers: cfg.InferenceWorkers,
		agentWorkers:     cfg.AgentWorkers,
		pipelineWorkers:  cfg.PipelineWorkers,
		done:            make(chan struct{}),
	}
	
	return ce
}

// InitializeTenant initializes cognitive resources for a new tenant
func (ce *CognitiveEngine) InitializeTenant(tenantID string) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	
	// Check if tenant already exists
	if _, exists := ce.inferenceEngines[tenantID]; exists {
		return fmt.Errorf("tenant %s already initialized", tenantID)
	}
	
	// Create a tenant-specific atomspace wrapper that queries across shards
	tenantAtomSpace := &tenantAtomSpaceWrapper{
		shardManager: ce.shardManager,
		tenantID:     tenantID,
	}
	
	// Create inference engine for this tenant
	inferenceEngine := inference.NewInferenceEngine(tenantAtomSpace, ce.inferenceWorkers)
	
	// Add default inference rules
	inferenceEngine.AddRule(inference.NewDeductionRule())
	inferenceEngine.AddRule(inference.NewInductionRule())
	inferenceEngine.AddRule(inference.NewAbductionRule())
	
	ce.inferenceEngines[tenantID] = inferenceEngine
	
	// Create default mind agent for this tenant
	mindAgent := agents.NewMindAgent(
		fmt.Sprintf("mind-%s", tenantID),
		"MindAgent",
		tenantID,
		tenantAtomSpace,
		inferenceEngine,
	)
	
	ce.agentScheduler.RegisterAgent(mindAgent)
	
	return nil
}

// tenantAtomSpaceWrapper wraps the shard manager to provide atomspace interface for a tenant
type tenantAtomSpaceWrapper struct {
	shardManager *sharding.ShardManager
	tenantID     string
}

func (w *tenantAtomSpaceWrapper) AddAtom(atom atomspace.Atom) error {
	return w.shardManager.AddAtom(atom)
}

func (w *tenantAtomSpaceWrapper) GetAtom(atomID, tenantID string) (atomspace.Atom, error) {
	return w.shardManager.GetAtom(atomID, tenantID)
}

func (w *tenantAtomSpaceWrapper) QueryAtoms(tenantID string, filter func(atomspace.Atom) bool) []atomspace.Atom {
	return w.shardManager.QueryAtoms(tenantID, filter)
}

func (w *tenantAtomSpaceWrapper) UpdateAtom(atomID, tenantID string, updater func(atomspace.Atom) error) error {
	return w.shardManager.UpdateAtom(atomID, tenantID, updater)
}

func (w *tenantAtomSpaceWrapper) DeleteAtom(atomID, tenantID string) error {
	return w.shardManager.DeleteAtom(atomID, tenantID)
}

func (w *tenantAtomSpaceWrapper) GetStats(tenantID string) map[string]interface{} {
	return w.shardManager.GetTenantStats(tenantID)
}

// AddAtom adds an atom to the cognitive engine
func (ce *CognitiveEngine) AddAtom(atom atomspace.Atom) error {
	return ce.shardManager.AddAtom(atom)
}

// GetAtom retrieves an atom
func (ce *CognitiveEngine) GetAtom(atomID, tenantID string) (atomspace.Atom, error) {
	return ce.shardManager.GetAtom(atomID, tenantID)
}

// QueryAtoms queries atoms for a tenant
func (ce *CognitiveEngine) QueryAtoms(tenantID string, filter func(atomspace.Atom) bool) []atomspace.Atom {
	return ce.shardManager.QueryAtoms(tenantID, filter)
}

// UpdateAtom updates an atom
func (ce *CognitiveEngine) UpdateAtom(atomID, tenantID string, updater func(atomspace.Atom) error) error {
	return ce.shardManager.UpdateAtom(atomID, tenantID, updater)
}

// DeleteAtom deletes an atom
func (ce *CognitiveEngine) DeleteAtom(atomID, tenantID string) error {
	return ce.shardManager.DeleteAtom(atomID, tenantID)
}

// RunInference runs inference for a tenant
func (ce *CognitiveEngine) RunInference(ctx context.Context, tenantID string, maxIterations int) ([]atomspace.Atom, error) {
	ce.mu.RLock()
	inferenceEngine, exists := ce.inferenceEngines[tenantID]
	ce.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("tenant %s not initialized", tenantID)
	}
	
	return inferenceEngine.RunInference(ctx, tenantID, maxIterations)
}

// CreatePipeline creates a new cognitive pipeline
func (ce *CognitiveEngine) CreatePipeline(pipelineID, name, tenantID string) (*pipeline.Pipeline, error) {
	p := pipeline.NewPipeline(pipelineID, name, tenantID)
	
	if err := ce.pipelineOrch.CreatePipeline(p); err != nil {
		return nil, err
	}
	
	return p, nil
}

// AddPipelineStage adds a stage to a pipeline
func (ce *CognitiveEngine) AddPipelineStage(pipelineID string, stage pipeline.PipelineStage) error {
	p, err := ce.pipelineOrch.GetPipeline(pipelineID)
	if err != nil {
		return err
	}
	
	p.AddStage(stage)
	return nil
}

// ExecutePipeline executes a pipeline
func (ce *CognitiveEngine) ExecutePipeline(ctx context.Context, pipelineID string, input interface{}) (interface{}, error) {
	return ce.pipelineOrch.ExecutePipeline(ctx, pipelineID, input)
}

// GetPipeline retrieves a pipeline
func (ce *CognitiveEngine) GetPipeline(pipelineID string) (*pipeline.Pipeline, error) {
	return ce.pipelineOrch.GetPipeline(pipelineID)
}

// RegisterAgent registers a cognitive agent
func (ce *CognitiveEngine) RegisterAgent(agent agents.Agent) {
	ce.agentScheduler.RegisterAgent(agent)
}

// UnregisterAgent unregisters an agent
func (ce *CognitiveEngine) UnregisterAgent(agentID string) {
	ce.agentScheduler.UnregisterAgent(agentID)
}

// GetAgent retrieves an agent
func (ce *CognitiveEngine) GetAgent(agentID string) (agents.Agent, bool) {
	return ce.agentScheduler.GetAgent(agentID)
}

// GetAgentsByTenant retrieves all agents for a tenant
func (ce *CognitiveEngine) GetAgentsByTenant(tenantID string) []agents.Agent {
	return ce.agentScheduler.GetAgentsByTenant(tenantID)
}

// GetStats returns comprehensive statistics about the cognitive engine
func (ce *CognitiveEngine) GetStats(tenantID string) map[string]interface{} {
	stats := map[string]interface{}{
		"config": map[string]interface{}{
			"num_shards":        ce.numShards,
			"workers_per_shard": ce.workersPerShard,
			"inference_workers": ce.inferenceWorkers,
			"agent_workers":     ce.agentWorkers,
			"pipeline_workers":  ce.pipelineWorkers,
		},
		"sharding": ce.shardManager.GetShardStats(),
		"agents":   ce.agentScheduler.GetStats(),
		"pipelines": ce.pipelineOrch.GetStats(),
	}
	
	if tenantID != "" {
		stats["tenant"] = ce.shardManager.GetTenantStats(tenantID)
	}
	
	return stats
}

// Close shuts down the cognitive engine gracefully
func (ce *CognitiveEngine) Close() error {
	close(ce.done)
	
	// Close all components
	ce.shardManager.Close()
	
	ce.mu.RLock()
	for _, engine := range ce.inferenceEngines {
		engine.Close()
	}
	ce.mu.RUnlock()
	
	ce.agentScheduler.Close()
	ce.pipelineOrch.Close()
	
	return nil
}

// ============================================================================
// Helper Functions for Common Operations
// ============================================================================

// CreateConceptNode creates a new concept node
func (ce *CognitiveEngine) CreateConceptNode(name, tenantID string) (atomspace.Atom, error) {
	atomID := atomspace.GenerateAtomID(atomspace.ConceptNodeType, name, nil)
	node := atomspace.NewNode(atomID, name, tenantID, atomspace.ConceptNodeType)
	
	if err := ce.AddAtom(node); err != nil {
		return nil, err
	}
	
	return node, nil
}

// CreateInheritanceLink creates an inheritance link between two atoms
func (ce *CognitiveEngine) CreateInheritanceLink(sourceID, targetID, tenantID string) (atomspace.Atom, error) {
	source, err := ce.GetAtom(sourceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("source atom not found: %w", err)
	}
	
	target, err := ce.GetAtom(targetID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("target atom not found: %w", err)
	}
	
	outgoing := []atomspace.Atom{source, target}
	atomID := atomspace.GenerateAtomID(atomspace.InheritanceLinkType, "inheritance", outgoing)
	link := atomspace.NewLink(atomID, "inheritance", tenantID, atomspace.InheritanceLinkType, outgoing)
	
	if err := ce.AddAtom(link); err != nil {
		return nil, err
	}
	
	return link, nil
}

// CreateDefaultPipeline creates a default cognitive processing pipeline
func (ce *CognitiveEngine) CreateDefaultPipeline(tenantID string) (string, error) {
	pipelineID := fmt.Sprintf("default-pipeline-%s-%d", tenantID, time.Now().Unix())
	p, err := ce.CreatePipeline(pipelineID, "Default Cognitive Pipeline", tenantID)
	if err != nil {
		return "", err
	}
	
	// Get tenant's inference engine
	ce.mu.RLock()
	inferenceEngine := ce.inferenceEngines[tenantID]
	ce.mu.RUnlock()
	
	if inferenceEngine == nil {
		return "", fmt.Errorf("tenant %s not initialized", tenantID)
	}
	
	// Add stages
	// Note: We need to get a shard's atomspace for the stages
	// In a real implementation, we'd create a tenant-specific view
	shard, _ := ce.shardManager.GetShardByID(0)
	
	p.AddStage(pipeline.NewInferenceStage(inferenceEngine, tenantID, 5))
	p.AddStage(pipeline.NewAttentionAllocationStage(shard.AtomSpace, tenantID))
	p.AddStage(pipeline.NewAgentExecutionStage(ce.agentScheduler, tenantID))
	
	return pipelineID, nil
}

// Health check
func (ce *CognitiveEngine) Health() map[string]interface{} {
	ce.mu.RLock()
	numTenants := len(ce.inferenceEngines)
	ce.mu.RUnlock()
	
	return map[string]interface{}{
		"status":      "healthy",
		"num_tenants": numTenants,
		"num_shards":  ce.numShards,
		"timestamp":   time.Now().UTC(),
	}
}
