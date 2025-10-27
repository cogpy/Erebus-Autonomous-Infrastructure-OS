implement Atomspace;

# AtomSpace Implementation - Knowledge Representation
# Pure Inferno Limbo Implementation

include "sys.m";
	sys: Sys;
include "draw.m";
include "string.m";
	str: String;
include "daytime.m";
	daytime: Daytime;

# Module definition
Atomspace: module
{
	PATH: con "/dis/limbo/modules/atomspace.dis";
	
	# Atom types
	CONCEPT_NODE: con 1;
	PREDICATE_NODE: con 2;
	VARIABLE_NODE: con 3;
	INHERITANCE_LINK: con 4;
	SIMILARITY_LINK: con 5;
	EXECUTION_LINK: con 6;
};

init()
{
	sys = load Sys Sys->PATH;
	str = load String String->PATH;
	daytime = load Daytime Daytime->PATH;
}

# Generate unique ID
generate_id(): string
{
	# Use current time and random component
	now := daytime->now();
	return sys->sprint("atom_%d_%d", now, sys->pctl(0, nil));
}

# Get current timestamp
current_time(): int
{
	return daytime->now();
}

# TruthValue implementation
TruthValue.new(strength: real, confidence: real): ref TruthValue
{
	tv := ref TruthValue;
	tv.strength = strength;
	tv.confidence = confidence;
	return tv;
}

# AttentionValue implementation
AttentionValue.new(sti: int, lti: int, vlti: int): ref AttentionValue
{
	av := ref AttentionValue;
	av.sti = sti;
	av.lti = lti;
	av.vlti = vlti;
	return av;
}

# Atom implementation
Atom.new(atomtype: int, name: string, tenantid: string): ref Atom
{
	atom := ref Atom;
	atom.id = generate_id();
	atom.atomtype = atomtype;
	atom.name = name;
	atom.tenantid = tenantid;
	atom.truthvalue = TruthValue.new(1.0, 1.0);
	atom.attentionvalue = AttentionValue.new(0, 0, 0);
	atom.outgoing = array[0] of ref Atom;
	atom.incoming = array[0] of ref Atom;
	atom.created = current_time();
	return atom;
}

# HashTable implementation (simple)
HashTable.new(): ref HashTable
{
	ht := ref HashTable;
	ht.buckets = array[256] of list of (string, ref Atom);
	return ht;
}

hash(key: string): int
{
	h := 0;
	for (i := 0; i < len key; i++)
		h = (h * 31 + key[i]) % 256;
	return h;
}

HashTable.put(ht: self ref HashTable, key: string, value: ref Atom)
{
	idx := hash(key);
	# Simple linear search in bucket
	for (l := ht.buckets[idx]; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k == key) {
			# Update existing entry
			l = (key, value) :: tl l;
			return;
		}
	}
	# Add new entry
	ht.buckets[idx] = (key, value) :: ht.buckets[idx];
}

HashTable.get(ht: self ref HashTable, key: string): ref Atom
{
	idx := hash(key);
	for (l := ht.buckets[idx]; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k == key)
			return v;
	}
	return nil;
}

HashTable.remove(ht: self ref HashTable, key: string)
{
	idx := hash(key);
	newlist: list of (string, ref Atom);
	for (l := ht.buckets[idx]; l != nil; l = tl l) {
		(k, v) := hd l;
		if (k != key)
			newlist = (k, v) :: newlist;
	}
	ht.buckets[idx] = newlist;
}

HashTable.size(ht: self ref HashTable): int
{
	count := 0;
	for (i := 0; i < len ht.buckets; i++) {
		for (l := ht.buckets[i]; l != nil; l = tl l)
			count++;
	}
	return count;
}

# AtomSpace implementation
AtomSpace.new(): ref AtomSpace
{
	space := ref AtomSpace;
	space.atoms = nil;
	space.index_by_id = HashTable.new();
	space.index_by_type = HashTable.new();
	space.index_by_name = HashTable.new();
	return space;
}

AtomSpace.add_atom(space: self ref AtomSpace, atom: ref Atom): string
{
	# Add to main list
	space.atoms = atom :: space.atoms;
	
	# Index by ID
	space.index_by_id.put(atom.id, atom);
	
	# Index by type (composite key)
	typekey := sys->sprint("%d:%s", atom.atomtype, atom.tenantid);
	space.index_by_type.put(typekey, atom);
	
	# Index by name (composite key)
	namekey := sys->sprint("%s:%s", atom.name, atom.tenantid);
	space.index_by_name.put(namekey, atom);
	
	return atom.id;
}

AtomSpace.get_atom(space: self ref AtomSpace, id: string): ref Atom
{
	return space.index_by_id.get(id);
}

