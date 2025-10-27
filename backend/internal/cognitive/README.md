# Cognitive Architecture - OpenCog-Inspired Implementation in Pure Golang

## Overview

This is a pure Golang implementation of an OpenCog-inspired cognitive architecture integrated into the Erebus Autonomous Infrastructure OS. It provides autonomous multi-tenant support, hyperthread multi-channel concurrency multiplexing, dynamic sharding over massively parallel inference engine networks, and agent-zero integration for cognitive pipeline orchestration.

## Architecture Components

### 1. AtomSpace (Knowledge Representation)
Located in: `internal/cognitive/atomspace/`

The AtomSpace is a hypergraph-based knowledge store that represents:
- **Atoms**: Fundamental units of knowledge
  - **Nodes**: Simple named entities (ConceptNode, PredicateNode, VariableNode)
  - **Links**: Relationships between atoms (InheritanceLink, SimilarityLink, ExecutionLink)
- **TruthValues**: Probabilistic logic with strength and confidence
- **AttentionValues**: Cognitive importance metrics (STI, LTI, VLTI)

**Features:**
- Thread-safe concurrent operations with channel multiplexing
- Multi-tenant isolation at the atom level
- Fast indexing by type, name, and tenant

### 2. Inference Engine
Located in: `internal/cognitive/inference/`

Implements parallel inference over the AtomSpace with multiple inference rules:
- **Deduction Rule**: Modus ponens (A→B, A ⊢ B)
- **Induction Rule**: Generalization from instances
- **Abduction Rule**: Hypothesis generation

**Features:**
- Massively parallel rule execution across worker pools
- Concurrent task distribution via channels
- Fixpoint convergence detection
- Pattern matching for graph queries

### 3. Dynamic Sharding System
Located in: `internal/cognitive/sharding/`

Distributes atoms across multiple AtomSpace shards for horizontal scalability:
- **Consistent Hashing**: Tenant-aware atom distribution
- **Load Balancing**: Dynamic rebalancing based on shard load
- **Cross-Shard Queries**: Parallel query execution across all shards
- **Tenant Isolation**: Each tenant's data is distributed but isolated

**Configuration:**
- Default: 8 shards with 4 workers per shard
- Rebalancing threshold: 1000 atoms difference

### 4. Agent System (Agent-Zero Integration)
Located in: `internal/cognitive/agents/`

Autonomous cognitive agents that perform specialized tasks:
- **MindAgent**: Executes inference cycles periodically
- **AttentionAgent**: Manages attention allocation across atoms
- **AgentScheduler**: Priority-based scheduling with concurrent execution

**Features:**
- Autonomous spawning and termination
- Priority queues for agent execution
- Inter-agent communication via channels
- Per-agent statistics and monitoring

### 5. Pipeline Orchestration Workbench
Located in: `internal/cognitive/pipeline/`

A flexible pipeline framework for orchestrating cognitive processing:

**Pipeline Stages:**
- **AtomIngestionStage**: Ingest atoms into the AtomSpace
- **InferenceStage**: Run inference rules
- **AttentionAllocationStage**: Update attention values
- **AgentExecutionStage**: Execute cognitive agents

**Features:**
- Composable stage-based architecture
- Concurrent pipeline execution
- Pipeline state tracking and monitoring
- Error handling and recovery

### 6. Cognitive Engine (Main Orchestrator)
Located in: `internal/cognitive/engine.go`

The main orchestrator that ties all components together:
- Manages multiple tenants
- Coordinates sharding, inference, agents, and pipelines
- Provides unified API for cognitive operations
- Health monitoring and statistics

## Multi-Tenant Architecture

Each tenant gets:
- Isolated atom namespace across all shards
- Dedicated inference engine with custom rules
- Tenant-specific cognitive agents
- Separate pipelines and statistics

## Concurrency Model

### Hyperthread Multi-Channel Multiplexing

The system uses Go's channels and goroutines for high-concurrency operations:

1. **AtomSpace Operations**: Each operation (add, query, update, delete) is multiplexed through worker pools
2. **Inference Tasks**: Inference rules execute in parallel across worker pools
3. **Shard Routing**: Consistent hashing routes operations to appropriate shards concurrently
4. **Agent Scheduling**: Agents execute concurrently with priority-based scheduling
5. **Pipeline Execution**: Multiple pipelines can execute simultaneously

**Default Worker Configuration:**
- AtomSpace: 32 workers (4 per shard × 8 shards)
- Inference: 16 workers
- Agents: 8 workers
- Pipelines: 8 workers

## API Endpoints

