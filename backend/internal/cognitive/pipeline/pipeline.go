package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive/agents"
	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
	"github.com/Avik2024/erebus/backend/internal/cognitive/inference"
)

// PipelineStage represents a stage in the cognitive pipeline
type PipelineStage interface {
	GetName() string
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// Pipeline represents a cognitive processing pipeline
type Pipeline struct {
	ID          string
	Name        string
	TenantID    string
	Stages      []PipelineStage
	State       PipelineState
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	mu          sync.RWMutex
}

// PipelineState represents the state of a pipeline
type PipelineState int

const (
	PipelineStateCreated PipelineState = iota
	PipelineStateRunning
	PipelineStateCompleted
	PipelineStateFailed
	PipelineStatePaused
)

// NewPipeline creates a new cognitive pipeline
func NewPipeline(id, name, tenantID string) *Pipeline {
	return &Pipeline{
		ID:        id,
		Name:      name,
		TenantID:  tenantID,
		Stages:    make([]PipelineStage, 0),
		State:     PipelineStateCreated,
		CreatedAt: time.Now(),
	}
}

// AddStage adds a stage to the pipeline
func (p *Pipeline) AddStage(stage PipelineStage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Stages = append(p.Stages, stage)
}

// Execute runs the pipeline
func (p *Pipeline) Execute(ctx context.Context, initialInput interface{}) (interface{}, error) {
	p.mu.Lock()
	p.State = PipelineStateRunning
	p.StartedAt = time.Now()
	p.mu.Unlock()
	
	currentInput := initialInput
	
	for i, stage := range p.Stages {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			p.State = PipelineStateFailed
			p.mu.Unlock()
			return nil, fmt.Errorf("pipeline execution cancelled at stage %d", i)
		default:
		}
		
		output, err := stage.Execute(ctx, currentInput)
		if err != nil {
			p.mu.Lock()
			p.State = PipelineStateFailed
			p.CompletedAt = time.Now()
			p.mu.Unlock()
			return nil, fmt.Errorf("stage %s failed: %w", stage.GetName(), err)
		}
		
		currentInput = output
	}
	
	p.mu.Lock()
	p.State = PipelineStateCompleted
	p.CompletedAt = time.Now()
	p.mu.Unlock()
	
	return currentInput, nil
}

// GetStats returns pipeline statistics
func (p *Pipeline) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var duration time.Duration
	if !p.CompletedAt.IsZero() {
		duration = p.CompletedAt.Sub(p.StartedAt)
	} else if !p.StartedAt.IsZero() {
		duration = time.Since(p.StartedAt)
	}
	
	return map[string]interface{}{
		"id":           p.ID,
		"name":         p.Name,
		"tenant_id":    p.TenantID,
		"state":        p.State,
		"stages":       len(p.Stages),
		"created_at":   p.CreatedAt,
		"started_at":   p.StartedAt,
		"completed_at": p.CompletedAt,
		"duration_ms":  duration.Milliseconds(),
	}
}

// ============================================================================
// Pipeline Stages
// ============================================================================

// AtomIngestionStage ingests atoms into the AtomSpace
type AtomIngestionStage struct {
	atomSpace atomspace.AtomSpaceInterface
	tenantID  string
}

func NewAtomIngestionStage(atomSpace atomspace.AtomSpaceInterface, tenantID string) *AtomIngestionStage {
	return &AtomIngestionStage{
		atomSpace: atomSpace,
		tenantID:  tenantID,
	}
}

func (s *AtomIngestionStage) GetName() string {
	return "atom-ingestion"
}

func (s *AtomIngestionStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	atoms, ok := input.([]atomspace.Atom)
	if !ok {
		return nil, fmt.Errorf("expected []Atom, got %T", input)
	}
	
	for _, atom := range atoms {
		if err := s.atomSpace.AddAtom(atom); err != nil {
			// Ignore duplicate atoms
			continue
		}
	}
	
	return atoms, nil
}

