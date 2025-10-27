package atomspace

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

// AtomSpace is a thread-safe, multi-tenant knowledge store with concurrent access
type AtomSpace struct {
	atoms    map[string]Atom          // atomID -> Atom
	byTenant map[string]map[string]Atom // tenantID -> atomID -> Atom
	byType   map[AtomType]map[string]Atom // atomType -> atomID -> Atom
	indices  map[string]map[string]bool  // name -> atomID -> exists (for fast lookups)
	mu       sync.RWMutex
	
	// Concurrency channels for multiplexed operations
	addChan    chan atomRequest
	queryChan  chan queryRequest
	updateChan chan updateRequest
	deleteChan chan deleteRequest
	done       chan struct{}
}

type atomRequest struct {
	atom     Atom
	response chan error
}

type queryRequest struct {
	tenantID string
	filter   func(Atom) bool
	response chan []Atom
}

type updateRequest struct {
	atomID   string
	tenantID string
	updater  func(Atom) error
	response chan error
}

type deleteRequest struct {
	atomID   string
	tenantID string
	response chan error
}

// NewAtomSpace creates a new multi-tenant AtomSpace with concurrent channels
func NewAtomSpace(workers int) *AtomSpace {
	as := &AtomSpace{
		atoms:      make(map[string]Atom),
		byTenant:   make(map[string]map[string]Atom),
		byType:     make(map[AtomType]map[string]Atom),
		indices:    make(map[string]map[string]bool),
		addChan:    make(chan atomRequest, 1000),
		queryChan:  make(chan queryRequest, 1000),
		updateChan: make(chan updateRequest, 1000),
		deleteChan: make(chan deleteRequest, 1000),
		done:       make(chan struct{}),
	}
	
	// Start worker goroutines for concurrent operation handling
	for i := 0; i < workers; i++ {
		go as.worker()
	}
	
	return as
}

// worker processes requests from multiple channels concurrently
func (as *AtomSpace) worker() {
	for {
		select {
		case req := <-as.addChan:
			req.response <- as.addAtomInternal(req.atom)
		case req := <-as.queryChan:
			req.response <- as.queryAtomsInternal(req.tenantID, req.filter)
		case req := <-as.updateChan:
			req.response <- as.updateAtomInternal(req.atomID, req.tenantID, req.updater)
		case req := <-as.deleteChan:
			req.response <- as.deleteAtomInternal(req.atomID, req.tenantID)
		case <-as.done:
			return
		}
	}
}

// AddAtom adds an atom to the atomspace (thread-safe, multiplexed)
func (as *AtomSpace) AddAtom(atom Atom) error {
	response := make(chan error, 1)
	as.addChan <- atomRequest{atom: atom, response: response}
	return <-response
}

// addAtomInternal is the internal implementation of AddAtom
func (as *AtomSpace) addAtomInternal(atom Atom) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	atomID := atom.GetID()
	tenantID := atom.GetTenantID()
	atomType := atom.GetType()
	
	// Check if atom already exists
	if _, exists := as.atoms[atomID]; exists {
		return fmt.Errorf("atom with ID %s already exists", atomID)
	}
	
	// Add to main store
	as.atoms[atomID] = atom
	
	// Add to tenant index
	if as.byTenant[tenantID] == nil {
		as.byTenant[tenantID] = make(map[string]Atom)
	}
	as.byTenant[tenantID][atomID] = atom
	
	// Add to type index
	if as.byType[atomType] == nil {
		as.byType[atomType] = make(map[string]Atom)
	}
	as.byType[atomType][atomID] = atom
	
	// Add to name index
	name := atom.GetName()
	if as.indices[name] == nil {
		as.indices[name] = make(map[string]bool)
	}
	as.indices[name][atomID] = true
	
	return nil
}

// GetAtom retrieves an atom by ID and tenant
func (as *AtomSpace) GetAtom(atomID, tenantID string) (Atom, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	atom, exists := as.atoms[atomID]
	if !exists {
		return nil, fmt.Errorf("atom with ID %s not found", atomID)
	}
	
	if atom.GetTenantID() != tenantID {
		return nil, fmt.Errorf("atom does not belong to tenant %s", tenantID)
	}
	
	return atom, nil
}

