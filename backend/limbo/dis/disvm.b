implement DisVM;

# Dis Virtual Machine Implementation
# Loads and executes Limbo bytecode (.dis files)

include "sys.m";
	sys: Sys;
include "draw.m";
include "string.m";
	str: String;

DisVM: module
{
	PATH: con "/dis/limbo/dis/disvm.dis";
};

init()
{
	sys = load Sys Sys->PATH;
	str = load String String->PATH;
}

default_config(): ref Config
{
	cfg := ref Config;
	cfg.heap_size = 1024 * 1024 * 10;  # 10 MB
	cfg.stack_size = 1024 * 64;        # 64 KB
	cfg.max_procs = 100;
	cfg.gc_enabled = 1;
	return cfg;
}

# ModuleInstance implementation
ModuleInstance.load(path: string): ref ModuleInstance
{
	inst := ref ModuleInstance;
	inst.name = extract_module_name(path);
	inst.path = path;
	inst.loaded = 0;
	inst.entry_point = 0;
	
	# Read bytecode from file
	fd := sys->open(path, Sys->OREAD);
	if (fd == nil) {
		sys->fprint(sys->fildes(2), "disvm: cannot open %s: %r\n", path);
		return nil;
	}
	
	# Get file size
	(ok, stat) := sys->fstat(fd);
	if (ok < 0) {
		sys->fprint(sys->fildes(2), "disvm: cannot stat %s: %r\n", path);
		return nil;
	}
	
	# Read entire file
	bytecode := array[int stat.length] of byte;
	n := sys->read(fd, bytecode, len bytecode);
	if (n != len bytecode) {
		sys->fprint(sys->fildes(2), "disvm: short read from %s\n", path);
		return nil;
	}
	
	inst.bytecode = bytecode;
	inst.loaded = 1;
	
	sys->fprint(sys->fildes(1), "disvm: loaded module %s (%d bytes)\n", inst.name, len bytecode);
	
	return inst;
}

extract_module_name(path: string): string
{
	# Extract filename from path
	for (i := len path - 1; i >= 0; i--) {
		if (path[i] == '/')
			return path[i+1:];
	}
	return path;
}

ModuleInstance.link(inst: self ref ModuleInstance, deps: array of ref ModuleInstance): int
{
	if (!inst.loaded)
		return 0;
	
	# Placeholder for linking logic
	# In a real implementation, would resolve symbols and dependencies
	sys->fprint(sys->fildes(1), "disvm: linking module %s with %d dependencies\n", 
	           inst.name, len deps);
	
	return 1;
}

ModuleInstance.execute(inst: self ref ModuleInstance, args: array of string): int
{
	if (!inst.loaded)
		return 0;
	
	sys->fprint(sys->fildes(1), "disvm: executing module %s\n", inst.name);
	
	# Placeholder for execution
	# In a real implementation, would:
	# 1. Set up execution environment
	# 2. Initialize module state
	# 3. Call entry point with args
	# 4. Handle return value
	
	# For now, just indicate successful execution
	sys->fprint(sys->fildes(1), "disvm: module %s executed successfully\n", inst.name);
	
	return 1;
}

ModuleInstance.unload(inst: self ref ModuleInstance)
{
	if (!inst.loaded)
		return;
	
	inst.bytecode = nil;
	inst.loaded = 0;
	
	sys->fprint(sys->fildes(1), "disvm: unloaded module %s\n", inst.name);
}

# Runtime implementation
Runtime.new(cfg: ref Config): ref Runtime
{
	if (cfg == nil)
		cfg = default_config();
	
	vm := ref Runtime;
	vm.config = cfg;
	vm.loaded_modules = nil;
	vm.running = 1;
	
	sys->fprint(sys->fildes(1), "disvm: initialized runtime (heap=%d, stack=%d)\n",
	           cfg.heap_size, cfg.stack_size);
	
	return vm;
}

Runtime.load_module(vm: self ref Runtime, path: string): ref ModuleInstance
{
	# Check if already loaded
	name := extract_module_name(path);
	existing := vm.get_module(name);
	if (existing != nil)
		return existing;
	
	# Load new module
	inst := ModuleInstance.load(path);
	if (inst == nil)
		return nil;
	
	# Add to loaded modules list
	vm.loaded_modules = inst :: vm.loaded_modules;
	
	return inst;
}

Runtime.get_module(vm: self ref Runtime, name: string): ref ModuleInstance
{
	for (l := vm.loaded_modules; l != nil; l = tl l) {
		inst := hd l;
		if (inst.name == name)
			return inst;
	}
	return nil;
}

Runtime.unload_module(vm: self ref Runtime, name: string): int
{
	newlist: list of ref ModuleInstance;
	found := 0;
	
	for (l := vm.loaded_modules; l != nil; l = tl l) {
		inst := hd l;
		if (inst.name != name)
			newlist = inst :: newlist;
		else {
			inst.unload();
			found = 1;
		}
	}
	
	vm.loaded_modules = newlist;
	return found;
}

Runtime.run(vm: self ref Runtime, module_name: string, args: array of string): int
{
	if (!vm.running)
		return 0;
	
	inst := vm.get_module(module_name);
	if (inst == nil) {
		sys->fprint(sys->fildes(2), "disvm: module %s not loaded\n", module_name);
		return 0;
	}
	
	return inst.execute(args);
}

Runtime.shutdown(vm: self ref Runtime)
{
	if (!vm.running)
		return;
	
	sys->fprint(sys->fildes(1), "disvm: shutting down runtime\n");
	
	# Unload all modules
	for (l := vm.loaded_modules; l != nil; l = tl l) {
		inst := hd l;
		inst.unload();
	}
	
	vm.loaded_modules = nil;
	vm.running = 0;
}

# BytecodeOps implementation
BytecodeOps.verify(bytecode: array of byte): int
{
	if (bytecode == nil || len bytecode < 16)
		return 0;
	
	# Check for Dis magic number (placeholder)
	# Real implementation would verify:
	# - Magic number
	# - Version
	# - Checksums
	# - Symbol table integrity
	
	return 1;
}

BytecodeOps.disassemble(bytecode: array of byte): string
{
	if (bytecode == nil)
		return "Error: null bytecode";
	
	# Placeholder disassembly
	return sys->sprint("Dis bytecode: %d bytes\n[Disassembly not yet implemented]", len bytecode);
}

BytecodeOps.get_metadata(bytecode: array of byte): string
{
	if (bytecode == nil)
		return "Error: null bytecode";
	
	# Placeholder metadata extraction
	return sys->sprint("Module metadata: %d bytes total", len bytecode);
}
