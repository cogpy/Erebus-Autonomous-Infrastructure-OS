package inference

import (
	"context"
	"sync"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
)

// InferenceRule represents a rule that can be applied to atoms
type InferenceRule interface {
	GetName() string
	GetPriority() int
	CanApply(atoms []atomspace.Atom) bool
	Apply(ctx context.Context, atoms []atomspace.Atom) ([]atomspace.Atom, error)
}

// InferenceEngine performs parallel reasoning over the AtomSpace
type InferenceEngine struct {
	atomSpace atomspace.AtomSpaceInterface
	rules     []InferenceRule
	workers   int
	mu        sync.RWMutex
	
	// Channels for concurrent inference
	taskChan   chan inferenceTask
	resultChan chan inferenceResult
	done       chan struct{}
}

type inferenceTask struct {
	tenantID string
	atoms    []atomspace.Atom
	rule     InferenceRule
	ctx      context.Context
}

type inferenceResult struct {
	newAtoms []atomspace.Atom
	err      error
	rule     string
}

// NewInferenceEngine creates a new parallel inference engine
func NewInferenceEngine(atomSpace atomspace.AtomSpaceInterface, workers int) *InferenceEngine {
	ie := &InferenceEngine{
		atomSpace:  atomSpace,
		rules:      make([]InferenceRule, 0),
		workers:    workers,
		taskChan:   make(chan inferenceTask, 1000),
		resultChan: make(chan inferenceResult, 1000),
		done:       make(chan struct{}),
	}
	
	// Start worker pool for parallel inference
	for i := 0; i < workers; i++ {
		go ie.worker()
	}
	
	return ie
}

// worker processes inference tasks concurrently
func (ie *InferenceEngine) worker() {
	for {
		select {
		case task := <-ie.taskChan:
			newAtoms, err := task.rule.Apply(task.ctx, task.atoms)
			ie.resultChan <- inferenceResult{
				newAtoms: newAtoms,
				err:      err,
				rule:     task.rule.GetName(),
			}
		case <-ie.done:
			return
		}
	}
}

// AddRule adds an inference rule to the engine
func (ie *InferenceEngine) AddRule(rule InferenceRule) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.rules = append(ie.rules, rule)
}

// RunInference executes inference rules on atoms for a tenant
func (ie *InferenceEngine) RunInference(ctx context.Context, tenantID string, maxIterations int) ([]atomspace.Atom, error) {
	var allNewAtoms []atomspace.Atom
	
	for iteration := 0; iteration < maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			return allNewAtoms, ctx.Err()
		default:
		}
		
		// Get all atoms for this tenant
		atoms := ie.atomSpace.QueryAtoms(tenantID, nil)
		
		if len(atoms) == 0 {
			break
		}
		
		// Try to apply each rule in parallel
		ie.mu.RLock()
		tasksSubmitted := 0
		for _, rule := range ie.rules {
			if rule.CanApply(atoms) {
				ie.taskChan <- inferenceTask{
					tenantID: tenantID,
					atoms:    atoms,
					rule:     rule,
					ctx:      ctx,
				}
				tasksSubmitted++
			}
		}
		ie.mu.RUnlock()
		
		// Collect results from parallel inference
		if tasksSubmitted == 0 {
			break
		}
		
		newAtomsThisIteration := 0
		for i := 0; i < tasksSubmitted; i++ {
			result := <-ie.resultChan
			if result.err != nil {
				continue
			}
			
			// Add new atoms to the atomspace
			for _, atom := range result.newAtoms {
				if err := ie.atomSpace.AddAtom(atom); err == nil {
					allNewAtoms = append(allNewAtoms, atom)
					newAtomsThisIteration++
				}
			}
		}
		
		// If no new atoms were created, we've reached fixpoint
		if newAtomsThisIteration == 0 {
			break
		}
	}
	
	return allNewAtoms, nil
}

