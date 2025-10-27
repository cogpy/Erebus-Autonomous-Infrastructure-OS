# 🌌 Erebus – Autonomous Infrastructure OS
<img src="./images/erebus.png" alt="Erebus Image">
Erebus is a next-generation **Autonomous Infrastructure Operating System** designed for cloudless, distributed computing.  
It brings together concepts from **Kubernetes, Terraform, Helm, and monitoring** into a unified, self-healing platform with an **integrated OpenCog-inspired cognitive architecture** for truly autonomous operation.

Erebus is a next-generation Autonomous Infrastructure Operating System designed for cloudless, distributed computing environments. It unifies the power of modern infrastructure management by integrating concepts from Kubernetes, Terraform, Helm, and comprehensive monitoring into a single, self-healing, and intelligent platform.
Our vision for Erebus is to revolutionize how distributed systems are managed, moving towards a truly autonomous model where infrastructure provisions, scales, and repairs itself without human intervention, ensuring unparalleled resilience and efficiency.

## ✨ Features

### Infrastructure Management
- **Self-Healing Capabilities**: Automatic detection and resolution of infrastructure failures.
- **Autonomous Resource Allocation**: Intelligent scaling and optimization of resources based on demand and predefined policies.
- **Unified Control Plane**: A single interface to manage compute, storage, and networking across diverse environments.
- **Declarative Infrastructure as Code**: Leverage familiar tools like Terraform and Helm for defining desired states.
- **Real-time Observability**: Integrated monitoring (Prometheus, Grafana) for deep insights into system health and performance.
- **Multi-Cloud & Edge Support**: Designed from the ground up to operate seamlessly across different cloud providers and edge locations.

### 🧠 Cognitive Architecture (NEW!)
Erebus now features a **pure Golang implementation of an OpenCog-inspired cognitive architecture** that enables true autonomous operation:

- **Knowledge Representation**: Hypergraph-based AtomSpace with nodes and links for representing complex relationships
- **Probabilistic Reasoning**: TruthValue system for handling uncertainty and confidence
- **Parallel Inference Engine**: Massively parallel deduction, induction, and abduction across worker pools
- **Dynamic Sharding**: Horizontal scalability with consistent hashing and automatic rebalancing
- **Autonomous Agents**: Agent-zero integration with MindAgent and AttentionAgent for cognitive cycles
- **Pipeline Orchestration**: Flexible pipeline system for composing cognitive processing workflows
- **Multi-Tenant Architecture**: Complete tenant isolation with dedicated inference engines and agents
- **Hyperthread Concurrency**: Channel-based multiplexing across 70+ concurrent workers by default

See [Cognitive Architecture Documentation](./backend/internal/cognitive/README.md) for details.

---

## 📂 Project Structure
```
erebus/
├── backend/                    # Go services (core logic, APIs, system modules)
│   ├── cmd/                   # Entry points for services
│   │   └── erebusd/          # Main server with cognitive engine
│   ├── internal/              # Private app modules
│   │   ├── cognitive/        # 🧠 Cognitive architecture (NEW!)
│   │   │   ├── atomspace/   # Knowledge representation
│   │   │   ├── inference/   # Parallel inference engine
│   │   │   ├── sharding/    # Dynamic sharding system
│   │   │   ├── agents/      # Autonomous agents
│   │   │   ├── pipeline/    # Pipeline orchestration
│   │   │   └── api/         # Cognitive API handlers
│   │   ├── config/          # Configuration management
│   │   ├── health/          # Health checks
│   │   ├── metrics/         # Prometheus metrics
│   │   └── ...              # Other modules
│   ├── examples/             # Example programs
│   └── pkg/                  # Public reusable packages
├── deploy/                    # Infrastructure (Terraform, Helm, K8s manifests)
├── docs/                      # Documentation & design notes
├── frontend/                  # Web UI (Next.js/React)
└── monitoring/                # Observability stack (Prometheus, Grafana, etc.)
```

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.23.0 or higher
- **Make**: For building the project
- **jq**: For running the API test script (optional)
- **curl**: For API testing (optional)

### Backend (with Cognitive Engine)
```bash
cd backend

# Build the server
make build

# Run the server
make run

# Or run directly
./bin/erebusd

# Run the cognitive demo
go run examples/cognitive_demo.go

# Test the cognitive API (requires server running and jq installed)
./examples/test_cognitive_api.sh
```

### Cognitive API Examples

Once the server is running, you can interact with the cognitive engine:

```bash
# Initialize a tenant
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/init

# Create a concept
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/concepts \
  -H "Content-Type: application/json" \
  -d '{"name": "Cat"}'

# Run inference
curl -X POST http://localhost:8080/api/cognitive/tenants/my-tenant/inference \
  -H "Content-Type: application/json" \
  -d '{"max_iterations": 10}'

# Get tenant statistics
curl http://localhost:8080/api/cognitive/tenants/my-tenant/stats
```

See [Cognitive Architecture Documentation](./backend/internal/cognitive/README.md) for complete API reference.

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Deploy
```bash
cd deploy
terraform init
terraform apply
```

## 🛠️ Tech Stack

**Backend**: Go (Golang)
- Chi router for HTTP handling
- Zap for structured logging
- Prometheus for metrics
- Custom cognitive architecture

**Cognitive Engine**:
- Pure Golang implementation
- Channel-based concurrency
- Hypergraph knowledge representation
- Parallel inference engine
- Agent-based automation

**Frontend**: Next.js + TypeScript

**Infrastructure**: Terraform + Kubernetes + Helm

**Monitoring**: Prometheus + Grafana

**CI/CD**: GitHub Actions (planned)


📖 Documentation

All project documentation is inside the /docs
 directory.

### 2️⃣ Initialize Git and Commit
Run these commands:

```bash
cd ~/projects/erebus
git init
git add .
git commit -m "Initial commit: Erebus project structure with README"

3️⃣ Push to GitHub

Create a new empty repo on GitHub named erebus.
Then connect and push:

git remote add origin git@github.com:Avik2024/erebus.git
git branch -M main
git push -u origin main
```
📄 License
Erebus is open-source and licensed under the MIT License.