AtomSpace.delete_atom(space: self ref AtomSpace, id: string): int
{
	atom := space.get_atom(id);
	if (atom == nil)
		return 0;
	
	# Remove from indices
	space.index_by_id.remove(id);
	
	typekey := sys->sprint("%d:%s", atom.atomtype, atom.tenantid);
	space.index_by_type.remove(typekey);
	
	namekey := sys->sprint("%s:%s", atom.name, atom.tenantid);
	space.index_by_name.remove(namekey);
	
	# Remove from main list
	newlist: list of ref Atom;
	for (l := space.atoms; l != nil; l = tl l) {
		a := hd l;
		if (a.id != id)
			newlist = a :: newlist;
	}
	space.atoms = newlist;
	
	return 1;
}

AtomSpace.update_atom(space: self ref AtomSpace, id: string, atom: ref Atom): int
{
	existing := space.get_atom(id);
	if (existing == nil)
		return 0;
	
	# Update the atom (shallow copy relevant fields)
	existing.name = atom.name;
	existing.truthvalue = atom.truthvalue;
	existing.attentionvalue = atom.attentionvalue;
	existing.outgoing = atom.outgoing;
	
	return 1;
}

AtomSpace.query_by_type(space: self ref AtomSpace, atomtype: int, tenantid: string): array of ref Atom
{
	results: list of ref Atom;
	for (l := space.atoms; l != nil; l = tl l) {
		atom := hd l;
		if (atom.atomtype == atomtype && atom.tenantid == tenantid)
			results = atom :: results;
	}
	
	# Convert list to array
	count := 0;
	for (r := results; r != nil; r = tl r)
		count++;
	
	arr := array[count] of ref Atom;
	i := 0;
	for (r = results; r != nil; r = tl r) {
		arr[i] = hd r;
		i++;
	}
	
	return arr;
}

AtomSpace.query_by_name(space: self ref AtomSpace, name: string, tenantid: string): array of ref Atom
{
	results: list of ref Atom;
	for (l := space.atoms; l != nil; l = tl l) {
		atom := hd l;
		if (atom.name == name && atom.tenantid == tenantid)
			results = atom :: results;
	}
	
	# Convert list to array
	count := 0;
	for (r := results; r != nil; r = tl r)
		count++;
	
	arr := array[count] of ref Atom;
	i := 0;
	for (r = results; r != nil; r = tl r) {
		arr[i] = hd r;
		i++;
	}
	
	return arr;
}

AtomSpace.query_all(space: self ref AtomSpace, tenantid: string): array of ref Atom
{
	results: list of ref Atom;
	for (l := space.atoms; l != nil; l = tl l) {
		atom := hd l;
		if (atom.tenantid == tenantid)
			results = atom :: results;
	}
	
	# Convert list to array
	count := 0;
	for (r := results; r != nil; r = tl r)
		count++;
	
	arr := array[count] of ref Atom;
	i := 0;
	for (r = results; r != nil; r = tl r) {
		arr[i] = hd r;
		i++;
	}
	
	return arr;
}

AtomSpace.update_truth(space: self ref AtomSpace, id: string, tv: ref TruthValue): int
{
	atom := space.get_atom(id);
	if (atom == nil)
		return 0;
	
	atom.truthvalue = tv;
	return 1;
}

AtomSpace.merge_truth(tv1: ref TruthValue, tv2: ref TruthValue): ref TruthValue
{
	# Simple weighted average based on confidence
	w1 := tv1.confidence;
	w2 := tv2.confidence;
	total_w := w1 + w2;
	
	if (total_w == 0.0)
		return TruthValue.new(0.5, 0.0);
	
	new_strength := (tv1.strength * w1 + tv2.strength * w2) / total_w;
	new_confidence := (w1 + w2) / 2.0;  # Average confidence
	
	return TruthValue.new(new_strength, new_confidence);
}

AtomSpace.update_attention(space: self ref AtomSpace, id: string, av: ref AttentionValue): int
{
	atom := space.get_atom(id);
	if (atom == nil)
		return 0;
	
	atom.attentionvalue = av;
	return 1;
}

AtomSpace.spread_attention(space: self ref AtomSpace, from_id: string, amount: int): int
{
	atom := space.get_atom(from_id);
	if (atom == nil)
		return 0;
	
	# Spread attention to outgoing atoms
	for (i := 0; i < len atom.outgoing; i++) {
		target := atom.outgoing[i];
		if (target != nil) {
			target.attentionvalue.sti += amount / len atom.outgoing;
		}
	}
	
	return 1;
}

AtomSpace.atom_count(space: self ref AtomSpace, tenantid: string): int
{
	count := 0;
	for (l := space.atoms; l != nil; l = tl l) {
		atom := hd l;
		if (atom.tenantid == tenantid)
			count++;
	}
	return count;
}

AtomSpace.get_stats(space: self ref AtomSpace): string
{
	total := 0;
	for (l := space.atoms; l != nil; l = tl l)
		total++;
	
	return sys->sprint("Total atoms: %d", total);
}
