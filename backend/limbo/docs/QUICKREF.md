# Limbo Cognitive Architecture - Quick Reference

## Module Hierarchy

```
DisVM (Virtual Machine Runtime)
  ↓
AtomSpace (Knowledge Representation)
  ↓
Inference (Reasoning Engine)
  ↓
Agents (Autonomous Agents)
  ↓
Pipeline (Orchestration)
```

## Quick Start

### 1. Build Modules
```bash
cd backend/limbo
./build.sh
```

### 2. Run Demo
```bash
emu examples/cognitive_demo.dis
```

## Common Operations

### Create AtomSpace
```limbo
space := Atomspace->AtomSpace.new();
```

### Add Concept
```limbo
concept := Atomspace->Atom.new(
    Atomspace->CONCEPT_NODE,
    "MyThing",
    "tenant-id"
);
id := space.add_atom(concept);
```

### Create Link
```limbo
link := Atomspace->Atom.new(
    Atomspace->INHERITANCE_LINK,
    "A->B",
    "tenant-id"
);
link.outgoing = array[2] of ref Atomspace->Atom;
link.outgoing[0] = atom_a;
link.outgoing[1] = atom_b;
space.add_atom(link);
```

### Run Inference
```limbo
engine := Inference->InferenceEngine.new(space, 4);
engine.add_rule(Inference->DeductionRule.new());
result := engine.run_inference("tenant-id", 10);
```

### Create Agent
```limbo
scheduler := Agents->AgentScheduler.new(2);
agent := Agents->MindAgent.new("tenant-id", Agents->PRIORITY_HIGH);
scheduler.register_agent(agent.base);
scheduler.start(space, engine);
```

### Execute Pipeline
```limbo
orchestrator := Pipeline->PipelineOrchestrator.new(2);
pipeline := Pipeline->create_default_pipeline("tenant-id");
ctx := Pipeline->PipelineContext.new("tenant-id");
ctx.atomspace = space;
ctx.inference = engine;
pipeline.execute(ctx);
```

## Atom Types

| Type | Constant | Description |
|------|----------|-------------|
| Concept | `CONCEPT_NODE` | Named entity |
| Predicate | `PREDICATE_NODE` | Property or relation |
| Variable | `VARIABLE_NODE` | Placeholder in patterns |
| Inheritance | `INHERITANCE_LINK` | "is-a" relationship |
| Similarity | `SIMILARITY_LINK` | "similar-to" relationship |
| Execution | `EXECUTION_LINK` | Function call |

## Truth Values

```limbo
tv := Atomspace->TruthValue.new(
    0.9,  # strength: probability [0.0, 1.0]
    0.8   # confidence: certainty [0.0, 1.0]
);
atom.truthvalue = tv;
```

## Attention Values

```limbo
av := Atomspace->AttentionValue.new(
    100,  # STI: Short-term importance
    50,   # LTI: Long-term importance
    10    # VLTI: Very long-term importance
);
atom.attentionvalue = av;
```

## Inference Rules

| Rule | Type | Description |
|------|------|-------------|
| Deduction | `RULE_DEDUCTION` | (A→B, A) ⊢ B |
| Induction | `RULE_INDUCTION` | Multiple instances ⊢ generalization |
| Abduction | `RULE_ABDUCTION` | (B, A→B) ⊢ hypothesis A |

## Agent Types

| Agent | Type | Function |
|-------|------|----------|
| MindAgent | `MIND_AGENT` | Runs inference cycles |
| AttentionAgent | `ATTENTION_AGENT` | Manages attention allocation |

## Pipeline Stages

| Stage | Type | Function |
|-------|------|----------|
| Ingestion | `STAGE_INGESTION` | Load atoms into AtomSpace |
| Inference | `STAGE_INFERENCE` | Run inference rules |
| Attention | `STAGE_ATTENTION` | Update attention values |
| Agent Exec | `STAGE_AGENT_EXEC` | Execute cognitive agents |

## Module Loading

```limbo
include "atomspace.m";
    atomspace: Atomspace;

init()
{
    atomspace = load Atomspace Atomspace->PATH;
    atomspace->init();
}

Atomspace: import atomspace;
```

## Error Handling

```limbo
atom := space.get_atom(id);
if (atom == nil) {
    sys->fprint(sys->fildes(2), "Error: atom not found\n");
    return;
}
```

## Debugging

```limbo
# Print to stdout
sys->print("Debug: %s\n", message);

# Print to stderr
sys->fprint(sys->fildes(2), "Error: %s\n", error);

# Get stats
stats := space.get_stats();
sys->print("%s\n", stats);
```

## Best Practices

1. **Always init modules** before use
2. **Check for nil** when getting atoms
3. **Use meaningful IDs** for tenants
4. **Set truth values** on new atoms
5. **Clean up** resources when done
6. **Handle errors** gracefully
7. **Monitor stats** for performance

## Performance Tips

- Use 2-8 workers per engine
- Limit inference iterations (10-100)
- Cache frequently used atoms
- Batch operations when possible
- Monitor atom count per tenant
- Clean up unused atoms

## File Organization

```
project/
├── mymodule.m      # Module definition
├── mymodule.b      # Implementation
└── mymodule.dis    # Compiled bytecode (generated)
```

## Compilation Order

1. Compile `.m` files (definitions)
2. Compile `.b` files (implementations)
3. Run `.dis` files (bytecode)

## Common Issues

### Module Not Found
```
Error: module not found
```
**Solution**: Check PATH constant and module location

### Type Mismatch
```
Error: type mismatch
```
**Solution**: Check ADT field types and function signatures

### Nil Reference
```
Error: nil reference
```
**Solution**: Check for nil before dereferencing

## Resources

- [Limbo Documentation](limbo/docs/README.md)
- [Full Examples](limbo/examples/)
- [Module Reference](limbo/modules/)
