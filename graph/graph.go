package graph

import (
	"context"
	"errors"
	"fmt"
)

// END is a special constant used to represent the end node in the graph.
const END = "END"

var (
	// ErrEntryPointNotSet is returned when the entry point of the graph is not set.
	ErrEntryPointNotSet = errors.New("entry point not set")

	// ErrNodeNotFound is returned when a node is not found in the graph.
	ErrNodeNotFound = errors.New("node not found")

	// ErrNoOutgoingEdge is returned when no outgoing edge is found for a node.
	ErrNoOutgoingEdge = errors.New("no outgoing edge found for node")
)

// Node represents a node in the message graph.
type Node[T any] struct {
	// Name is the unique identifier for the node.
	Name string

	// Function is the function associated with the node.
	Function func(ctx context.Context, state T) (T, error)
}

// Edge represents an edge in the message graph.
type Edge struct {
	// From is the name of the node from which the edge originates.
	From string

	// To is the name of the node to which the edge points.
	To string
}

// MessageGraph represents a message graph.
type MessageGraph[T any] struct {
	// nodes is a map of node names to their corresponding Node objects.
	nodes map[string]Node[T]

	// edges is a slice of Edge objects representing the connections between nodes.
	edges []Edge

	// entryPoint is the name of the entry point node in the graph.
	entryPoint string
}

// NewMessageGraph creates a new instance of MessageGraph.
func NewMessageGraph[T any]() *MessageGraph[T] {
	g := &MessageGraph[T]{
		nodes: make(map[string]Node[T]),
	}

	g.AddNode(END, func(ctx context.Context, state T) (T, error) {
		return state, nil
	})
	return g
}

// AddNode adds a new node to the message graph with the given name and function.
func (g *MessageGraph[T]) AddNode(name string, fn func(ctx context.Context, state T) (T, error)) {
	g.nodes[name] = Node[T]{
		Name:     name,
		Function: fn,
	}
}

// AddEdge adds a new edge to the message graph between the "from" and "to" nodes.
func (g *MessageGraph[T]) AddEdge(from, to string) {
	g.edges = append(g.edges, Edge{
		From: from,
		To:   to,
	})
}

// SetEntryPoint sets the entry point node name for the message graph.
func (g *MessageGraph[T]) SetEntryPoint(name string) {
	g.entryPoint = name
}

// Runnable represents a compiled message graph that can be invoked.
type Runnable[T any] struct {
	// graph is the underlying MessageGraph object.
	graph *MessageGraph[T]
}

// Compile compiles the message graph and returns a Runnable instance.
// It returns an error if the entry point is not set.
func (g *MessageGraph[T]) Compile() (*Runnable[T], error) {
	if g.entryPoint == "" {
		return nil, ErrEntryPointNotSet
	}

	return &Runnable[T]{
		graph: g,
	}, nil
}

// Invoke executes the compiled message graph with the given input messages.
// It returns the resulting state and an error if any occurs during the execution.
// Invoke executes the compiled message graph with the given input messages.
// It returns the resulting state and an error if any occurs during the execution.
func (r *Runnable[T]) Invoke(ctx context.Context, state T) (T, error) {
	currentNode := r.graph.entryPoint

	for {
		if currentNode == END {
			break
		}

		node, ok := r.graph.nodes[currentNode]
		if !ok {
			return state, fmt.Errorf("%w: %s", ErrNodeNotFound, currentNode)
		}

		var err error
		state, err = node.Function(ctx, state)
		if err != nil {
			return state, fmt.Errorf("error in node %s: %w", currentNode, err)
		}

		foundNext := false
		for _, edge := range r.graph.edges {
			if edge.From == currentNode {
				currentNode = edge.To
				foundNext = true
				break
			}
		}

		if !foundNext {
			return state, fmt.Errorf("%w: %s", ErrNoOutgoingEdge, currentNode)
		}
	}

	return state, nil
}