// Close shuts down the inference engine
func (ie *InferenceEngine) Close() {
	close(ie.done)
}

// ============================================================================
// Basic Inference Rules
// ============================================================================

// DeductionRule implements modus ponens: A->B, A |- B
type DeductionRule struct {
	priority int
}

func NewDeductionRule() *DeductionRule {
	return &DeductionRule{priority: 10}
}

func (r *DeductionRule) GetName() string {
	return "deduction"
}

func (r *DeductionRule) GetPriority() int {
	return r.priority
}

func (r *DeductionRule) CanApply(atoms []atomspace.Atom) bool {
	// Check if we have at least one inheritance link and related nodes
	hasInheritance := false
	for _, atom := range atoms {
		if atom.GetType() == atomspace.InheritanceLinkType {
			hasInheritance = true
			break
		}
	}
	return hasInheritance && len(atoms) >= 2
}

func (r *DeductionRule) Apply(ctx context.Context, atoms []atomspace.Atom) ([]atomspace.Atom, error) {
	var newAtoms []atomspace.Atom
	
	// Find inheritance links: A->B and B->C, infer A->C
	inheritanceLinks := make([]*atomspace.Link, 0)
	for _, atom := range atoms {
		if atom.GetType() == atomspace.InheritanceLinkType {
			if link, ok := atom.(*atomspace.Link); ok {
				inheritanceLinks = append(inheritanceLinks, link)
			}
		}
	}
	
	// Try to chain inheritance links
	for i, link1 := range inheritanceLinks {
		if len(link1.Outgoing) != 2 {
			continue
		}
		
		for j, link2 := range inheritanceLinks {
			if i == j || len(link2.Outgoing) != 2 {
				continue
			}
			
			// Check if link1.B == link2.A
			if link1.Outgoing[1].GetID() == link2.Outgoing[0].GetID() {
				// Create new inheritance link A->C
				tenantID := link1.GetTenantID()
				newOutgoing := []atomspace.Atom{link1.Outgoing[0], link2.Outgoing[1]}
				newID := atomspace.GenerateAtomID(atomspace.InheritanceLinkType, "inheritance", newOutgoing)
				
				newLink := atomspace.NewLink(newID, "inheritance", tenantID, atomspace.InheritanceLinkType, newOutgoing)
				
				// Calculate new truth value (simplified PLN formula)
				tv1 := link1.GetTruthValue()
				tv2 := link2.GetTruthValue()
				newTV := atomspace.TruthValue{
					Strength:   tv1.Strength * tv2.Strength,
					Confidence: tv1.Confidence * tv2.Confidence * 0.9, // Reduce confidence slightly
				}
				newLink.SetTruthValue(newTV)
				
				newAtoms = append(newAtoms, newLink)
			}
		}
	}
	
	return newAtoms, nil
}

// InductionRule implements generalization from instances
type InductionRule struct {
	priority int
}

func NewInductionRule() *InductionRule {
	return &InductionRule{priority: 5}
}

func (r *InductionRule) GetName() string {
	return "induction"
}

func (r *InductionRule) GetPriority() int {
	return r.priority
}

func (r *InductionRule) CanApply(atoms []atomspace.Atom) bool {
	// Need multiple similar inheritance links to generalize
	count := 0
	for _, atom := range atoms {
		if atom.GetType() == atomspace.InheritanceLinkType {
			count++
		}
	}
	return count >= 3
}

