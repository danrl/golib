// Package directedgraph implements an directed graph with nodes (vertices),
// edges, and supporting methods.
package directedgraph

import (
	"bytes"
	"fmt"
	"sync"
)

var (
	// ErrorNodeNotFound is returned when trying to access a non-existent node
	ErrorNodeNotFound = fmt.Errorf("node not found")
	// ErrorNodeAlreadyExists is returned when trying to create duplicate nodes
	ErrorNodeAlreadyExists = fmt.Errorf("node already exists")
	// ErrorGraphIsCyclic is returned when trying to perform an operation on a
	// cyclic graph that requires the graph to be acyclic
	ErrorGraphIsCyclic = fmt.Errorf("graph is cyclic")
)

// DirectedGraph holds a directed graph data structure
type DirectedGraph struct {
	lock  sync.RWMutex
	nodes map[string]interface{}
	edges map[string]map[string]bool
}

// New initializes a new graph
func New() *DirectedGraph {
	return &DirectedGraph{
		nodes: make(map[string]interface{}),
		edges: make(map[string]map[string]bool),
	}
}

// NewNode adds a new node to the graph
func (g *DirectedGraph) NewNode(key string, value interface{}) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.nodes[key]; ok {
		return ErrorNodeAlreadyExists
	}
	g.nodes[key] = value
	g.edges[key] = make(map[string]bool)

	return nil
}

// Value retrieves the value assigned to the node identified by key
func (g *DirectedGraph) Value(key string) (interface{}, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if _, ok := g.nodes[key]; !ok {
		return nil, ErrorNodeNotFound
	}
	value := g.nodes[key]
	return value, nil
}

// UpdateValue sets the value of the node identified by key
func (g *DirectedGraph) UpdateValue(key string, value interface{}) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.nodes[key]; !ok {
		return ErrorNodeNotFound
	}
	g.nodes[key] = value
	return nil
}

// NewEdge adds an edge between to nodes in the graph
func (g *DirectedGraph) NewEdge(from, to string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.nodes[from]; !ok {
		return ErrorNodeNotFound
	}
	if _, ok := g.nodes[to]; !ok {
		return ErrorNodeNotFound
	}

	g.edges[from][to] = true
	return nil
}

// Edges returns the keys of nodes that are directly connected to the node
func (g *DirectedGraph) Edges(from string) ([]string, error) {
	var edges []string

	g.lock.RLock()
	if _, ok := g.nodes[from]; !ok {
		return edges, ErrorNodeNotFound
	}
	for to := range g.edges[from] {
		if g.edges[from][to] {
			edges = append(edges, to)
		}
	}
	g.lock.RUnlock()

	return edges, nil
}

// Nodes returns a list of all nodes in the graph
func (g *DirectedGraph) Nodes() []string {
	g.lock.RLock()
	defer g.lock.RUnlock()

	i := 0
	nodes := make([]string, len(g.nodes))
	for key := range g.nodes {
		nodes[i] = key
		i++
	}
	return nodes
}

// isCyclicDFS recursively tests nodes for back edges in a depth first way. It
// expects a `seen` map that it updates and a `rs` (recursive stack) map that it
// uses to find back edges.
func (g *DirectedGraph) isCyclicDFS(seen, rs map[string]bool, key string) bool {
	seen[key] = true
	if rs[key] {
		return true
	}
	rs[key] = true
	for to, active := range g.edges[key] {
		if active && g.isCyclicDFS(seen, rs, to) {
			return true
		}
		// deactivates the item in the map, which we mis-use as
		// stack here to improve lookup times. we don't care about the order
		// when looking for cycles
		rs[to] = false
	}
	return false
}

// IsCyclic tests a directed graph for cycles and returns true if a cycle has
// been detected
func (g *DirectedGraph) IsCyclic() bool {
	g.lock.RLock()
	defer g.lock.RUnlock()

	seen := make(map[string]bool)
	for key := range g.nodes {
		if seen[key] {
			continue
		}
		rs := make(map[string]bool) // new recursion stack for each partition
		if g.isCyclicDFS(seen, rs, key) {
			return true
		}
	}

	return false
}

// topSort sorts a graph recursively in topological order (non-deterministic)
func (g *DirectedGraph) topSort(seen map[string]bool, order []string, i int, key string) int {
	seen[key] = true

	for to := range g.edges[key] {
		if seen[to] {
			continue
		}
		i = g.topSort(seen, order, i, to)
	}
	order[i] = key
	return i - 1
}

// TopSort returns topological sorted slice of all node keys of the graph. This
// functions returns a list of all nodes in undefined order if the graph happens
// to be cyclic. Test with IsCyclic() before using TopSort() if you want to know
// if there is a valid topological order at all.
func (g *DirectedGraph) TopSort() []string {
	g.lock.RLock()
	defer g.lock.RUnlock()

	order := make([]string, len(g.nodes))
	i := len(order) - 1

	seen := make(map[string]bool)
	for key := range g.nodes {
		if seen[key] {
			continue
		}
		i = g.topSort(seen, order, i, key)
	}
	return order
}

// String returns a human readable multi-line string describing the graph
func (g *DirectedGraph) String() string {
	var out bytes.Buffer

	g.lock.RLock()
	for key, value := range g.nodes {
		out.WriteString(fmt.Sprintf("⦿ `%v` (%v)\n", key, value))
		for to, active := range g.edges[key] {
			if active {
				out.WriteString(fmt.Sprintf("⤷ `%v`\n", to))
			}
		}
	}
	g.lock.RUnlock()

	return out.String()
}
