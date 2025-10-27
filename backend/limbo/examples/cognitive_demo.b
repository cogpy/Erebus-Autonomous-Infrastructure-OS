implement CognitiveDemo;

# Cognitive Architecture Demo
# Demonstrates OpenCog-inspired features in Limbo
# This example shows: AtomSpace, Inference, Agents, and Pipelines

include "sys.m";
	sys: Sys;
include "draw.m";
include "atomspace.m";
	atomspace: Atomspace;
include "inference.m";
	inference: Inference;
include "agents.m";
	agents: Agents;
include "pipeline.m";
	pipeline: Pipeline;
include "disvm.m";
	disvm: DisVM;

CognitiveDemo: module
{
	init: fn(nil: ref Draw->Context, nil: list of string);
};

Atomspace: import atomspace;
Inference: import inference;
Agents: import agents;
Pipeline: import pipeline;
DisVM: import disvm;

init(nil: ref Draw->Context, args: list of string)
{
	sys = load Sys Sys->PATH;
	atomspace = load Atomspace Atomspace->PATH;
	inference = load Inference Inference->PATH;
	agents = load Agents Agents->PATH;
	pipeline = load Pipeline Pipeline->PATH;
	disvm = load DisVM DisVM->PATH;
	
	if (atomspace == nil || inference == nil || agents == nil || pipeline == nil) {
		sys->fprint(sys->fildes(2), "cognitive_demo: failed to load modules\n");
		raise "fail:load";
	}
	
	atomspace->init();
	inference->init();
	agents->init();
	pipeline->init();
	disvm->init();
	
	sys->print("\n=== Erebus Cognitive Architecture Demo (Limbo) ===\n\n");
	
	# Initialize dis-vm runtime
	sys->print("1. Initializing Dis VM Runtime...\n");
	vm_config := disvm->default_config();
	vm := DisVM->Runtime.new(vm_config);
	sys->print("   ✓ VM runtime initialized\n\n");
	
	# Create AtomSpace
	sys->print("2. Creating AtomSpace (Knowledge Store)...\n");
	space := Atomspace->AtomSpace.new();
	tenantid := "demo-tenant";
	sys->print("   ✓ AtomSpace created for tenant: %s\n\n", tenantid);
	
	# Add some concepts
	sys->print("3. Adding Concepts to AtomSpace...\n");
	cat := Atomspace->Atom.new(atomspace->CONCEPT_NODE, "Cat", tenantid);
	cat.truthvalue = Atomspace->TruthValue.new(0.9, 0.8);
	cat_id := space.add_atom(cat);
	sys->print("   ✓ Added concept: Cat (id=%s)\n", cat_id);
	
	animal := Atomspace->Atom.new(atomspace->CONCEPT_NODE, "Animal", tenantid);
	animal.truthvalue = Atomspace->TruthValue.new(1.0, 0.95);
	animal_id := space.add_atom(animal);
	sys->print("   ✓ Added concept: Animal (id=%s)\n", animal_id);
	
	mammal := Atomspace->Atom.new(atomspace->CONCEPT_NODE, "Mammal", tenantid);
	mammal.truthvalue = Atomspace->TruthValue.new(0.95, 0.9);
	mammal_id := space.add_atom(mammal);
	sys->print("   ✓ Added concept: Mammal (id=%s)\n\n", mammal_id);
	
	# Create inheritance links
	sys->print("4. Creating Inheritance Links...\n");
	cat_is_mammal := Atomspace->Atom.new(atomspace->INHERITANCE_LINK, "Cat->Mammal", tenantid);
	cat_is_mammal.outgoing = array[2] of ref Atomspace->Atom;
	cat_is_mammal.outgoing[0] = cat;
	cat_is_mammal.outgoing[1] = mammal;
	cat_is_mammal.truthvalue = Atomspace->TruthValue.new(0.99, 0.95);
	space.add_atom(cat_is_mammal);
	sys->print("   ✓ Created link: Cat → Mammal\n");
	
	mammal_is_animal := Atomspace->Atom.new(atomspace->INHERITANCE_LINK, "Mammal->Animal", tenantid);
	mammal_is_animal.outgoing = array[2] of ref Atomspace->Atom;
	mammal_is_animal.outgoing[0] = mammal;
	mammal_is_animal.outgoing[1] = animal;
	mammal_is_animal.truthvalue = Atomspace->TruthValue.new(1.0, 0.99);
	space.add_atom(mammal_is_animal);
	sys->print("   ✓ Created link: Mammal → Animal\n\n");
	
	# Show AtomSpace statistics
	atom_count := space.atom_count(tenantid);
	sys->print("   AtomSpace contains %d atoms\n\n", atom_count);
	
	# Create inference engine
	sys->print("5. Creating Inference Engine...\n");
	engine := Inference->InferenceEngine.new(space, 4);
	sys->print("   ✓ Inference engine created with 4 workers\n");
	
	# Add inference rules
	deduction := Inference->DeductionRule.new();
	engine.add_rule(deduction);
	sys->print("   ✓ Added deduction rule\n");
	
	induction := Inference->InductionRule.new();
	engine.add_rule(induction);
	sys->print("   ✓ Added induction rule\n\n");
	
	# Run inference
	sys->print("6. Running Inference (max 10 iterations)...\n");
	result := engine.run_inference(tenantid, 10);
	sys->print("   ✓ Inference completed:\n");
	sys->print("     - Iterations: %d\n", result.iterations);
	sys->print("     - Converged: %s\n", result.convergence ? "Yes" : "No");
	sys->print("     - Total atoms: %d\n\n", len result.new_atoms);
	
	# Create agents
	sys->print("7. Creating Cognitive Agents...\n");
	scheduler := Agents->AgentScheduler.new(2);
	
	mind_agent := Agents->MindAgent.new(tenantid, agents->PRIORITY_HIGH);
	scheduler.register_agent(mind_agent.base);
	sys->print("   ✓ Registered MindAgent (priority: HIGH)\n");
	
	attention_agent := Agents->AttentionAgent.new(tenantid, agents->PRIORITY_NORMAL);
	scheduler.register_agent(attention_agent.base);
	sys->print("   ✓ Registered AttentionAgent (priority: NORMAL)\n\n");
	
	# Run agent scheduler
	sys->print("8. Running Agent Scheduler...\n");
	scheduler.start(space, engine);
	sys->print("   ✓ Agents executed\n");
	sys->print("   %s\n\n", scheduler.get_stats());
	
	# Create pipeline
	sys->print("9. Creating Cognitive Processing Pipeline...\n");
	orchestrator := Pipeline->PipelineOrchestrator.new(2);
	demo_pipeline := pipeline->create_default_pipeline(tenantid);
	orchestrator.pipelines = demo_pipeline :: orchestrator.pipelines;
	sys->print("   ✓ Pipeline '%s' created\n", demo_pipeline.name);
	sys->print("   ✓ Pipeline has %d stages\n\n", len demo_pipeline.stages);
	
	# Execute pipeline
	sys->print("10. Executing Pipeline...\n");
	ctx := Pipeline->PipelineContext.new(tenantid);
	ctx.atomspace = space;
	ctx.inference = engine;
	ctx.scheduler = scheduler;
	
	success := demo_pipeline.execute(ctx);
	if (success) {
		sys->print("   ✓ Pipeline executed successfully\n");
		sys->print("   ✓ Final state: ");
		case demo_pipeline.state {
		pipeline->STATE_COMPLETED =>
			sys->print("COMPLETED\n");
		pipeline->STATE_FAILED =>
			sys->print("FAILED\n");
		* =>
			sys->print("UNKNOWN\n");
		}
	}
	
	# Show results
	results := demo_pipeline.get_results();
	sys->print("\n   Pipeline Stage Results:\n");
	for (i := 0; i < len results; i++) {
		if (results[i] != nil) {
			stage := demo_pipeline.stages[i];
			sys->print("   - %s: %s (%d ms, %d atoms)\n",
			          stage.config.name,
			          results[i].success ? "SUCCESS" : "FAILED",
			          results[i].duration_ms,
			          results[i].atoms_processed);
		}
	}
	
	sys->print("\n=== Demo Complete ===\n");
	sys->print("Total atoms in AtomSpace: %d\n", space.atom_count(tenantid));
	sys->print("%s\n", space.get_stats());
	
	# Shutdown
	scheduler.stop();
	vm.shutdown();
	
	sys->print("\n✓ All systems shutdown gracefully\n");
}
