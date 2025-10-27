package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive"
	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
	"github.com/go-chi/chi/v5"
)

// CognitiveHandler handles HTTP requests for the cognitive engine
type CognitiveHandler struct {
	engine *cognitive.CognitiveEngine
}

// NewCognitiveHandler creates a new cognitive API handler
func NewCognitiveHandler(engine *cognitive.CognitiveEngine) *CognitiveHandler {
	return &CognitiveHandler{engine: engine}
}

// RegisterRoutes registers all cognitive API routes
func (h *CognitiveHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/cognitive", func(r chi.Router) {
		// Tenant management
		r.Post("/tenants/{tenantID}/init", h.InitializeTenant)
		
		// AtomSpace operations
		r.Post("/tenants/{tenantID}/atoms", h.CreateAtom)
		r.Get("/tenants/{tenantID}/atoms/{atomID}", h.GetAtom)
		r.Get("/tenants/{tenantID}/atoms", h.QueryAtoms)
		r.Put("/tenants/{tenantID}/atoms/{atomID}", h.UpdateAtom)
		r.Delete("/tenants/{tenantID}/atoms/{atomID}", h.DeleteAtom)
		
		// Concept nodes
		r.Post("/tenants/{tenantID}/concepts", h.CreateConcept)
		
		// Links
		r.Post("/tenants/{tenantID}/links/inheritance", h.CreateInheritanceLink)
		
		// Inference
		r.Post("/tenants/{tenantID}/inference", h.RunInference)
		
		// Pipelines
		r.Post("/tenants/{tenantID}/pipelines", h.CreatePipeline)
		r.Get("/tenants/{tenantID}/pipelines/{pipelineID}", h.GetPipeline)
		r.Post("/tenants/{tenantID}/pipelines/{pipelineID}/execute", h.ExecutePipeline)
		
		// Agents
		r.Get("/tenants/{tenantID}/agents", h.GetAgents)
		r.Get("/tenants/{tenantID}/agents/{agentID}", h.GetAgent)
		
		// Statistics
		r.Get("/tenants/{tenantID}/stats", h.GetStats)
		r.Get("/stats", h.GetGlobalStats)
		
		// Health
		r.Get("/health", h.Health)
	})
}

// InitializeTenant initializes a new tenant
func (h *CognitiveHandler) InitializeTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	if err := h.engine.InitializeTenant(tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Tenant initialized successfully",
		"tenant_id": tenantID,
	})
}

// CreateAtom creates a new atom
func (h *CognitiveHandler) CreateAtom(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	var req struct {
		Type   int     `json:"type"`
		Name   string  `json:"name"`
		Strength float64 `json:"strength"`
		Confidence float64 `json:"confidence"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	atomID := atomspace.GenerateAtomID(atomspace.AtomType(req.Type), req.Name, nil)
	node := atomspace.NewNode(atomID, req.Name, tenantID, atomspace.AtomType(req.Type))
	
	if req.Strength > 0 || req.Confidence > 0 {
		node.SetTruthValue(atomspace.TruthValue{
			Strength:   req.Strength,
			Confidence: req.Confidence,
		})
	}
	
	if err := h.engine.AddAtom(node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"atom_id": atomID,
		"name":    req.Name,
		"type":    req.Type,
	})
}

// GetAtom retrieves an atom
func (h *CognitiveHandler) GetAtom(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	atomID := chi.URLParam(r, "atomID")
	
	atom, err := h.engine.GetAtom(atomID, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	tv := atom.GetTruthValue()
	av := atom.GetAttentionValue()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"atom_id": atom.GetID(),
		"name":    atom.GetName(),
		"type":    atom.GetType(),
		"truth_value": map[string]float64{
			"strength":   tv.Strength,
			"confidence": tv.Confidence,
		},
		"attention_value": map[string]int16{
			"sti":  av.STI,
			"lti":  av.LTI,
			"vlti": av.VLTI,
		},
	})
}

// QueryAtoms queries atoms
func (h *CognitiveHandler) QueryAtoms(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	// Optional query parameters
	atomTypeStr := r.URL.Query().Get("type")
	name := r.URL.Query().Get("name")
	
	var atoms []atomspace.Atom
	
	if atomTypeStr != "" {
		// Parse atom type
		var atomType atomspace.AtomType
		switch atomTypeStr {
		case "node":
			atomType = atomspace.NodeType
		case "concept":
			atomType = atomspace.ConceptNodeType
		case "inheritance":
			atomType = atomspace.InheritanceLinkType
		default:
			atomType = atomspace.NodeType
		}
		
		atoms = h.engine.QueryAtoms(tenantID, func(a atomspace.Atom) bool {
			return a.GetType() == atomType
		})
	} else if name != "" {
		atoms = h.engine.QueryAtoms(tenantID, func(a atomspace.Atom) bool {
			return a.GetName() == name
		})
	} else {
		atoms = h.engine.QueryAtoms(tenantID, nil)
	}
	
	// Convert to JSON-friendly format
	result := make([]map[string]interface{}, len(atoms))
	for i, atom := range atoms {
		tv := atom.GetTruthValue()
		result[i] = map[string]interface{}{
			"atom_id": atom.GetID(),
			"name":    atom.GetName(),
			"type":    atom.GetType(),
			"truth_value": map[string]float64{
				"strength":   tv.Strength,
				"confidence": tv.Confidence,
			},
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"atoms": result,
		"count": len(result),
	})
}

// UpdateAtom updates an atom
func (h *CognitiveHandler) UpdateAtom(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	atomID := chi.URLParam(r, "atomID")
	
	var req struct {
		Strength   *float64 `json:"strength"`
		Confidence *float64 `json:"confidence"`
		STI        *int16   `json:"sti"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	err := h.engine.UpdateAtom(atomID, tenantID, func(atom atomspace.Atom) error {
		if req.Strength != nil || req.Confidence != nil {
			tv := atom.GetTruthValue()
			if req.Strength != nil {
				tv.Strength = *req.Strength
			}
			if req.Confidence != nil {
				tv.Confidence = *req.Confidence
			}
			atom.SetTruthValue(tv)
		}
		
		if req.STI != nil {
			av := atom.GetAttentionValue()
			av.STI = *req.STI
			atom.SetAttentionValue(av)
		}
		
		return nil
	})
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Atom updated successfully",
		"atom_id": atomID,
	})
}

