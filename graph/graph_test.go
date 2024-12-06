package graph_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cesto93/langgraphgo/graph"
	"github.com/stretchr/testify/assert"
)

func TestExampleMessageGraph(t *testing.T) {
	g := graph.NewMessageGraph[[]string]()

	g.AddNode("oracle", func(ctx context.Context, state []string) ([]string, error) {
		return append(state, "1 + 1 equals 2."), nil
	})
	g.AddNode(graph.END, func(_ context.Context, state []string) ([]string, error) {
		return state, nil
	})

	g.AddEdge("oracle", graph.END)
	g.SetEntryPoint("oracle")

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Let's run it!
	res, err := runnable.Invoke(ctx, []string{"What is 1 + 1?"})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, res, []string{"What is 1 + 1?", "1 + 1 equals 2."})
}

func TestMessageGraph(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		buildGraph     func() *graph.MessageGraph[[]string]
		inputMessages  []string
		expectedOutput []string
		expectedError  error
	}{
		{
			name: "Simple graph",
			buildGraph: func() *graph.MessageGraph[[]string] {
				g := graph.NewMessageGraph[[]string]()
				g.AddNode("node1", func(_ context.Context, state []string) ([]string, error) {
					return append(state, "Node 1"), nil
				})
				g.AddNode("node2", func(_ context.Context, state []string) ([]string, error) {
					return append(state, "Node 2"), nil
				})
				g.AddEdge("node1", "node2")
				g.AddEdge("node2", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			inputMessages:  []string{"Input"},
			expectedOutput: []string{"Input", "Node 1", "Node 2"},
			expectedError:  nil,
		},
		{
			name: "Entry point not set",
			buildGraph: func() *graph.MessageGraph[[]string] {
				g := graph.NewMessageGraph[[]string]()
				g.AddNode("node1", func(_ context.Context, state []string) ([]string, error) {
					return state, nil
				})
				return g
			},
			expectedError: graph.ErrEntryPointNotSet,
		},
		{
			name: "Node not found",
			buildGraph: func() *graph.MessageGraph[[]string] {
				g := graph.NewMessageGraph[[]string]()
				g.AddNode("node1", func(_ context.Context, state []string) ([]string, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: fmt.Errorf("%w: node2", graph.ErrNodeNotFound),
		},
		{
			name: "No outgoing edge",
			buildGraph: func() *graph.MessageGraph[[]string] {
				g := graph.NewMessageGraph[[]string]()
				g.AddNode("node1", func(_ context.Context, state []string) ([]string, error) {
					return state, nil
				})
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: fmt.Errorf("%w: node1", graph.ErrNoOutgoingEdge),
		},
		{
			name: "Error in node function",
			buildGraph: func() *graph.MessageGraph[[]string] {
				g := graph.NewMessageGraph[[]string]()
				g.AddNode("node1", func(_ context.Context, _ []string) ([]string, error) {
					return nil, errors.New("node error")
				})
				g.AddEdge("node1", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: errors.New("error in node node1: node error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := tc.buildGraph()
			runnable, err := g.Compile()
			if err != nil {
				if tc.expectedError == nil || !errors.Is(err, tc.expectedError) {
					t.Fatalf("unexpected compile error: %v", err)
				}
				return
			}

			output, err := runnable.Invoke(context.Background(), tc.inputMessages)
			if err != nil {
				if tc.expectedError == nil || err.Error() != tc.expectedError.Error() {
					t.Fatalf("unexpected invoke error: '%v', expected '%v'", err, tc.expectedError)
				}
				return
			}

			if tc.expectedError != nil {
				t.Fatalf("expected error %v, but got nil", tc.expectedError)
			}

			if len(output) != len(tc.expectedOutput) {
				t.Fatalf("expected output length %d, but got %d", len(tc.expectedOutput), len(output))
			}

			for i, msg := range output {
				got := msg
				expected := tc.expectedOutput[i]
				if got != expected {
					t.Errorf("expected output[%d] content %q, but got %q", i, expected, got)
				}
			}
		})
	}
}
