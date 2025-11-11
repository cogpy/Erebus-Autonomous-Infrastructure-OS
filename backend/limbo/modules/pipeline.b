implement Pipeline;

# Pipeline Implementation - Cognitive Processing Pipeline
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
include "agents.m";
	agents: Agents;

# Module definition
Pipeline: module
{
	PATH: con "/dis/limbo/modules/pipeline.dis";
	
	STAGE_INGESTION: con 1;
	STAGE_INFERENCE: con 2;
	STAGE_ATTENTION: con 3;
	STAGE_AGENT_EXEC: con 4;
	STAGE_CUSTOM: con 5;
	
	STATE_IDLE: con 0;
	STATE_RUNNING: con 1;
	STATE_COMPLETED: con 2;
	STATE_FAILED: con 3;
};

Atomspace: import atomspace;
Inference: import inference;
Agents: import agents;

init()
{
	sys = load Sys Sys->PATH;
	str = load String String->PATH;
	daytime = load Daytime Daytime->PATH;
	atomspace = load Atomspace Atomspace->PATH;
	inference = load Inference Inference->PATH;
	agents = load Agents Agents->PATH;
	if (atomspace != nil)
		atomspace->init();
	if (inference != nil)
		inference->init();
	if (agents != nil)
		agents->init();
}

generate_pipeline_id(): string
{
	now := daytime->now();
	return sys->sprint("pipeline_%d_%d", now, sys->pctl(0, nil));
}

# Pipeline implementation
Pipeline.new(name: string, tenantid: string): ref Pipeline
{
	pipeline := ref Pipeline;
	pipeline.id = generate_pipeline_id();
	pipeline.name = name;
	pipeline.stages = array[0] of ref PipelineStage;
	pipeline.state = STATE_IDLE;
	pipeline.tenantid = tenantid;
	pipeline.created = daytime->now();
	return pipeline;
}

Pipeline.add_stage(pipeline: self ref Pipeline, stage: ref PipelineStage): int
{
	# Resize array and add stage
	old_len := len pipeline.stages;
	new_stages := array[old_len + 1] of ref PipelineStage;
	for (i := 0; i < old_len; i++)
		new_stages[i] = pipeline.stages[i];
	new_stages[old_len] = stage;
	pipeline.stages = new_stages;
	return 1;
}

Pipeline.remove_stage(pipeline: self ref Pipeline, index: int): int
{
	if (index < 0 || index >= len pipeline.stages)
		return 0;
	
	new_stages := array[len pipeline.stages - 1] of ref PipelineStage;
	j := 0;
	for (i := 0; i < len pipeline.stages; i++) {
		if (i != index) {
			new_stages[j] = pipeline.stages[i];
			j++;
		}
	}
	pipeline.stages = new_stages;
	return 1;
}

Pipeline.get_stage(pipeline: self ref Pipeline, index: int): ref PipelineStage
{
	if (index < 0 || index >= len pipeline.stages)
		return nil;
	return pipeline.stages[index];
}

Pipeline.execute(pipeline: self ref Pipeline, ctx: ref PipelineContext): int
{
	pipeline.state = STATE_RUNNING;
	
	# Execute each stage in sequence
	for (i := 0; i < len pipeline.stages; i++) {
		stage := pipeline.stages[i];
		if (stage.config.enabled) {
			result := stage.execute(ctx);
			stage.result = result;
			
			if (!result.success) {
				pipeline.state = STATE_FAILED;
				return 0;
			}
		}
	}
	
	pipeline.state = STATE_COMPLETED;
	return 1;
}

Pipeline.execute_async(pipeline: self ref Pipeline, ctx: ref PipelineContext): int
{
	# For now, just call synchronous version
	# In full implementation, would spawn as separate process
	return pipeline.execute(ctx);
}

Pipeline.get_state(pipeline: self ref Pipeline): int
{
	return pipeline.state;
}

Pipeline.reset(pipeline: self ref Pipeline)
{
	pipeline.state = STATE_IDLE;
	for (i := 0; i < len pipeline.stages; i++)
		pipeline.stages[i].result = nil;
}

Pipeline.get_results(pipeline: self ref Pipeline): array of ref StageResult
{
	results := array[len pipeline.stages] of ref StageResult;
	for (i := 0; i < len pipeline.stages; i++)
		results[i] = pipeline.stages[i].result;
	return results;
}

# PipelineStage implementation
PipelineStage.execute(stage: self ref PipelineStage, ctx: ref PipelineContext): ref StageResult
{
	start := daytime->now();
	
	result := ref StageResult;
	result.success = 1;
	result.atoms_processed = 0;
	result.error_msg = "";
	
	# Execute based on stage type
	case stage.config.stagetype {
	STAGE_INFERENCE =>
		execute_inference_stage(stage, ctx, result);
	STAGE_ATTENTION =>
		execute_attention_stage(stage, ctx, result);
	* =>
		result.success = 1;
	}
	
	end := daytime->now();
	result.duration_ms = end - start;
	
	return result;
}

execute_inference_stage(stage: ref PipelineStage, ctx: ref PipelineContext, result: ref StageResult)
{
	if (ctx.inference != nil) {
		infer_result := ctx.inference.run_inference(ctx.tenantid, 5);
		if (infer_result != nil)
			result.atoms_processed = len infer_result.new_atoms;
	}
}