// DeleteAtom deletes an atom
func (h *CognitiveHandler) DeleteAtom(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	atomID := chi.URLParam(r, "atomID")
	
	if err := h.engine.DeleteAtom(atomID, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Atom deleted successfully",
		"atom_id": atomID,
	})
}

// CreateConcept creates a concept node
func (h *CognitiveHandler) CreateConcept(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	var req struct {
		Name string `json:"name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	atom, err := h.engine.CreateConceptNode(req.Name, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"atom_id": atom.GetID(),
		"name":    atom.GetName(),
		"type":    "concept",
	})
}

// CreateInheritanceLink creates an inheritance link
func (h *CognitiveHandler) CreateInheritanceLink(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	var req struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	link, err := h.engine.CreateInheritanceLink(req.SourceID, req.TargetID, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"link_id":   link.GetID(),
		"source_id": req.SourceID,
		"target_id": req.TargetID,
		"type":      "inheritance",
	})
}

// RunInference runs inference
func (h *CognitiveHandler) RunInference(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	var req struct {
		MaxIterations int `json:"max_iterations"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.MaxIterations = 10
	}
	
	ctx := r.Context()
	newAtoms, err := h.engine.RunInference(ctx, tenantID, req.MaxIterations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"new_atoms_count": len(newAtoms),
		"max_iterations":  req.MaxIterations,
	})
}

// CreatePipeline creates a new pipeline
func (h *CognitiveHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	var req struct {
		Name string `json:"name"`
		UseDefault bool `json:"use_default"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	var pipelineID string
	var err error
	
	if req.UseDefault {
		pipelineID, err = h.engine.CreateDefaultPipeline(tenantID)
	} else {
		pipelineID = req.Name + "-" + time.Now().Format("20060102150405")
		_, err = h.engine.CreatePipeline(pipelineID, req.Name, tenantID)
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipeline_id": pipelineID,
		"name":        req.Name,
	})
}

// GetPipeline gets a specific pipeline
func (h *CognitiveHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineID := chi.URLParam(r, "pipelineID")
	
	pipeline, err := h.engine.GetPipeline(pipelineID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipeline.GetStats())
}

// ExecutePipeline executes a pipeline
func (h *CognitiveHandler) ExecutePipeline(w http.ResponseWriter, r *http.Request) {
	pipelineID := chi.URLParam(r, "pipelineID")
	
	ctx := r.Context()
	_, err := h.engine.ExecutePipeline(ctx, pipelineID, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Pipeline executed successfully",
		"pipeline_id": pipelineID,
	})
}

// GetAgents gets all agents for a tenant
func (h *CognitiveHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	agents := h.engine.GetAgentsByTenant(tenantID)
	
	agentStats := make([]map[string]interface{}, len(agents))
	for i, agent := range agents {
		agentStats[i] = agent.GetStats()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agentStats,
		"count":  len(agents),
	})
}

// GetAgent gets a specific agent
func (h *CognitiveHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	agent, exists := h.engine.GetAgent(agentID)
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent.GetStats())
}

// GetStats gets statistics for a tenant
func (h *CognitiveHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	
	stats := h.engine.GetStats(tenantID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetGlobalStats gets global statistics
func (h *CognitiveHandler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	stats := h.engine.GetStats("")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Health returns health status
func (h *CognitiveHandler) Health(w http.ResponseWriter, r *http.Request) {
	health := h.engine.Health()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
