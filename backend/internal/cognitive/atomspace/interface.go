package atomspace

// AtomSpaceInterface defines the interface for atomspace operations
type AtomSpaceInterface interface {
	AddAtom(atom Atom) error
	GetAtom(atomID, tenantID string) (Atom, error)
	QueryAtoms(tenantID string, filter func(Atom) bool) []Atom
	UpdateAtom(atomID, tenantID string, updater func(Atom) error) error
	DeleteAtom(atomID, tenantID string) error
	GetStats(tenantID string) map[string]interface{}
}

// Ensure AtomSpace implements the interface
var _ AtomSpaceInterface = (*AtomSpace)(nil)
