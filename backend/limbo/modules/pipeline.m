# Pipeline Module - Cognitive Processing Pipeline Orchestration
# Pure Inferno Limbo Implementation
#
# This module provides flexible pipeline framework for orchestrating
# cognitive processing stages (ingestion, inference, attention, agents).

Pipeline: module
{
	PATH: con "/dis/limbo/modules/pipeline.dis";

	# Import required modules
	Atomspace: import Atomspace;
	Inference: import Inference;
	Agents: import Agents;

	# Pipeline stage types
	STAGE_INGESTION: con 1;
	STAGE_INFERENCE: con 2;
	STAGE_ATTENTION: con 3;
	STAGE_AGENT_EXEC: con 4;
	STAGE_CUSTOM: con 5;

	# Pipeline states
	STATE_IDLE: con 0;
	STATE_RUNNING: con 1;
	STATE_COMPLETED: con 2;
	STATE_FAILED: con 3;

	# StageConfig holds stage configuration
	StageConfig: adt {
		stagetype: int;
		name: string;
		enabled: int;      # Boolean
		timeout_ms: int;
		retry_count: int;
	};

	# StageResult represents execution result of a stage
	StageResult: adt {
		success: int;      # Boolean
		duration_ms: int;
		atoms_processed: int;
		error_msg: string;
	};

	# PipelineStage represents a processing stage
	PipelineStage: adt {
		config: ref StageConfig;
		result: ref StageResult;
		
		# Execute stage
		execute: fn(stage: self ref PipelineStage,
		           ctx: ref PipelineContext): ref StageResult;
		
		# Validation
		validate: fn(stage: self ref PipelineStage): int;
	};

	# PipelineContext holds execution context
	PipelineContext: adt {
		tenantid: string;
		atomspace: ref Atomspace->AtomSpace;
		inference: ref Inference->InferenceEngine;
		scheduler: ref Agents->AgentScheduler;
		data: ref StringTable;  # Key-value storage for stage data
		
		# Context operations
		set_data: fn(ctx: self ref PipelineContext, key: string, value: string);
		get_data: fn(ctx: self ref PipelineContext, key: string): string;
	};

	# Pipeline represents a cognitive processing pipeline
	Pipeline: adt {
		id: string;
		name: string;
		stages: array of ref PipelineStage;
		state: int;
		tenantid: string;
		created: int;
		
		# Constructor
		new: fn(name: string, tenantid: string): ref Pipeline;
		
		# Stage management
		add_stage: fn(pipeline: self ref Pipeline, stage: ref PipelineStage): int;
		remove_stage: fn(pipeline: self ref Pipeline, index: int): int;
		get_stage: fn(pipeline: self ref Pipeline, index: int): ref PipelineStage;
		
		# Execution
		execute: fn(pipeline: self ref Pipeline, ctx: ref PipelineContext): int;
		execute_async: fn(pipeline: self ref Pipeline, ctx: ref PipelineContext): int;
		
		# State management
		get_state: fn(pipeline: self ref Pipeline): int;
		reset: fn(pipeline: self ref Pipeline);
		
		# Results
		get_results: fn(pipeline: self ref Pipeline): array of ref StageResult;
	};

	# Standard pipeline stages
	AtomIngestionStage: adt {
		base: ref PipelineStage;
		atoms_to_ingest: array of ref Atomspace->Atom;
		
		new: fn(): ref AtomIngestionStage;
		ingest: fn(stage: self ref AtomIngestionStage,
		          ctx: ref PipelineContext): ref StageResult;
	};

	InferenceStage: adt {
		base: ref PipelineStage;
		max_iterations: int;
		
		new: fn(max_iter: int): ref InferenceStage;
		run_inference: fn(stage: self ref InferenceStage,
		                 ctx: ref PipelineContext): ref StageResult;
	};

	AttentionAllocationStage: adt {
		base: ref PipelineStage;
		decay_rate: real;
		
		new: fn(decay: real): ref AttentionAllocationStage;
		allocate: fn(stage: self ref AttentionAllocationStage,
		            ctx: ref PipelineContext): ref StageResult;
	};

	AgentExecutionStage: adt {
		base: ref PipelineStage;
		agent_ids: array of string;
		
		new: fn(ids: array of string): ref AgentExecutionStage;
		execute_agents: fn(stage: self ref AgentExecutionStage,
		                  ctx: ref PipelineContext): ref StageResult;
	};

	# PipelineOrchestrator manages multiple pipelines
	PipelineOrchestrator: adt {
		pipelines: list of ref Pipeline;
		num_workers: int;
		
		# Constructor
		new: fn(workers: int): ref PipelineOrchestrator;
		
		# Pipeline management
		create_pipeline: fn(orch: self ref PipelineOrchestrator,
		                   name: string,
		                   tenantid: string): ref Pipeline;
		
		get_pipeline: fn(orch: self ref PipelineOrchestrator,
		                pipeline_id: string): ref Pipeline;
		
		delete_pipeline: fn(orch: self ref PipelineOrchestrator,
		                   pipeline_id: string): int;
		
		list_pipelines: fn(orch: self ref PipelineOrchestrator,
		                  tenantid: string): array of ref Pipeline;
		
		# Execution
		execute_pipeline: fn(orch: self ref PipelineOrchestrator,
		                    pipeline_id: string,
		                    ctx: ref PipelineContext): int;
		
		# Statistics
		get_stats: fn(orch: self ref PipelineOrchestrator): string;
	};

	# Helper data structure
	StringTable: adt {
		new: fn(): ref StringTable;
		put: fn(st: self ref StringTable, key: string, value: string);
		get: fn(st: self ref StringTable, key: string): string;
		remove: fn(st: self ref StringTable, key: string);
	};

	# Utility functions
	init: fn();
	generate_pipeline_id: fn(): string;
	create_default_pipeline: fn(tenantid: string): ref Pipeline;
};
