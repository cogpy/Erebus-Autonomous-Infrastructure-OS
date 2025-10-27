# Dis Virtual Machine Loader and Runtime
# This file provides the interface for loading and executing Dis bytecode
# generated from Limbo modules

DisVM: module
{
	PATH: con "/dis/limbo/dis/disvm.dis";

	# Dis VM configuration
	Config: adt {
		heap_size: int;       # Heap size in bytes
		stack_size: int;      # Stack size in bytes
		max_procs: int;       # Maximum number of processes
		gc_enabled: int;      # Garbage collection enabled
	};

	# Module instance
	ModuleInstance: adt {
		name: string;
		path: string;
		bytecode: array of byte;
		loaded: int;          # Boolean
		entry_point: int;     # Offset to entry point
		
		# Load module from .dis file
		load: fn(path: string): ref ModuleInstance;
		
		# Link with other modules
		link: fn(inst: self ref ModuleInstance, deps: array of ref ModuleInstance): int;
		
		# Execute module
		execute: fn(inst: self ref ModuleInstance, args: array of string): int;
		
		# Unload module
		unload: fn(inst: self ref ModuleInstance);
	};

	# VM Runtime
	Runtime: adt {
		config: ref Config;
		loaded_modules: list of ref ModuleInstance;
		running: int;         # Boolean
		
		# Initialize VM
		new: fn(cfg: ref Config): ref Runtime;
		
		# Module management
		load_module: fn(vm: self ref Runtime, path: string): ref ModuleInstance;
		get_module: fn(vm: self ref Runtime, name: string): ref ModuleInstance;
		unload_module: fn(vm: self ref Runtime, name: string): int;
		
		# Execution
		run: fn(vm: self ref Runtime, module_name: string, args: array of string): int;
		
		# Shutdown
		shutdown: fn(vm: self ref Runtime);
	};

	# Bytecode operations
	BytecodeOps: adt {
		# Verify bytecode integrity
		verify: fn(bytecode: array of byte): int;
		
		# Disassemble bytecode (for debugging)
		disassemble: fn(bytecode: array of byte): string;
		
		# Get module metadata
		get_metadata: fn(bytecode: array of byte): string;
	};

	# Default configuration
	default_config: fn(): ref Config;
	
	# Initialize dis-vm
	init: fn();
};
