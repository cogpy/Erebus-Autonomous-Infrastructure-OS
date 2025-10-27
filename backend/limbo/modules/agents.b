implement Agents;

# Agents Implementation - Autonomous Cognitive Agents
# Pure Inferno Limbo Implementation

include "sys.m";
	sys: Sys;
include "draw.m";
include "string.m";
	str: String;
include "daytime.m";
	daytime: Daytime;
include "atomspace.m";
	atomspace: Atomspace;
include "inference.m";
	inference: Inference;

# Module definition
Agents: module
{
	PATH: con "/dis/limbo/modules/agents.dis";
	
	MIND_AGENT: con 1;
	ATTENTION_AGENT: con 2;
	CUSTOM_AGENT: con 3;
	
	PRIORITY_LOW: con 1;
	PRIORITY_NORMAL: con 5;
	PRIORITY_HIGH: con 10;
	
	STATE_IDLE: con 0;
	STATE_RUNNING: con 1;
	STATE_PAUSED: con 2;
	STATE_STOPPED: con 3;
};

Atomspace: import atomspace;
Inference: import inference;

init()
{
	sys = load Sys Sys->PATH;
	str = load String String->PATH;
	daytime = load Daytime Daytime->PATH;
	atomspace = load Atomspace Atomspace->PATH;
	inference = load Inference Inference->PATH;
	if (atomspace != nil)
		atomspace->init();
	if (inference != nil)
		inference->init();
}

generate_agent_id(): string
{
	now := daytime->now();
	return sys->sprint("agent_%d_%d", now, sys->pctl(0, nil));
}

get_timestamp(): int
{
	return daytime->now();
}

# Agent implementation
Agent.new(config: ref AgentConfig, tenantid: string): ref Agent
{
	agent := ref Agent;
	agent.id = generate_agent_id();
	agent.config = config;
	agent.state = STATE_IDLE;
	agent.tenantid = tenantid;
	agent.stats = ref AgentStats(0, 0, 0, 0, 0);
	return agent;
}

Agent.start(agent: self ref Agent): int
{
	if (agent.state == STATE_RUNNING)
		return 0;
	agent.state = STATE_RUNNING;
	return 1;
}

Agent.stop(agent: self ref Agent): int
{
	if (agent.state == STATE_STOPPED)
		return 0;
	agent.state = STATE_STOPPED;
	return 1;
}

Agent.pause(agent: self ref Agent): int
{
	if (agent.state != STATE_RUNNING)
		return 0;
	agent.state = STATE_PAUSED;
	return 1;
}

Agent.resume(agent: self ref Agent): int
{
	if (agent.state != STATE_PAUSED)
		return 0;
	agent.state = STATE_RUNNING;
	return 1;
}

Agent.run(agent: self ref Agent, space: ref Atomspace->AtomSpace, engine: ref Inference->InferenceEngine): int
{
	if (agent.state != STATE_RUNNING)
		return 0;
	
	start_time := get_timestamp();
	
	# Execute agent logic based on type
	case agent.config.agenttype {
	MIND_AGENT =>
		run_mind_agent(agent, space, engine);
	ATTENTION_AGENT =>
		run_attention_agent(agent, space);
	* =>
		;
	}
	
	# Update statistics
	end_time := get_timestamp();
	duration := end_time - start_time;
	agent.stats.executions++;
	agent.stats.total_time_ms += duration;
	agent.stats.last_execution = end_time;
	if (agent.stats.executions > 0)
		agent.stats.avg_time_ms = agent.stats.total_time_ms / agent.stats.executions;
	
	return 1;
}

run_mind_agent(agent: ref Agent, space: ref Atomspace->AtomSpace, engine: ref Inference->InferenceEngine)
{
	# Execute inference cycle
	result := engine.run_inference(agent.tenantid, 5);
	# Result is used implicitly by updating atomspace
}

run_attention_agent(agent: ref Agent, space: ref Atomspace->AtomSpace)
{
	# Allocate attention based on usage
	atoms := space.query_all(agent.tenantid);
	for (i := 0; i < len atoms; i++) {
		atom := atoms[i];
		# Simple attention decay
		if (atom.attentionvalue.sti > 0)
			atom.attentionvalue.sti = atom.attentionvalue.sti - 1;
	}
}

Agent.get_stats(agent: self ref Agent): ref AgentStats
{
	return agent.stats;
}

Agent.reset_stats(agent: self ref Agent)
{
	agent.stats = ref AgentStats(0, 0, 0, 0, 0);
}

# MindAgent implementation
MindAgent.new(tenantid: string, priority: int): ref MindAgent
{
	config := ref AgentConfig;
	config.agenttype = MIND_AGENT;
	config.name = "MindAgent";
	config.priority = priority;
	config.interval_ms = 1000;  # 1 second
	config.auto_start = 1;
	
	mind := ref MindAgent;
	mind.base = Agent.new(config, tenantid);
	mind.inference_iterations = 10;
	return mind;
}

