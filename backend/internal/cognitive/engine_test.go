package cognitive

import (
	"context"
	"testing"
	"time"
)

func TestNewCognitiveEngine(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	if engine == nil {
		t.Fatal("Expected non-nil cognitive engine")
	}
	
	health := engine.Health()
	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestInitializeTenant(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	// Try to initialize the same tenant again, should fail
	err = engine.InitializeTenant(tenantID)
	if err == nil {
		t.Error("Expected error when initializing same tenant twice")
	}
}

func TestCreateConceptNode(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	atom, err := engine.CreateConceptNode("TestConcept", tenantID)
	if err != nil {
		t.Fatalf("Failed to create concept node: %v", err)
	}
	
	if atom.GetName() != "TestConcept" {
		t.Errorf("Expected name 'TestConcept', got %s", atom.GetName())
	}
	
	if atom.GetTenantID() != tenantID {
		t.Errorf("Expected tenant ID %s, got %s", tenantID, atom.GetTenantID())
	}
}

func TestQueryAtoms(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	// Create some atoms
	_, err = engine.CreateConceptNode("Concept1", tenantID)
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}
	
	_, err = engine.CreateConceptNode("Concept2", tenantID)
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}
	
	// Query all atoms
	atoms := engine.QueryAtoms(tenantID, nil)
	if len(atoms) != 2 {
		t.Errorf("Expected 2 atoms, got %d", len(atoms))
	}
}

func TestCreateInheritanceLink(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	// Create concepts
	cat, err := engine.CreateConceptNode("Cat", tenantID)
	if err != nil {
		t.Fatalf("Failed to create Cat concept: %v", err)
	}
	
	animal, err := engine.CreateConceptNode("Animal", tenantID)
	if err != nil {
		t.Fatalf("Failed to create Animal concept: %v", err)
	}
	
	// Create inheritance link
	link, err := engine.CreateInheritanceLink(cat.GetID(), animal.GetID(), tenantID)
	if err != nil {
		t.Fatalf("Failed to create inheritance link: %v", err)
	}
	
	if link == nil {
		t.Error("Expected non-nil link")
	}
}

func TestRunInference(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	// Create a simple knowledge base
	cat, _ := engine.CreateConceptNode("Cat", tenantID)
	mammal, _ := engine.CreateConceptNode("Mammal", tenantID)
	animal, _ := engine.CreateConceptNode("Animal", tenantID)
	
	// Cat -> Mammal
	engine.CreateInheritanceLink(cat.GetID(), mammal.GetID(), tenantID)
	// Mammal -> Animal
	engine.CreateInheritanceLink(mammal.GetID(), animal.GetID(), tenantID)
	
	// Run inference (should infer Cat -> Animal)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	newAtoms, err := engine.RunInference(ctx, tenantID, 5)
	if err != nil {
		t.Fatalf("Failed to run inference: %v", err)
	}
	
	// Should have created some new atoms through inference
	t.Logf("Inference created %d new atoms", len(newAtoms))
}

func TestCreateDefaultPipeline(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	pipelineID, err := engine.CreateDefaultPipeline(tenantID)
	if err != nil {
		t.Fatalf("Failed to create default pipeline: %v", err)
	}
	
	if pipelineID == "" {
		t.Error("Expected non-empty pipeline ID")
	}
	
	pipeline, err := engine.GetPipeline(pipelineID)
	if err != nil {
		t.Fatalf("Failed to get pipeline: %v", err)
	}
	
	if pipeline.TenantID != tenantID {
		t.Errorf("Expected tenant ID %s, got %s", tenantID, pipeline.TenantID)
	}
}

func TestGetStats(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewCognitiveEngine(cfg)
	defer engine.Close()
	
	tenantID := "test-tenant"
	err := engine.InitializeTenant(tenantID)
	if err != nil {
		t.Fatalf("Failed to initialize tenant: %v", err)
	}
	
	// Create some atoms
	engine.CreateConceptNode("Concept1", tenantID)
	engine.CreateConceptNode("Concept2", tenantID)
	
	stats := engine.GetStats(tenantID)
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
	
	if _, ok := stats["config"]; !ok {
		t.Error("Expected 'config' in stats")
	}
	
	if _, ok := stats["sharding"]; !ok {
		t.Error("Expected 'sharding' in stats")
	}
}
