package atomspace

import (
	"sync"
	"time"
)

// AtomType represents the type of an atom
type AtomType int

const (
	// Node types
	NodeType AtomType = iota
	ConceptNodeType
	PredicateNodeType
	VariableNodeType
	
	// Link types
	LinkType
	InheritanceLinkType
	SimilarityLinkType
	ExecutionLinkType
	EvaluationLinkType
)

// TruthValue represents probabilistic truth with strength and confidence
type TruthValue struct {
	Strength   float64 // [0, 1] - probability that the statement is true
	Confidence float64 // [0, 1] - confidence in the strength value
}

// AttentionValue represents the importance of an atom in the cognitive system
type AttentionValue struct {
	STI int16 // Short-term importance
	LTI int16 // Long-term importance
	VLTI int16 // Very long-term importance
}

// Atom is the fundamental unit of knowledge representation
type Atom interface {
	GetID() string
	GetType() AtomType
	GetName() string
	GetTruthValue() TruthValue
	SetTruthValue(tv TruthValue)
	GetAttentionValue() AttentionValue
	SetAttentionValue(av AttentionValue)
	GetTenantID() string
	Clone() Atom
}

// BaseAtom provides common functionality for all atoms
type BaseAtom struct {
	ID             string
	Type           AtomType
	Name           string
	TruthVal       TruthValue
	AttentionVal   AttentionValue
	TenantID       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	mu             sync.RWMutex
}

func (a *BaseAtom) GetID() string {
	return a.ID
}

func (a *BaseAtom) GetType() AtomType {
	return a.Type
}

func (a *BaseAtom) GetName() string {
	return a.Name
}

func (a *BaseAtom) GetTruthValue() TruthValue {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.TruthVal
}

func (a *BaseAtom) SetTruthValue(tv TruthValue) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.TruthVal = tv
	a.UpdatedAt = time.Now()
}

func (a *BaseAtom) GetAttentionValue() AttentionValue {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.AttentionVal
}

func (a *BaseAtom) SetAttentionValue(av AttentionValue) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.AttentionVal = av
	a.UpdatedAt = time.Now()
}

func (a *BaseAtom) GetTenantID() string {
	return a.TenantID
}

// Node represents a simple named atom
type Node struct {
	BaseAtom
}

func NewNode(id, name, tenantID string, atomType AtomType) *Node {
	now := time.Now()
	return &Node{
		BaseAtom: BaseAtom{
			ID:             id,
			Type:           atomType,
			Name:           name,
			TenantID:       tenantID,
			TruthVal:       TruthValue{Strength: 1.0, Confidence: 1.0},
			AttentionVal:   AttentionValue{STI: 0, LTI: 0, VLTI: 0},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
}

func (n *Node) Clone() Atom {
	return &Node{
		BaseAtom: BaseAtom{
			ID:           n.ID,
			Type:         n.Type,
			Name:         n.Name,
			TenantID:     n.TenantID,
			TruthVal:     n.TruthVal,
			AttentionVal: n.AttentionVal,
			CreatedAt:    n.CreatedAt,
			UpdatedAt:    n.UpdatedAt,
		},
	}
}

// Link represents a relationship between atoms
type Link struct {
	BaseAtom
	Outgoing []Atom // The atoms this link connects
}

func NewLink(id, name, tenantID string, atomType AtomType, outgoing []Atom) *Link {
	now := time.Now()
	return &Link{
		BaseAtom: BaseAtom{
			ID:             id,
			Type:           atomType,
			Name:           name,
			TenantID:       tenantID,
			TruthVal:       TruthValue{Strength: 1.0, Confidence: 1.0},
			AttentionVal:   AttentionValue{STI: 0, LTI: 0, VLTI: 0},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Outgoing: outgoing,
	}
}

func (l *Link) GetOutgoing() []Atom {
	return l.Outgoing
}

func (l *Link) Clone() Atom {
	outgoingCopy := make([]Atom, len(l.Outgoing))
	copy(outgoingCopy, l.Outgoing)
	return &Link{
		BaseAtom: BaseAtom{
			ID:           l.ID,
			Type:         l.Type,
			Name:         l.Name,
			TenantID:     l.TenantID,
			TruthVal:     l.TruthVal,
			AttentionVal: l.AttentionVal,
			CreatedAt:    l.CreatedAt,
			UpdatedAt:    l.UpdatedAt,
		},
		Outgoing: outgoingCopy,
	}
}
