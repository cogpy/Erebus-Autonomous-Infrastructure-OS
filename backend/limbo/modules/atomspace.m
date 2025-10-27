# AtomSpace Module - Knowledge Representation for Cognitive Architecture
# Pure Inferno Limbo Implementation
#
# This module provides hypergraph-based knowledge representation
# with nodes, links, truth values, and attention values.

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

	# TruthValue represents probabilistic logic
	TruthValue: adt {
		strength: real;    # Probability [0.0, 1.0]
		confidence: real;  # Confidence [0.0, 1.0]
	};

	# AttentionValue represents cognitive importance
	AttentionValue: adt {
		sti: int;   # Short-term importance
		lti: int;   # Long-term importance
		vlti: int;  # Very long-term importance
	};

	# Atom is the fundamental unit of knowledge
	Atom: adt {
		id: string;
		atomtype: int;
		name: string;
		tenantid: string;
		truthvalue: ref TruthValue;
		attentionvalue: ref AttentionValue;
		outgoing: array of ref Atom;  # For links
		incoming: array of ref Atom;  # Backlinks
		created: int;                  # Unix timestamp
	};

	# AtomSpace is the hypergraph knowledge store
	AtomSpace: adt {
		atoms: list of ref Atom;
		index_by_id: ref HashTable;
		index_by_type: ref HashTable;
		index_by_name: ref HashTable;
		
		# Constructor
		new: fn(): ref AtomSpace;
		
		# Core operations
		add_atom: fn(space: self ref AtomSpace, atom: ref Atom): string;
		get_atom: fn(space: self ref AtomSpace, id: string): ref Atom;
		delete_atom: fn(space: self ref AtomSpace, id: string): int;
		update_atom: fn(space: self ref AtomSpace, id: string, atom: ref Atom): int;
		
		# Query operations
		query_by_type: fn(space: self ref AtomSpace, atomtype: int, tenantid: string): array of ref Atom;
		query_by_name: fn(space: self ref AtomSpace, name: string, tenantid: string): array of ref Atom;
		query_all: fn(space: self ref AtomSpace, tenantid: string): array of ref Atom;
		
		# Truth value operations
		update_truth: fn(space: self ref AtomSpace, id: string, tv: ref TruthValue): int;
		merge_truth: fn(tv1: ref TruthValue, tv2: ref TruthValue): ref TruthValue;
		
		# Attention value operations
		update_attention: fn(space: self ref AtomSpace, id: string, av: ref AttentionValue): int;
		spread_attention: fn(space: self ref AtomSpace, from_id: string, amount: int): int;
		
		# Statistics
		atom_count: fn(space: self ref AtomSpace, tenantid: string): int;
		get_stats: fn(space: self ref AtomSpace): string;
	};

	# Helper data structure
	HashTable: adt {
		new: fn(): ref HashTable;
		put: fn(ht: self ref HashTable, key: string, value: ref Atom);
		get: fn(ht: self ref HashTable, key: string): ref Atom;
		remove: fn(ht: self ref HashTable, key: string);
		size: fn(ht: self ref HashTable): int;
	};

	# Utility functions
	init: fn();
	generate_id: fn(): string;
	current_time: fn(): int;
};