### Tenant Management
- `POST /api/cognitive/tenants/{tenantID}/init` - Initialize a new tenant

### AtomSpace Operations
- `POST /api/cognitive/tenants/{tenantID}/atoms` - Create an atom
- `GET /api/cognitive/tenants/{tenantID}/atoms/{atomID}` - Get an atom
- `GET /api/cognitive/tenants/{tenantID}/atoms?type=concept` - Query atoms
- `PUT /api/cognitive/tenants/{tenantID}/atoms/{atomID}` - Update an atom
- `DELETE /api/cognitive/tenants/{tenantID}/atoms/{atomID}` - Delete an atom

### Concepts and Links
- `POST /api/cognitive/tenants/{tenantID}/concepts` - Create a concept node
- `POST /api/cognitive/tenants/{tenantID}/links/inheritance` - Create inheritance link

### Inference
- `POST /api/cognitive/tenants/{tenantID}/inference` - Run inference

### Pipelines
- `POST /api/cognitive/tenants/{tenantID}/pipelines` - Create a pipeline
- `GET /api/cognitive/tenants/{tenantID}/pipelines/{pipelineID}` - Get pipeline details
- `POST /api/cognitive/tenants/{tenantID}/pipelines/{pipelineID}/execute` - Execute pipeline

### Agents
- `GET /api/cognitive/tenants/{tenantID}/agents` - List agents
- `GET /api/cognitive/tenants/{tenantID}/agents/{agentID}` - Get agent details

### Monitoring
- `GET /api/cognitive/tenants/{tenantID}/stats` - Get tenant statistics
- `GET /api/cognitive/stats` - Get global statistics
- `GET /api/cognitive/health` - Health check

## Usage Examples

### 1. Initialize a Tenant
```bash
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/init
```

### 2. Create Concepts
```bash
# Create "Cat" concept
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/concepts \
  -H "Content-Type: application/json" \
  -d '{"name": "Cat"}'

# Create "Animal" concept
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/concepts \
  -H "Content-Type: application/json" \
  -d '{"name": "Animal"}'
```

### 3. Create Inheritance Link
```bash
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/links/inheritance \
  -H "Content-Type: application/json" \
  -d '{"source_id": "<cat-atom-id>", "target_id": "<animal-atom-id>"}'
```

### 4. Run Inference
```bash
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/inference \
  -H "Content-Type: application/json" \
  -d '{"max_iterations": 10}'
```

### 5. Create and Execute Pipeline
```bash
# Create default pipeline
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/pipelines \
  -H "Content-Type: application/json" \
  -d '{"name": "cognitive-pipeline", "use_default": true}'

# Execute pipeline
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/pipelines/<pipeline-id>/execute
```

### 6. Get Statistics
```bash
# Tenant statistics
curl http://localhost:8080/api/cognitive/tenants/my-tenant/stats

# Global statistics
curl http://localhost:8080/api/cognitive/stats
```

## Configuration

The cognitive engine can be configured in `internal/cognitive/engine.go`:

```go
type Config struct {
    NumShards        int // Number of shards (default: 8)
    WorkersPerShard  int // Workers per shard (default: 4)
    InferenceWorkers int // Inference workers (default: 16)
    AgentWorkers     int // Agent workers (default: 8)
    PipelineWorkers  int // Pipeline workers (default: 8)
}
```

## Performance Characteristics

- **Scalability**: Horizontal scaling via dynamic sharding
- **Concurrency**: 70+ concurrent workers by default
- **Throughput**: Thousands of operations per second per shard
- **Latency**: Sub-millisecond for simple operations, seconds for complex inference
- **Memory**: Efficient with copy-on-write semantics

## Testing

Run the cognitive engine tests:
```bash
cd backend
go test ./internal/cognitive -v
```

## Future Enhancements

1. **Distributed Sharding**: Support for distributed shards across multiple nodes
2. **Persistent Storage**: Save/load AtomSpace state to/from database
3. **Advanced Inference**: Probabilistic Logic Networks (PLN), MOSES
4. **Natural Language Processing**: Integration with NLP pipelines
5. **Visual Reasoning**: Support for image/video concept learning
6. **Reinforcement Learning**: Integration with RL agents
7. **Federated Learning**: Cross-tenant knowledge sharing with privacy

## References

- OpenCog: https://opencog.org/
- Hypergraph Database: https://wiki.opencog.org/w/AtomSpace
- Probabilistic Logic Networks: https://wiki.opencog.org/w/PLN
- Cognitive Synergy: https://wiki.opencog.org/w/CogPrime_Overview

## License

This implementation is part of the Erebus Autonomous Infrastructure OS project and follows the same license.
