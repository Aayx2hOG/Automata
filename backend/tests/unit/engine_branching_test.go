package unit

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/workflow"
)

func TestEngine_FollowsOnlyTheActiveBranch(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())
	engine := workflow.NewEngine(registry)

	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "trigger", Type: "manual_trigger"},
			{ID: "check", Type: "condition", Config: map[string]interface{}{
				"source_node": "trigger", "source_field": "triggered",
				"operator": "equals", "value": true,
			}},
			{ID: "yes_branch", Type: "logger", Config: map[string]interface{}{"message": "took yes branch"}},
			{ID: "no_branch", Type: "logger", Config: map[string]interface{}{"message": "took no branch"}},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "trigger", ToNodeID: "check"},
			{FromNodeID: "check", ToNodeID: "yes_branch", Condition: "true"},
			{FromNodeID: "check", ToNodeID: "no_branch", Condition: "false"},
		},
	}

	result := engine.Run(&node.ExecutionContext{Ctx: context.Background()}, graph)

	if result.Error != nil {
		t.Fatalf("expected successful run, got: %v", result.Error)
	}
	if _, ran := result.Outputs["yes_branch"]; !ran {
		t.Error("expected yes_branch to have executed")
	}
	if _, ran := result.Outputs["no_branch"]; ran {
		t.Error("expected no_branch to have been skipped, but it executed")
	}
}