// InferenceStage runs inference on atoms
type InferenceStage struct {
	engine       *inference.InferenceEngine
	tenantID     string
	maxIterations int
}

func NewInferenceStage(engine *inference.InferenceEngine, tenantID string, maxIterations int) *InferenceStage {
	return &InferenceStage{
		engine:       engine,
		tenantID:     tenantID,
		maxIterations: maxIterations,
	}
}

func (s *InferenceStage) GetName() string {
	return "inference"
}

func (s *InferenceStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	newAtoms, err := s.engine.RunInference(ctx, s.tenantID, s.maxIterations)
	if err != nil {
		return nil, err
	}
	
	return newAtoms, nil
}

// AttentionAllocationStage allocates attention to atoms
type AttentionAllocationStage struct {
	atomSpace atomspace.AtomSpaceInterface
	tenantID  string
}

func NewAttentionAllocationStage(atomSpace atomspace.AtomSpaceInterface, tenantID string) *AttentionAllocationStage {
	return &AttentionAllocationStage{
		atomSpace: atomSpace,
		tenantID:  tenantID,
	}
}

func (s *AttentionAllocationStage) GetName() string {
	return "attention-allocation"
}

func (s *AttentionAllocationStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	atoms := s.atomSpace.QueryAtoms(s.tenantID, nil)
	
	// Update attention values
	for _, atom := range atoms {
		av := atom.GetAttentionValue()
		tv := atom.GetTruthValue()
		
		// Increase attention for high-confidence atoms
		if tv.Confidence > 0.8 {
			av.STI += 5
		}
		
		// Decay attention over time
		av.STI = int16(float64(av.STI) * 0.95)
		
		atom.SetAttentionValue(av)
	}
	
	return atoms, nil
}

// AgentExecutionStage runs cognitive agents
type AgentExecutionStage struct {
	scheduler *agents.AgentScheduler
	tenantID  string
}

func NewAgentExecutionStage(scheduler *agents.AgentScheduler, tenantID string) *AgentExecutionStage {
	return &AgentExecutionStage{
		scheduler: scheduler,
		tenantID:  tenantID,
	}
}

func (s *AgentExecutionStage) GetName() string {
	return "agent-execution"
}

func (s *AgentExecutionStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	tenantAgents := s.scheduler.GetAgentsByTenant(s.tenantID)
	
	for _, agent := range tenantAgents {
		if err := agent.Run(ctx); err != nil {
			// Continue with other agents even if one fails
			continue
		}
	}
	
	return input, nil
}

// ============================================================================
// Pipeline Orchestrator
// ============================================================================

// PipelineOrchestrator manages multiple pipelines
type PipelineOrchestrator struct {
	pipelines map[string]*Pipeline
	mu        sync.RWMutex
	
	// Channels for concurrent pipeline management
	createChan chan pipelineCreateRequest
	executeChan chan pipelineExecuteRequest
	deleteChan chan string
	done       chan struct{}
	
	workers int
}

type pipelineCreateRequest struct {
	pipeline *Pipeline
	response chan error
}

type pipelineExecuteRequest struct {
	pipelineID string
	ctx        context.Context
	input      interface{}
	response   chan pipelineExecuteResponse
}

type pipelineExecuteResponse struct {
	output interface{}
	err    error
}

// NewPipelineOrchestrator creates a new pipeline orchestrator
func NewPipelineOrchestrator(workers int) *PipelineOrchestrator {
	po := &PipelineOrchestrator{
		pipelines:   make(map[string]*Pipeline),
		createChan:  make(chan pipelineCreateRequest, 100),
		executeChan: make(chan pipelineExecuteRequest, 1000),
		deleteChan:  make(chan string, 100),
		done:        make(chan struct{}),
		workers:     workers,
	}
	
	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go po.worker()
	}
	
	// Start management goroutine
	go po.manage()
	
	return po
}

