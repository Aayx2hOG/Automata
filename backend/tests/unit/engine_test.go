package unit

import (
	"context"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"go.uber.org/zap"
)

func TestEngine_RunsLinearWorkflow(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())
	engine := workflow.NewEngine(registry)

	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "manual_trigger"},
			{ID: "n2", Type: "logger", Config: map[string]any{"message": "step two"}},
			{ID: "n3", Type: "delay", Config: map[string]any{"seconds": 0.01}},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n2"},
			{FromNodeID: "n2", ToNodeID: "n3"},
		},
	}

	ec := &node.ExecutionContext{Ctx: context.Background()}
	result := engine.Run(ec, graph)

	if result.Error != nil {
		t.Fatalf("expected successful run, got error: %v", result.Error)
	}

	if len(result.Outputs) != 3 {
		t.Fatalf("expected 3 node outputs, got %d", len(result.Outputs))
	}

	if result.Outputs["n2"]["logged_message"] != "step two" {
		t.Fatalf("expected n2 output to reflect its config, got: %v", result.Outputs["n2"])
	}
}

func TestEngine_RejectsInvalidGraph(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())
	engine := workflow.NewEngine(registry)

	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{{ID: "n1", Type: "logger"}},
		Edges: []models.GraphEdge{{FromNodeID: "n1", ToNodeID: "n1"}},
	}

	ec := &node.ExecutionContext{Ctx: context.Background()}
	result := engine.Run(ec, graph)

	if result.Error == nil {
		t.Fatal("expected engine to reject a cyclic graph before executing anything")
	}
}

func TestEngine_StopsAtFailingNode(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())
	engine := workflow.NewEngine(registry)

	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "manual_trigger"},
			{ID: "n2", Type: "unknown_node_type"},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n2"},
		},
	}
	ec := &node.ExecutionContext{Ctx: context.Background()}
	result := engine.Run(ec, graph)

	if result.Error == nil {
		t.Fatal("expected engine to fail on an unknown node type")
	}
	if result.FailedNodeID != "n2" {
		t.Fatalf("expected failure attributed to n2, got %q", result.FailedNodeID)
	}
	if _, ran := result.Outputs["n1"]; !ran {
		t.Fatal("expected n1 to have executed successfully before n2 failed")
	}
}
