package unit

import (
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/workflow"
)

func TestValidateGraph_ValidLinearWorkflow(t *testing.T) {
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "webhook_trigger"},
			{ID: "n2", Type: "http_request"},
			{ID: "n3", Type: "logger"},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n2"},
			{FromNodeID: "n2", ToNodeID: "n3"},
		},
	}

	result := workflow.ValidatorGraph(graph)
	if !result.Valid {
		t.Fatalf("expected valid graph, got errors: %+v", result.Errors)
	}
}

func TestValidateGraph_BranchingWorkflow(t *testing.T) {
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "webhook_trigger"},
			{ID: "n2", Type: "condition"},
			{ID: "n3", Type: "discord_webhook"},
			{ID: "n4", Type: "database_query"},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n2"},
			{FromNodeID: "n2", ToNodeID: "n3", Condition: "true"},
			{FromNodeID: "n2", ToNodeID: "n4", Condition: "false"},
		},
	}

	result := workflow.ValidatorGraph(graph)
	if !result.Valid {
		t.Fatalf("expected valid branching graph, got errors: %+v", result.Errors)
	}
}

func TestValidateGraph_DetectsCycle(t *testing.T) {
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "http_request"},
			{ID: "n2", Type: "logger"},
			{ID: "n3", Type: "delay"},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n2"},
			{FromNodeID: "n2", ToNodeID: "n3"},
			{FromNodeID: "n3", ToNodeID: "n1"},
		},
	}

	result := workflow.ValidatorGraph(graph)
	if result.Valid {
		t.Fatal("expected cycle to be detected, got valid")
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "workflow graph contains a cycle" {
		t.Fatalf("expected a single cycle error, got: %+v", result.Errors)
	}
}

func TestValidateGraph_DetectsDanglingReference(t *testing.T) {
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "webhook_trigger"},
		},
		Edges: []models.GraphEdge{
			{FromNodeID: "n1", ToNodeID: "n999"},
		},
	}

	result := workflow.ValidatorGraph(graph)
	if result.Valid {
		t.Fatal("expected dangling reference to be detected, got valid")
	}
}

func TestValidateGraph_RejectsEmptyWorkflow(t *testing.T) {
	result := workflow.ValidatorGraph(models.WorkflowGraph{})
	if result.Valid {
		t.Fatal("expected empty workflow to be invalid")
	}
}

func TestValidateGraph_DetectsDuplicateNodeID(t *testing.T) {
	graph := models.WorkflowGraph{
		Nodes: []models.GraphNode{
			{ID: "n1", Type: "webhook_trigger"},
			{ID: "n1", Type: "http_request"},
		},
	}

	result := workflow.ValidatorGraph(graph)
	if result.Valid {
		t.Fatal("expected duplicate node id to be detected")
	}
}