// worker processes pipeline execution requests
func (po *PipelineOrchestrator) worker() {
	for {
		select {
		case req := <-po.executeChan:
			po.mu.RLock()
			pipeline, exists := po.pipelines[req.pipelineID]
			po.mu.RUnlock()
			
			if !exists {
				req.response <- pipelineExecuteResponse{
					err: fmt.Errorf("pipeline %s not found", req.pipelineID),
				}
				continue
			}
			
			output, err := pipeline.Execute(req.ctx, req.input)
			req.response <- pipelineExecuteResponse{
				output: output,
				err:    err,
			}
		case <-po.done:
			return
		}
	}
}

// manage handles pipeline creation and deletion
func (po *PipelineOrchestrator) manage() {
	for {
		select {
		case req := <-po.createChan:
			req.response <- po.createPipelineInternal(req.pipeline)
		case pipelineID := <-po.deleteChan:
			po.deletePipelineInternal(pipelineID)
		case <-po.done:
			return
		}
	}
}

// CreatePipeline creates a new pipeline
func (po *PipelineOrchestrator) CreatePipeline(pipeline *Pipeline) error {
	response := make(chan error, 1)
	po.createChan <- pipelineCreateRequest{pipeline: pipeline, response: response}
	return <-response
}

// createPipelineInternal is the internal implementation
func (po *PipelineOrchestrator) createPipelineInternal(pipeline *Pipeline) error {
	po.mu.Lock()
	defer po.mu.Unlock()
	
	if _, exists := po.pipelines[pipeline.ID]; exists {
		return fmt.Errorf("pipeline %s already exists", pipeline.ID)
	}
	
	po.pipelines[pipeline.ID] = pipeline
	return nil
}

// ExecutePipeline executes a pipeline
func (po *PipelineOrchestrator) ExecutePipeline(ctx context.Context, pipelineID string, input interface{}) (interface{}, error) {
	response := make(chan pipelineExecuteResponse, 1)
	po.executeChan <- pipelineExecuteRequest{
		pipelineID: pipelineID,
		ctx:        ctx,
		input:      input,
		response:   response,
	}
	
	result := <-response
	return result.output, result.err
}

// GetPipeline retrieves a pipeline by ID
func (po *PipelineOrchestrator) GetPipeline(pipelineID string) (*Pipeline, error) {
	po.mu.RLock()
	defer po.mu.RUnlock()
	
	pipeline, exists := po.pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", pipelineID)
	}
	
	return pipeline, nil
}

// GetPipelinesByTenant returns all pipelines for a tenant
func (po *PipelineOrchestrator) GetPipelinesByTenant(tenantID string) []*Pipeline {
	po.mu.RLock()
	defer po.mu.RUnlock()
	
	var pipelines []*Pipeline
	for _, pipeline := range po.pipelines {
		if pipeline.TenantID == tenantID {
			pipelines = append(pipelines, pipeline)
		}
	}
	
	return pipelines
}

// DeletePipeline deletes a pipeline
func (po *PipelineOrchestrator) DeletePipeline(pipelineID string) {
	po.deleteChan <- pipelineID
}

// deletePipelineInternal is the internal implementation
func (po *PipelineOrchestrator) deletePipelineInternal(pipelineID string) {
	po.mu.Lock()
	defer po.mu.Unlock()
	delete(po.pipelines, pipelineID)
}

// GetStats returns orchestrator statistics
func (po *PipelineOrchestrator) GetStats() map[string]interface{} {
	po.mu.RLock()
	defer po.mu.RUnlock()
	
	pipelineStats := make([]map[string]interface{}, 0, len(po.pipelines))
	for _, pipeline := range po.pipelines {
		pipelineStats = append(pipelineStats, pipeline.GetStats())
	}
	
	return map[string]interface{}{
		"total_pipelines": len(po.pipelines),
		"workers":         po.workers,
		"pipelines":       pipelineStats,
	}
}

// Close shuts down the orchestrator
func (po *PipelineOrchestrator) Close() {
	close(po.done)
}
