# Agents Module - Autonomous Cognitive Agents (AgentZero Integration)
# Pure Inferno Limbo Implementation
#
# This module provides autonomous agents with priority-based scheduling
# for cognitive cycles, attention allocation, and inference execution.

Agents: module
{
	PATH: con "/dis/limbo/modules/agents.dis";

	# Import required modules
	Atomspace: import Atomspace;
	Inference: import Inference;

	# Agent types
	MIND_AGENT: con 1;
	ATTENTION_AGENT: con 2;
	CUSTOM_AGENT: con 3;

	# Agent priorities
	PRIORITY_LOW: con 1;
	PRIORITY_NORMAL: con 5;
	PRIORITY_HIGH: con 10;

	# Agent states
	STATE_IDLE: con 0;
	STATE_RUNNING: con 1;
	STATE_PAUSED: con 2;
	STATE_STOPPED: con 3;

	# AgentConfig holds agent configuration
	AgentConfig: adt {
		agenttype: int;
		name: string;
		priority: int;
		interval_ms: int;  # Execution interval in milliseconds
		auto_start: int;   # Boolean: 1=auto-start, 0=manual
	};

	# AgentStats tracks agent statistics
	AgentStats: adt {
		executions: int;
		total_time_ms: int;
		last_execution: int;  # Unix timestamp
		avg_time_ms: int;
		errors: int;
	};

	# Agent represents an autonomous cognitive agent
	Agent: adt {
		id: string;
		config: ref AgentConfig;
		state: int;
		stats: ref AgentStats;
		tenantid: string;
		
		# Lifecycle methods
		start: fn(agent: self ref Agent): int;
		stop: fn(agent: self ref Agent): int;
		pause: fn(agent: self ref Agent): int;
		resume: fn(agent: self ref Agent): int;
		
		# Core execution
		run: fn(agent: self ref Agent, 
		       atomspace: ref Atomspace->AtomSpace,
		       inference: ref Inference->InferenceEngine): int;
		
		# Statistics
		get_stats: fn(agent: self ref Agent): ref AgentStats;
		reset_stats: fn(agent: self ref Agent);
	};

	# MindAgent executes inference cycles periodically
	MindAgent: adt {
		base: ref Agent;
		inference_iterations: int;
		
		# Constructor
		new: fn(tenantid: string, priority: int): ref MindAgent;
		
		# Execute cognitive cycle
		execute_cycle: fn(agent: self ref MindAgent,
		                 atomspace: ref Atomspace->AtomSpace,
		                 inference: ref Inference->InferenceEngine): int;
	};

	# AttentionAgent manages attention allocation across atoms
	AttentionAgent: adt {
		base: ref Agent;
		decay_rate: real;      # Rate of attention decay
		spread_factor: real;   # Factor for attention spreading
		
		# Constructor
		new: fn(tenantid: string, priority: int): ref AttentionAgent;
		
		# Attention allocation
		allocate_attention: fn(agent: self ref AttentionAgent,
		                      atomspace: ref Atomspace->AtomSpace): int;
		
		# Attention spreading
		spread_attention: fn(agent: self ref AttentionAgent,
		                    atomspace: ref Atomspace->AtomSpace,
		                    from_atom: ref Atomspace->Atom): int;
		
		# Attention decay
		decay_attention: fn(agent: self ref AttentionAgent,
		                   atomspace: ref Atomspace->AtomSpace): int;
	};

	# AgentScheduler coordinates execution of multiple agents
	AgentScheduler: adt {
		agents: list of ref Agent;
		num_workers: int;
		running: int;  # Boolean
		
		# Constructor
		new: fn(workers: int): ref AgentScheduler;
		
		# Agent management
		register_agent: fn(scheduler: self ref AgentScheduler, agent: ref Agent): int;
		unregister_agent: fn(scheduler: self ref AgentScheduler, agent_id: string): int;
		get_agent: fn(scheduler: self ref AgentScheduler, agent_id: string): ref Agent;
		list_agents: fn(scheduler: self ref AgentScheduler): array of ref Agent;
		
		# Scheduler control
		start: fn(scheduler: self ref AgentScheduler,
		         atomspace: ref Atomspace->AtomSpace,
		         inference: ref Inference->InferenceEngine);
		stop: fn(scheduler: self ref AgentScheduler);
		
		# Statistics
		get_stats: fn(scheduler: self ref AgentScheduler): string;
	};

	# Utility functions
	init: fn();
	generate_agent_id: fn(): string;
	get_timestamp: fn(): int;
};
