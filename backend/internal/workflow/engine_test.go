package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
)

type nodeFunc func(*node.ExecutionContext) (map[string]any, error)

func (f nodeFunc) Execute(ctx *node.ExecutionContext) (map[string]any, error) { return f(ctx) }

func TestPersistedSeedBecomesTriggerOutput(t *testing.T) {
	registry := &Registry{constructors: map[string]NodeConstructor{
		"inspect": func() node.Node {
			return nodeFunc(func(ctx *node.ExecutionContext) (map[string]any, error) {
				if got, ok := ctx.Output("hook", "payload"); !ok || got != "persisted" {
					t.Fatalf("downstream lost webhook seed: %v", ctx.Outputs)
				}
				return map[string]any{"received": true}, nil
			})
		},
	}}
	// The trigger has no registered constructor: seeded execution must bypass it.
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{{ID: "hook", Type: "webhook_trigger"}, {ID: "child", Type: "inspect"}},
		Edges: []models.GraphEdge{{FromNodeID: "hook", ToNodeID: "child"}},
	}
	result := NewEngine(registry).RunWithSeededOutput(&node.ExecutionContext{Ctx: context.Background()}, graph, "webhook_trigger", map[string]any{"payload": "persisted"})
	if result.Error != nil || result.Outputs["child"]["received"] != true {
		t.Fatalf("result: %+v", result)
	}
}

func TestJoinWaitsForUnequalBranches(t *testing.T) {
	calls := map[string]int{}
	registry := &Registry{constructors: map[string]NodeConstructor{
		"step": func() node.Node {
			return nodeFunc(func(ctx *node.ExecutionContext) (map[string]any, error) {
				calls[ctx.NodeID]++
				if ctx.NodeID == "join" && (ctx.Outputs["short"] == nil || ctx.Outputs["long2"] == nil) {
					t.Fatalf("join ran before both branches: %v", ctx.Outputs)
				}
				return map[string]any{"done": true}, nil
			})
		},
	}}
	graph := models.WorkflowGraph{}
	for _, id := range []string{"root", "short", "long1", "long2", "join"} {
		graph.Nodes = append(graph.Nodes, models.GraphNode{ID: id, Type: "step"})
	}
	for _, edge := range [][2]string{{"root", "short"}, {"root", "long1"}, {"short", "join"}, {"long1", "long2"}, {"long2", "join"}} {
		graph.Edges = append(graph.Edges, models.GraphEdge{FromNodeID: edge[0], ToNodeID: edge[1]})
	}
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, graph)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	for _, n := range graph.Nodes {
		if calls[n.ID] != 1 {
			t.Fatalf("%s executed %d times", n.ID, calls[n.ID])
		}
	}
}

func TestJoinResolvesSkippedBranches(t *testing.T) {
	registry := &Registry{constructors: map[string]NodeConstructor{
		"step": func() node.Node {
			return nodeFunc(func(ctx *node.ExecutionContext) (map[string]any, error) {
				if ctx.NodeID == "skipped" || ctx.NodeID == "skippedChild" {
					t.Fatalf("inactive branch executed: %s", ctx.NodeID)
				}
				return map[string]any{"result": true}, nil
			})
		},
	}}
	graph := models.WorkflowGraph{}
	for _, id := range []string{"root", "selected", "skipped", "skippedChild", "join"} {
		graph.Nodes = append(graph.Nodes, models.GraphNode{ID: id, Type: "step"})
	}
	graph.Edges = []models.GraphEdge{
		{FromNodeID: "root", ToNodeID: "selected", Condition: "true"},
		{FromNodeID: "root", ToNodeID: "skipped", Condition: "false"},
		{FromNodeID: "selected", ToNodeID: "join"},
		{FromNodeID: "skipped", ToNodeID: "skippedChild"},
		{FromNodeID: "skippedChild", ToNodeID: "join"},
	}
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, graph)
	if result.Error != nil || len(result.Outputs) != 3 || result.Outputs["join"] == nil {
		t.Fatalf("result: %+v", result)
	}
	graph.Edges[0].Condition = "false"
	result = NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, graph)
	if result.Error != nil || len(result.Outputs) != 1 {
		t.Fatalf("fully skipped join: %+v", result)
	}
}

func TestCanceledExecutionDoesNotRunNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := NewEngine(&Registry{}).Run(&node.ExecutionContext{Ctx: ctx}, models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "root", Type: "unused"}}})
	if !errors.Is(result.Error, context.Canceled) {
		t.Fatalf("result: %+v", result)
	}
}

func TestRetryOnlyExecutesFailedNode(t *testing.T) {
	calls := map[string]int{}
	registry := &Registry{constructors: map[string]NodeConstructor{
		"step": func() node.Node {
			return nodeFunc(func(ctx *node.ExecutionContext) (map[string]any, error) {
				calls[ctx.NodeID]++
				if ctx.NodeID == "flaky" && calls[ctx.NodeID] < 3 {
					return nil, errors.New("temporary failure")
				}
				return map[string]any{"done": true}, nil
			})
		},
	}}
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{{ID: "write", Type: "step"}, {ID: "flaky", Type: "step", Retry: &models.RetryPolicy{MaxAttempts: 3}}},
		Edges: []models.GraphEdge{{FromNodeID: "write", ToNodeID: "flaky"}},
	}
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, graph)
	if result.Error != nil || calls["write"] != 1 || calls["flaky"] != 3 {
		t.Fatalf("calls=%v error=%v", calls, result.Error)
	}
}

func TestDefaultPolicyDoesNotRetry(t *testing.T) {
	calls := 0
	registry := &Registry{constructors: map[string]NodeConstructor{
		"write": func() node.Node {
			return nodeFunc(func(*node.ExecutionContext) (map[string]any, error) {
				calls++
				return nil, errors.New("unknown delivery outcome")
			})
		},
	}}
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "write", Type: "write"}}})
	if result.Error == nil || calls != 1 {
		t.Fatalf("default retried a write: calls=%d error=%v", calls, result.Error)
	}
}

func TestGraphResourceAndRetryLimits(t *testing.T) {
	graph := models.WorkflowGraph{Nodes: make([]models.GraphNode, MaxNodes+1)}
	if ValidatorGraph(graph).Valid {
		t.Fatal("accepted oversized graph")
	}
	for _, policy := range []models.RetryPolicy{{MaxAttempts: 0}, {MaxAttempts: 6}, {MaxAttempts: 1, BackoffSeconds: -1}, {MaxAttempts: 1, BackoffSeconds: 31}} {
		graph.Nodes = []models.GraphNode{{ID: "root", Type: "manual_trigger", Retry: &policy}}
		if ValidatorGraph(graph).Valid {
			t.Fatalf("accepted invalid retry policy: %+v", policy)
		}
	}
}