// QueryAtoms returns atoms matching a filter for a specific tenant (concurrent)
func (as *AtomSpace) QueryAtoms(tenantID string, filter func(Atom) bool) []Atom {
	response := make(chan []Atom, 1)
	as.queryChan <- queryRequest{tenantID: tenantID, filter: filter, response: response}
	return <-response
}

// queryAtomsInternal is the internal implementation
func (as *AtomSpace) queryAtomsInternal(tenantID string, filter func(Atom) bool) []Atom {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	var results []Atom
	tenantAtoms := as.byTenant[tenantID]
	
	for _, atom := range tenantAtoms {
		if filter == nil || filter(atom) {
			results = append(results, atom)
		}
	}
	
	return results
}

// GetAtomsByType returns all atoms of a specific type for a tenant
func (as *AtomSpace) GetAtomsByType(tenantID string, atomType AtomType) []Atom {
	return as.QueryAtoms(tenantID, func(a Atom) bool {
		return a.GetType() == atomType
	})
}

// GetAtomsByName returns all atoms with a specific name for a tenant
func (as *AtomSpace) GetAtomsByName(tenantID string, name string) []Atom {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	var results []Atom
	atomIDs := as.indices[name]
	
	for atomID := range atomIDs {
		atom := as.atoms[atomID]
		if atom.GetTenantID() == tenantID {
			results = append(results, atom)
		}
	}
	
	return results
}

// UpdateAtom updates an atom using an updater function (thread-safe)
func (as *AtomSpace) UpdateAtom(atomID, tenantID string, updater func(Atom) error) error {
	response := make(chan error, 1)
	as.updateChan <- updateRequest{atomID: atomID, tenantID: tenantID, updater: updater, response: response}
	return <-response
}

// updateAtomInternal is the internal implementation
func (as *AtomSpace) updateAtomInternal(atomID, tenantID string, updater func(Atom) error) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	atom, exists := as.atoms[atomID]
	if !exists {
		return fmt.Errorf("atom with ID %s not found", atomID)
	}
	
	if atom.GetTenantID() != tenantID {
		return fmt.Errorf("atom does not belong to tenant %s", tenantID)
	}
	
	return updater(atom)
}

// DeleteAtom removes an atom (thread-safe)
func (as *AtomSpace) DeleteAtom(atomID, tenantID string) error {
	response := make(chan error, 1)
	as.deleteChan <- deleteRequest{atomID: atomID, tenantID: tenantID, response: response}
	return <-response
}

// deleteAtomInternal is the internal implementation
func (as *AtomSpace) deleteAtomInternal(atomID, tenantID string) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	atom, exists := as.atoms[atomID]
	if !exists {
		return fmt.Errorf("atom with ID %s not found", atomID)
	}
	
	if atom.GetTenantID() != tenantID {
		return fmt.Errorf("atom does not belong to tenant %s", tenantID)
	}
	
	// Remove from main store
	delete(as.atoms, atomID)
	
	// Remove from tenant index
	delete(as.byTenant[tenantID], atomID)
	
	// Remove from type index
	delete(as.byType[atom.GetType()], atomID)
	
	// Remove from name index
	name := atom.GetName()
	delete(as.indices[name], atomID)
	if len(as.indices[name]) == 0 {
		delete(as.indices, name)
	}
	
	return nil
}

// GetStats returns statistics about the AtomSpace
func (as *AtomSpace) GetStats(tenantID string) map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	tenantAtoms := as.byTenant[tenantID]
	
	stats := map[string]interface{}{
		"total_atoms": len(tenantAtoms),
		"atoms_by_type": make(map[AtomType]int),
	}
	
	for _, atom := range tenantAtoms {
		typeCount := stats["atoms_by_type"].(map[AtomType]int)
		typeCount[atom.GetType()]++
	}
	
	return stats
}

// Close shuts down the AtomSpace workers
func (as *AtomSpace) Close() {
	close(as.done)
}

// GenerateAtomID generates a unique ID for an atom based on its content
func GenerateAtomID(atomType AtomType, name string, outgoing []Atom) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:%s", atomType, name)))
	for _, atom := range outgoing {
		h.Write([]byte(atom.GetID()))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