MindAgent.execute_cycle(agent: self ref MindAgent, space: ref Atomspace->AtomSpace, engine: ref Inference->InferenceEngine): int
{
	return agent.base.run(space, engine);
}

# AttentionAgent implementation
AttentionAgent.new(tenantid: string, priority: int): ref AttentionAgent
{
	config := ref AgentConfig;
	config.agenttype = ATTENTION_AGENT;
	config.name = "AttentionAgent";
	config.priority = priority;
	config.interval_ms = 500;  # 0.5 seconds
	config.auto_start = 1;
	
	attention := ref AttentionAgent;
	attention.base = Agent.new(config, tenantid);
	attention.decay_rate = 0.1;
	attention.spread_factor = 0.5;
	return attention;
}

AttentionAgent.allocate_attention(agent: self ref AttentionAgent, space: ref Atomspace->AtomSpace): int
{
	atoms := space.query_all(agent.base.tenantid);
	
	# Simple attention allocation based on connectivity
	for (i := 0; i < len atoms; i++) {
		atom := atoms[i];
		# Atoms with more connections get more attention
		connection_boost := len atom.outgoing + len atom.incoming;
		atom.attentionvalue.sti += connection_boost;
	}
	
	return 1;
}

AttentionAgent.spread_attention(agent: self ref AttentionAgent, space: ref Atomspace->AtomSpace, from_atom: ref Atomspace->Atom): int
{
	if (from_atom == nil)
		return 0;
	
	amount := int(real(from_atom.attentionvalue.sti) * agent.spread_factor);
	return space.spread_attention(from_atom.id, amount);
}

AttentionAgent.decay_attention(agent: self ref AttentionAgent, space: ref Atomspace->AtomSpace): int
{
	atoms := space.query_all(agent.base.tenantid);
	
	for (i := 0; i < len atoms; i++) {
		atom := atoms[i];
		decay := int(real(atom.attentionvalue.sti) * agent.decay_rate);
		atom.attentionvalue.sti -= decay;
		if (atom.attentionvalue.sti < 0)
			atom.attentionvalue.sti = 0;
	}
	
	return 1;
}

# AgentScheduler implementation
AgentScheduler.new(workers: int): ref AgentScheduler
{
	scheduler := ref AgentScheduler;
	scheduler.agents = nil;
	scheduler.num_workers = workers;
	scheduler.running = 0;
	return scheduler;
}

AgentScheduler.register_agent(scheduler: self ref AgentScheduler, agent: ref Agent): int
{
	scheduler.agents = agent :: scheduler.agents;
	return 1;
}

AgentScheduler.unregister_agent(scheduler: self ref AgentScheduler, agent_id: string): int
{
	newlist: list of ref Agent;
	for (l := scheduler.agents; l != nil; l = tl l) {
		agent := hd l;
		if (agent.id != agent_id)
			newlist = agent :: newlist;
	}
	scheduler.agents = newlist;
	return 1;
}

AgentScheduler.get_agent(scheduler: self ref AgentScheduler, agent_id: string): ref Agent
{
	for (l := scheduler.agents; l != nil; l = tl l) {
		agent := hd l;
		if (agent.id == agent_id)
			return agent;
	}
	return nil;
}

AgentScheduler.list_agents(scheduler: self ref AgentScheduler): array of ref Agent
{
	count := 0;
	for (l := scheduler.agents; l != nil; l = tl l)
		count++;
	
	arr := array[count] of ref Agent;
	i := 0;
	for (l = scheduler.agents; l != nil; l = tl l) {
		arr[i] = hd l;
		i++;
	}
	
	return arr;
}

AgentScheduler.start(scheduler: self ref AgentScheduler, space: ref Atomspace->AtomSpace, engine: ref Inference->InferenceEngine)
{
	scheduler.running = 1;
	
	# Execute all agents (in a real implementation, this would be async)
	for (l := scheduler.agents; l != nil; l = tl l) {
		agent := hd l;
		if (agent.config.auto_start)
			agent.start();
		if (agent.state == STATE_RUNNING)
			agent.run(space, engine);
	}
}

AgentScheduler.stop(scheduler: self ref AgentScheduler)
{
	scheduler.running = 0;
	
	# Stop all agents
	for (l := scheduler.agents; l != nil; l = tl l) {
		agent := hd l;
		agent.stop();
	}
}

AgentScheduler.get_stats(scheduler: self ref AgentScheduler): string
{
	count := 0;
	for (l := scheduler.agents; l != nil; l = tl l)
		count++;
	
	return sys->sprint("Agents: %d, Workers: %d, Running: %d", 
	                   count, scheduler.num_workers, scheduler.running);
}
