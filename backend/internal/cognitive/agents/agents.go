package agents

import (
	"context"
	"sync"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
	"github.com/Avik2024/erebus/backend/internal/cognitive/inference"
)

// Agent represents an autonomous cognitive agent
type Agent interface {
	GetID() string
	GetName() string
	GetTenantID() string
	GetPriority() int
	Run(ctx context.Context) error
	GetStats() map[string]interface{}
}

// AgentState represents the state of an agent
type AgentState int

const (
	AgentStateIdle AgentState = iota
	AgentStateRunning
	AgentStateStopped
	AgentStateError
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	ID        string
	Name      string
	TenantID  string
	Priority  int
	State     AgentState
	RunCount  int64
	LastRun   time.Time
	TotalTime time.Duration
	mu        sync.RWMutex
}

func (a *BaseAgent) GetID() string {
	return a.ID
}

func (a *BaseAgent) GetName() string {
	return a.Name
}

func (a *BaseAgent) GetTenantID() string {
	return a.TenantID
}

func (a *BaseAgent) GetPriority() int {
	return a.Priority
}

func (a *BaseAgent) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	return map[string]interface{}{
		"id":           a.ID,
		"name":         a.Name,
		"tenant_id":    a.TenantID,
		"priority":     a.Priority,
		"state":        a.State,
		"run_count":    a.RunCount,
		"last_run":     a.LastRun,
		"total_time_ms": a.TotalTime.Milliseconds(),
		"avg_time_ms":  func() int64 {
			if a.RunCount == 0 {
				return 0
			}
			return a.TotalTime.Milliseconds() / a.RunCount
		}(),
	}
}

// MindAgent is a cognitive agent that performs inference cycles
type MindAgent struct {
	BaseAgent
	atomSpace atomspace.AtomSpaceInterface
	inference *inference.InferenceEngine
	cycleTime time.Duration
}

// NewMindAgent creates a new cognitive mind agent
func NewMindAgent(id, name, tenantID string, atomSpace atomspace.AtomSpaceInterface, inferenceEngine *inference.InferenceEngine) *MindAgent {
	return &MindAgent{
		BaseAgent: BaseAgent{
			ID:       id,
			Name:     name,
			TenantID: tenantID,
			Priority: 10,
			State:    AgentStateIdle,
		},
		atomSpace: atomSpace,
		inference: inferenceEngine,
		cycleTime: 100 * time.Millisecond,
	}
}

// Run executes the agent's cognitive cycle
func (ma *MindAgent) Run(ctx context.Context) error {
	ma.mu.Lock()
	ma.State = AgentStateRunning
	ma.mu.Unlock()
	
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		ma.mu.Lock()
		ma.RunCount++
		ma.LastRun = time.Now()
		ma.TotalTime += duration
		ma.State = AgentStateIdle
		ma.mu.Unlock()
	}()
	
	// Run inference cycle
	_, err := ma.inference.RunInference(ctx, ma.TenantID, 5)
	if err != nil {
		ma.mu.Lock()
		ma.State = AgentStateError
		ma.mu.Unlock()
		return err
	}
	
	return nil
}

// AttentionAgent manages attention allocation across atoms
type AttentionAgent struct {
	BaseAgent
	atomSpace atomspace.AtomSpaceInterface
	focusSize int
}

// NewAttentionAgent creates a new attention allocation agent
func NewAttentionAgent(id, name, tenantID string, atomSpace atomspace.AtomSpaceInterface) *AttentionAgent {
	return &AttentionAgent{
		BaseAgent: BaseAgent{
			ID:       id,
			Name:     name,
			TenantID: tenantID,
			Priority: 8,
			State:    AgentStateIdle,
		},
		atomSpace: atomSpace,
		focusSize: 100,
	}
}

// Run executes the attention allocation cycle
func (aa *AttentionAgent) Run(ctx context.Context) error {
	aa.mu.Lock()
	aa.State = AgentStateRunning
	aa.mu.Unlock()
	
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		aa.mu.Lock()
		aa.RunCount++
		aa.LastRun = time.Now()
		aa.TotalTime += duration
		aa.State = AgentStateIdle
		aa.mu.Unlock()
	}()
	
	// Get all atoms for this tenant
	atoms := aa.atomSpace.QueryAtoms(aa.TenantID, nil)
	
	// Update attention values based on usage and importance
	for _, atom := range atoms {
		av := atom.GetAttentionValue()
		
		// Decay STI over time
		av.STI = int16(float64(av.STI) * 0.95)
		
		// Boost important atoms (high truth value)
		tv := atom.GetTruthValue()
		if tv.Strength > 0.8 && tv.Confidence > 0.8 {
			av.STI += 10
			av.LTI += 1
		}
		
		atom.SetAttentionValue(av)
	}
	
	return nil
}

