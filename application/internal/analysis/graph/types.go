// Package graph defines shared engineering graph types (spec-my-A §24.16).
package graph

// Node is a vertex in an engineering graph.
type Node struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// Edge links two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// Graph is a local structural graph export.
type Graph struct {
	Kind  string `json:"kind"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Bundle aggregates all analysis graphs for a product.
type Bundle struct {
	Product string           `json:"product"`
	Graphs  map[string]Graph `json:"graphs"`
}
