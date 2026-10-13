package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
)

func TestMappingPreservesTypesAndRejectsMissingFields(t *testing.T) {
	definition := models.GraphNode{Config: map[string]any{"body": "default"}, Inputs: map[string]models.OutputReference{
		"body": {NodeID: "source", Path: []string{"items", "0", "count"}},
	}}
	outputs := map[string]map[string]any{"source": {"items": []any{map[string]any{"count": float64(3)}}}}
	config, err := resolveConfig(definition, outputs)
	if err != nil || config["body"] != float64(3) {
		t.Fatalf("config=%v err=%v", config, err)
	}
	if definition.Config["body"] != "default" {
		t.Fatal("mutated saved definition")
	}
	definition.Inputs["body"] = models.OutputReference{NodeID: "source", Path: []string{"items", "1"}}
	if _, err := resolveConfig(definition, outputs); err == nil {
		t.Fatal("accepted missing index")
	}
	if _, err := resolveConfig(definition, nil); err == nil {
		t.Fatal("accepted missing source")
	}
}

func TestMappingRequiresUpstreamNode(t *testing.T) {
	graph := models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "source"}, {ID: "target", Inputs: map[string]models.OutputReference{"body": {NodeID: "source", Path: []string{"body"}}}}}}
	if ValidatorGraph(graph).Valid {
		t.Fatal("accepted unconnected source")
	}
	graph.Edges = []models.GraphEdge{{FromNodeID: "source", ToNodeID: "target"}}
	if result := ValidatorGraph(graph); !result.Valid {
		t.Fatal(result.Errors)
	}
	graph.Nodes[1].Inputs["credential_id"] = graph.Nodes[1].Inputs["body"]
	if ValidatorGraph(graph).Valid {
		t.Fatal("accepted dynamic credential")
	}
}

func TestMappedInputReachesNodeAndMissingInputPreventsExecution(t *testing.T) {
	calls := 0
	registry := &Registry{constructors: map[string]NodeConstructor{"write": func() node.Node {
		return nodeFunc(func(ctx *node.ExecutionContext) (map[string]any, error) {
			calls++
			if ctx.Config["body"] != "customer-123" {
				t.Fatalf("unresolved config: %v", ctx.Config)
			}
			return map[string]any{"done": true}, nil
		})
	}}}
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{{ID: "hook", Type: "webhook_trigger"}, {ID: "write", Type: "write", Inputs: map[string]models.OutputReference{"body": {NodeID: "hook", Path: []string{"customer", "id"}}}, Retry: &models.RetryPolicy{MaxAttempts: 3}}},
		Edges: []models.GraphEdge{{FromNodeID: "hook", ToNodeID: "write"}},
	}
	engine := NewEngine(registry)
	ctx := &node.ExecutionContext{Ctx: context.Background()}
	result := engine.RunWithSeededOutput(ctx, graph, "webhook_trigger", map[string]any{"customer": map[string]any{"id": "customer-123"}})
	if result.Error != nil || calls != 1 {
		t.Fatalf("mapped execution: calls=%d err=%v", calls, result.Error)
	}
	result = engine.RunWithSeededOutput(ctx, graph, "webhook_trigger", map[string]any{})
	if result.Error == nil || calls != 1 {
		t.Fatalf("missing input executed: calls=%d err=%v", calls, result.Error)
	}
}

func TestPermanentFailureStopsRetries(t *testing.T) {
	calls := 0
	registry := &Registry{constructors: map[string]NodeConstructor{"step": func() node.Node {
		return nodeFunc(func(*node.ExecutionContext) (map[string]any, error) {
			calls++
			return nil, &node.RetryError{Message: "HTTP 401"}
		})
	}}}
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: context.Background()}, models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "step", Type: "step", Retry: &models.RetryPolicy{MaxAttempts: 5}}}})
	if result.Error == nil || calls != 1 {
		t.Fatalf("calls=%d error=%v", calls, result.Error)
	}
}

func TestRetryAfterHonorsCancellation(t *testing.T) {
	calls := 0
	registry := &Registry{constructors: map[string]NodeConstructor{"step": func() node.Node {
		return nodeFunc(func(*node.ExecutionContext) (map[string]any, error) {
			calls++
			return nil, &node.RetryError{Message: "HTTP 429", Retryable: true, After: time.Hour}
		})
	}}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	result := NewEngine(registry).Run(&node.ExecutionContext{Ctx: ctx}, models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "step", Type: "step", Retry: &models.RetryPolicy{MaxAttempts: 5}}}})
	if result.Error == nil || calls != 1 {
		t.Fatalf("calls=%d error=%v", calls, result.Error)
	}
}