// AgentScheduler manages and schedules autonomous agents
type AgentScheduler struct {
	agents    map[string]Agent
	priority  []Agent // Sorted by priority
	mu        sync.RWMutex
	
	// Channels for agent communication
	registerChan   chan Agent
	unregisterChan chan string
	runChan        chan agentRunRequest
	done           chan struct{}
	
	workers int
}

type agentRunRequest struct {
	agent    Agent
	ctx      context.Context
	response chan error
}

// NewAgentScheduler creates a new agent scheduler
func NewAgentScheduler(workers int) *AgentScheduler {
	as := &AgentScheduler{
		agents:         make(map[string]Agent),
		priority:       make([]Agent, 0),
		registerChan:   make(chan Agent, 100),
		unregisterChan: make(chan string, 100),
		runChan:        make(chan agentRunRequest, 1000),
		done:           make(chan struct{}),
		workers:        workers,
	}
	
	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go as.worker()
	}
	
	// Start management goroutine
	go as.manage()
	
	return as
}

// worker processes agent run requests
func (as *AgentScheduler) worker() {
	for {
		select {
		case req := <-as.runChan:
			err := req.agent.Run(req.ctx)
			req.response <- err
		case <-as.done:
			return
		}
	}
}

// manage handles agent registration and scheduling
func (as *AgentScheduler) manage() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case agent := <-as.registerChan:
			as.registerInternal(agent)
		case agentID := <-as.unregisterChan:
			as.unregisterInternal(agentID)
		case <-ticker.C:
			as.scheduleAgents()
		case <-as.done:
			return
		}
	}
}

// RegisterAgent registers a new agent
func (as *AgentScheduler) RegisterAgent(agent Agent) {
	as.registerChan <- agent
}

// registerInternal is the internal implementation
func (as *AgentScheduler) registerInternal(agent Agent) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	as.agents[agent.GetID()] = agent
	as.rebuildPriorityQueue()
}

// UnregisterAgent removes an agent
func (as *AgentScheduler) UnregisterAgent(agentID string) {
	as.unregisterChan <- agentID
}

// unregisterInternal is the internal implementation
func (as *AgentScheduler) unregisterInternal(agentID string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	delete(as.agents, agentID)
	as.rebuildPriorityQueue()
}

// rebuildPriorityQueue rebuilds the priority queue
func (as *AgentScheduler) rebuildPriorityQueue() {
	as.priority = make([]Agent, 0, len(as.agents))
	for _, agent := range as.agents {
		as.priority = append(as.priority, agent)
	}
	
	// Sort by priority (higher priority first)
	for i := 0; i < len(as.priority); i++ {
		for j := i + 1; j < len(as.priority); j++ {
			if as.priority[i].GetPriority() < as.priority[j].GetPriority() {
				as.priority[i], as.priority[j] = as.priority[j], as.priority[i]
			}
		}
	}
}

// scheduleAgents runs agents in priority order
func (as *AgentScheduler) scheduleAgents() {
	as.mu.RLock()
	agentsToRun := make([]Agent, len(as.priority))
	copy(agentsToRun, as.priority)
	as.mu.RUnlock()
	
	// Run agents in priority order
	for _, agent := range agentsToRun {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		response := make(chan error, 1)
		as.runChan <- agentRunRequest{
			agent:    agent,
			ctx:      ctx,
			response: response,
		}
		
		// Wait for completion or timeout
		select {
		case <-response:
			// Agent completed
		case <-ctx.Done():
			// Timeout
		}
		
		cancel()
	}
}

// GetAgent retrieves an agent by ID
func (as *AgentScheduler) GetAgent(agentID string) (Agent, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	agent, exists := as.agents[agentID]
	return agent, exists
}

// GetAgentsByTenant returns all agents for a specific tenant
func (as *AgentScheduler) GetAgentsByTenant(tenantID string) []Agent {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	var agents []Agent
	for _, agent := range as.agents {
		if agent.GetTenantID() == tenantID {
			agents = append(agents, agent)
		}
	}
	
	return agents
}

// GetAllAgents returns all registered agents
func (as *AgentScheduler) GetAllAgents() []Agent {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	agents := make([]Agent, 0, len(as.agents))
	for _, agent := range as.agents {
		agents = append(agents, agent)
	}
	
	return agents
}

// GetStats returns scheduler statistics
func (as *AgentScheduler) GetStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	agentStats := make([]map[string]interface{}, 0, len(as.agents))
	for _, agent := range as.agents {
		agentStats = append(agentStats, agent.GetStats())
	}
	
	return map[string]interface{}{
		"total_agents": len(as.agents),
		"workers":      as.workers,
		"agents":       agentStats,
	}
}

// Close shuts down the scheduler
func (as *AgentScheduler) Close() {
	close(as.done)
}

// SpawnAgent autonomously creates and registers a new agent
func (as *AgentScheduler) SpawnAgent(agent Agent) error {
	as.RegisterAgent(agent)
	return nil
}

// TerminateAgent autonomously terminates an agent
func (as *AgentScheduler) TerminateAgent(agentID string) error {
	as.UnregisterAgent(agentID)
	return nil
}