func (r *InductionRule) Apply(ctx context.Context, atoms []atomspace.Atom) ([]atomspace.Atom, error) {
	var newAtoms []atomspace.Atom
	
	// Find common patterns in inheritance links
	inheritanceLinks := make([]*atomspace.Link, 0)
	for _, atom := range atoms {
		if atom.GetType() == atomspace.InheritanceLinkType {
			if link, ok := atom.(*atomspace.Link); ok {
				inheritanceLinks = append(inheritanceLinks, link)
			}
		}
	}
	
	// Group by target (B in A->B)
	targetGroups := make(map[string][]*atomspace.Link)
	for _, link := range inheritanceLinks {
		if len(link.Outgoing) == 2 {
			targetID := link.Outgoing[1].GetID()
			targetGroups[targetID] = append(targetGroups[targetID], link)
		}
	}
	
	// If multiple instances inherit from same concept, create similarity links
	for _, group := range targetGroups {
		if len(group) >= 2 {
			for i := 0; i < len(group)-1; i++ {
				for j := i + 1; j < len(group); j++ {
					source1 := group[i].Outgoing[0]
					source2 := group[j].Outgoing[0]
					
					tenantID := group[i].GetTenantID()
					newOutgoing := []atomspace.Atom{source1, source2}
					newID := atomspace.GenerateAtomID(atomspace.SimilarityLinkType, "similarity", newOutgoing)
					
					newLink := atomspace.NewLink(newID, "similarity", tenantID, atomspace.SimilarityLinkType, newOutgoing)
					
					// Similarity strength based on common inheritance
					newTV := atomspace.TruthValue{
						Strength:   0.7, // Moderate similarity
						Confidence: 0.8,
					}
					newLink.SetTruthValue(newTV)
					
					newAtoms = append(newAtoms, newLink)
				}
			}
		}
	}
	
	return newAtoms, nil
}

// AbductionRule implements hypothesis generation: B, A->B |- A
type AbductionRule struct {
	priority int
}

func NewAbductionRule() *AbductionRule {
	return &AbductionRule{priority: 3}
}

func (r *AbductionRule) GetName() string {
	return "abduction"
}

func (r *AbductionRule) GetPriority() int {
	return r.priority
}

func (r *AbductionRule) CanApply(atoms []atomspace.Atom) bool {
	return len(atoms) >= 2
}

func (r *AbductionRule) Apply(ctx context.Context, atoms []atomspace.Atom) ([]atomspace.Atom, error) {
	// Abduction is hypothesis generation - we'll create it with lower confidence
	// This is a simplified version
	return []atomspace.Atom{}, nil
}

// PatternMatcher finds atoms matching a pattern
type PatternMatcher struct {
	atomSpace atomspace.AtomSpaceInterface
}

func NewPatternMatcher(atomSpace atomspace.AtomSpaceInterface) *PatternMatcher {
	return &PatternMatcher{atomSpace: atomSpace}
}

// MatchPattern finds atoms matching the given pattern
func (pm *PatternMatcher) MatchPattern(tenantID string, pattern atomspace.Atom) []atomspace.Atom {
	// Simple pattern matching by type and name
	return pm.atomSpace.QueryAtoms(tenantID, func(a atomspace.Atom) bool {
		return a.GetType() == pattern.GetType() && 
		       (pattern.GetName() == "" || a.GetName() == pattern.GetName())
	})
}

// QueryStats returns inference statistics
type InferenceStats struct {
	TotalInferences  int64
	SuccessfulRules  map[string]int64
	FailedRules      map[string]int64
	AverageTime      time.Duration
	LastRun          time.Time
	mu               sync.RWMutex
}

func NewInferenceStats() *InferenceStats {
	return &InferenceStats{
		SuccessfulRules: make(map[string]int64),
		FailedRules:     make(map[string]int64),
	}
}

func (is *InferenceStats) RecordSuccess(ruleName string, duration time.Duration) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.TotalInferences++
	is.SuccessfulRules[ruleName]++
	is.LastRun = time.Now()
}

func (is *InferenceStats) RecordFailure(ruleName string) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.FailedRules[ruleName]++
}

func (is *InferenceStats) GetStats() map[string]interface{} {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	return map[string]interface{}{
		"total_inferences":  is.TotalInferences,
		"successful_rules":  is.SuccessfulRules,
		"failed_rules":      is.FailedRules,
		"average_time_ms":   is.AverageTime.Milliseconds(),
		"last_run":          is.LastRun,
	}
}
