package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Avik2024/erebus/backend/internal/cognitive"
	"github.com/Avik2024/erebus/backend/internal/cognitive/atomspace"
)

func main() {
	fmt.Println("=== Erebus Cognitive Engine Demo ===\n")
	
	// Create cognitive engine with default config
	cfg := cognitive.DefaultConfig()
	engine := cognitive.NewCognitiveEngine(cfg)
	defer engine.Close()
	
	fmt.Printf("Initialized cognitive engine with:\n")
	fmt.Printf("  - %d shards\n", cfg.NumShards)
	fmt.Printf("  - %d inference workers\n", cfg.InferenceWorkers)
	fmt.Printf("  - %d agent workers\n", cfg.AgentWorkers)
	fmt.Printf("  - %d pipeline workers\n\n", cfg.PipelineWorkers)
	
	// Initialize tenant
	tenantID := "demo-tenant"
	fmt.Printf("Initializing tenant: %s\n", tenantID)
	if err := engine.InitializeTenant(tenantID); err != nil {
		log.Fatal(err)
	}
	
	// Create a simple knowledge base
	fmt.Println("\n=== Building Knowledge Base ===")
	
	// Create concepts
	fmt.Println("Creating concepts...")
	cat, _ := engine.CreateConceptNode("Cat", tenantID)
	dog, _ := engine.CreateConceptNode("Dog", tenantID)
	mammal, _ := engine.CreateConceptNode("Mammal", tenantID)
	animal, _ := engine.CreateConceptNode("Animal", tenantID)
	livingThing, _ := engine.CreateConceptNode("LivingThing", tenantID)
	
	fmt.Printf("  ✓ Created: Cat, Dog, Mammal, Animal, LivingThing\n")
	
	// Create inheritance links
	fmt.Println("\nCreating inheritance relationships...")
	engine.CreateInheritanceLink(cat.GetID(), mammal.GetID(), tenantID)
	engine.CreateInheritanceLink(dog.GetID(), mammal.GetID(), tenantID)
	engine.CreateInheritanceLink(mammal.GetID(), animal.GetID(), tenantID)
	engine.CreateInheritanceLink(animal.GetID(), livingThing.GetID(), tenantID)
	
	fmt.Println("  ✓ Cat → Mammal")
	fmt.Println("  ✓ Dog → Mammal")
	fmt.Println("  ✓ Mammal → Animal")
	fmt.Println("  ✓ Animal → LivingThing")
	
	// Query initial state
	atoms := engine.QueryAtoms(tenantID, nil)
	fmt.Printf("\nInitial atom count: %d\n", len(atoms))
	
	// Run inference
	fmt.Println("\n=== Running Inference ===")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	newAtoms, err := engine.RunInference(ctx, tenantID, 10)
	if err != nil {
		fmt.Printf("⚠ Inference error: %v\n", err)
		fmt.Println("Continuing with remaining operations...\n")
	} else {
		fmt.Printf("✓ Inference created %d new atoms through deduction!\n", len(newAtoms))
		
		if len(newAtoms) > 0 {
			fmt.Println("\nSome inferred relationships:")
			for i, atom := range newAtoms {
				if i >= 3 {
					break
				}
				if link, ok := atom.(*atomspace.Link); ok {
					if len(link.Outgoing) >= 2 {
						fmt.Printf("  • %s → %s (inferred)\n", 
							link.Outgoing[0].GetName(), 
							link.Outgoing[1].GetName())
					}
				}
			}
		}
	}
	
	// Query final state
	allAtoms := engine.QueryAtoms(tenantID, nil)
	fmt.Printf("\nFinal atom count: %d\n", len(allAtoms))
	
	// Get statistics
	fmt.Println("\n=== System Statistics ===")
	stats := engine.GetStats(tenantID)
	
	if config, ok := stats["config"].(map[string]interface{}); ok {
		fmt.Println("Configuration:")
		fmt.Printf("  Shards: %v\n", config["num_shards"])
		fmt.Printf("  Inference Workers: %v\n", config["inference_workers"])
	}
	
	if tenantStats, ok := stats["tenant"].(map[string]interface{}); ok {
		fmt.Println("\nTenant Statistics:")
		fmt.Printf("  Total Atoms: %v\n", tenantStats["total_atoms"])
		
		if shardDist, ok := tenantStats["shard_distribution"].(map[int]int); ok {
			fmt.Println("  Shard Distribution:")
			for shardID, count := range shardDist {
				if count > 0 {
					fmt.Printf("    Shard %d: %d atoms\n", shardID, count)
				}
			}
		}
	}
	
	// Create and execute a pipeline
	fmt.Println("\n=== Creating Cognitive Pipeline ===")
	pipelineID, err := engine.CreateDefaultPipeline(tenantID)
	if err != nil {
		log.Printf("Pipeline creation error: %v", err)
	} else {
		fmt.Printf("Created pipeline: %s\n", pipelineID)
		
		// Execute pipeline
		fmt.Println("Executing pipeline...")
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()
		
		_, err = engine.ExecutePipeline(ctx2, pipelineID, nil)
		if err != nil {
			log.Printf("Pipeline execution error: %v", err)
		} else {
			fmt.Println("  ✓ Pipeline executed successfully")
		}
	}
	
	// Get agent statistics
	fmt.Println("\n=== Agent Statistics ===")
	agents := engine.GetAgentsByTenant(tenantID)
	fmt.Printf("Active agents: %d\n", len(agents))
	for _, agent := range agents {
		agentStats := agent.GetStats()
		fmt.Printf("  %s: %v runs\n", agentStats["name"], agentStats["run_count"])
	}
	
	// Health check
	health := engine.Health()
	fmt.Println("\n=== Health Check ===")
	fmt.Printf("Status: %v\n", health["status"])
	fmt.Printf("Tenants: %v\n", health["num_tenants"])
	fmt.Printf("Timestamp: %v\n", health["timestamp"])
	
	fmt.Println("\n=== Demo Complete ===")
}