execute_attention_stage(stage: ref PipelineStage, ctx: ref PipelineContext, result: ref StageResult)
{
	# Simple attention update
	if (ctx.atomspace != nil) {
		atoms := ctx.atomspace.query_all(ctx.tenantid);
		result.atoms_processed = len atoms;
	}
}

PipelineStage.validate(stage: self ref PipelineStage): int
{
	if (stage.config == nil)
		return 0;
	if (stage.config.timeout_ms < 0)
		return 0;
	return 1;
}

# PipelineContext implementation
PipelineContext.new(tenantid: string): ref PipelineContext
{
	ctx := ref PipelineContext;
	ctx.tenantid = tenantid;
	ctx.atomspace = nil;
	ctx.inference = nil;
	ctx.scheduler = nil;
	ctx.data = StringTable.new();
	return ctx;
}

PipelineContext.set_data(ctx: self ref PipelineContext, key: string, value: string)
{
	ctx.data.put(key, value);
}

PipelineContext.get_data(ctx: self ref PipelineContext, key: string): string
{
	return ctx.data.get(key);
}

# StringTable implementation
StringTable.new(): ref StringTable
{
	st := ref StringTable;
	st.entries = nil;
	return st;
}

StringTable.put(st: self ref StringTable, key: string, value: string)
{
	# Update if exists
	for (l := st.entries; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k == key) {
			l = (key, value) :: tl l;
			return;
		}
	}
	# Add new
	st.entries = (key, value) :: st.entries;
}

StringTable.get(st: self ref StringTable, key: string): string
{
	for (l := st.entries; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k == key)
			return v;
	}
	return "";
}

StringTable.remove(st: self ref StringTable, key: string)
{
	newlist: list of (string, string);
	for (l := st.entries; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k != key)
			newlist = (k, v) :: newlist;
	}
	st.entries = newlist;
}

# PipelineOrchestrator implementation
PipelineOrchestrator.new(workers: int): ref PipelineOrchestrator
{
	orch := ref PipelineOrchestrator;
	orch.pipelines = nil;
	orch.num_workers = workers;
	return orch;
}

PipelineOrchestrator.create_pipeline(orch: self ref PipelineOrchestrator, name: string, tenantid: string): ref Pipeline
{
	pipeline := Pipeline.new(name, tenantid);
	orch.pipelines = pipeline :: orch.pipelines;
	return pipeline;
}

PipelineOrchestrator.get_pipeline(orch: self ref PipelineOrchestrator, pipeline_id: string): ref Pipeline
{
	for (l := orch.pipelines; l != nil; l = tl l) {
		p := hd l;
		if (p.id == pipeline_id)
			return p;
	}
	return nil;
}

PipelineOrchestrator.delete_pipeline(orch: self ref PipelineOrchestrator, pipeline_id: string): int
{
	newlist: list of ref Pipeline;
	found := 0;
	for (l := orch.pipelines; l != nil; l = tl l) {
		p := hd l;
		if (p.id != pipeline_id)
			newlist = p :: newlist;
		else
			found = 1;
	}
	orch.pipelines = newlist;
	return found;
}

PipelineOrchestrator.list_pipelines(orch: self ref PipelineOrchestrator, tenantid: string): array of ref Pipeline
{
	matches: list of ref Pipeline;
	for (l := orch.pipelines; l != nil; l = tl l) {
		p := hd l;
		if (p.tenantid == tenantid)
			matches = p :: matches;
	}
	
	count := 0;
	for (m := matches; m != nil; m = tl m)
		count++;
	
	arr := array[count] of ref Pipeline;
	i := 0;
	for (m = matches; m != nil; m = tl m) {
		arr[i] = hd m;
		i++;
	}
	
	return arr;
}

PipelineOrchestrator.execute_pipeline(orch: self ref PipelineOrchestrator, pipeline_id: string, ctx: ref PipelineContext): int
{
	pipeline := orch.get_pipeline(pipeline_id);
	if (pipeline == nil)
		return 0;
	
	return pipeline.execute(ctx);
}

PipelineOrchestrator.get_stats(orch: self ref PipelineOrchestrator): string
{
	count := 0;
	for (l := orch.pipelines; l != nil; l = tl l)
		count++;
	
	return sys->sprint("Pipelines: %d, Workers: %d", count, orch.num_workers);
}

# Create default pipeline
create_default_pipeline(tenantid: string): ref Pipeline
{
	pipeline := Pipeline.new("default-cognitive-pipeline", tenantid);
	
	# Add standard stages
	inference_stage := ref PipelineStage;
	inference_stage.config = ref StageConfig;
	inference_stage.config.stagetype = STAGE_INFERENCE;
	inference_stage.config.name = "Inference";
	inference_stage.config.enabled = 1;
	inference_stage.config.timeout_ms = 5000;
	inference_stage.config.retry_count = 0;
	pipeline.add_stage(inference_stage);
	
	attention_stage := ref PipelineStage;
	attention_stage.config = ref StageConfig;
	attention_stage.config.stagetype = STAGE_ATTENTION;
	attention_stage.config.name = "Attention";
	attention_stage.config.enabled = 1;
	attention_stage.config.timeout_ms = 1000;
	attention_stage.config.retry_count = 0;
	pipeline.add_stage(attention_stage);
	
	return pipeline;
}
