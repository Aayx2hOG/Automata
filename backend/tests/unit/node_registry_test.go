package unit

import (
	"context"
	"testing"

	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"go.uber.org/zap"
)

func TestRegistry_BuildsNodeTypes(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())

	for _, nodeType := range []string{"manual_trigger", "logger", "delay"} {
		n, err := registry.Build(nodeType)
		if err != nil {
			t.Fatalf("expected %q to build, got error: %v", nodeType, err)
		}
		if n == nil {
			t.Fatalf("expected %v to return a non-nil node", &nodeType)
		}
	}
}

func TestRegistry_UnknownNodeTypeReturnsError(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())

	_, err := registry.Build("does_not_exist")
	if err == nil {
		t.Fatal("expected an error for an unknown node type")
	}
}

func TestLoggerNode_ExecutesSucessfuly(t *testing.T) {
	registry := workflow.NewRegistry(zap.NewNop())
	n, err := registry.Build("logger")
	if err != nil {
		t.Fatalf("build logger node: %v", err)
	}

	ec := &node.ExecutionContext{
		Ctx:     context.Background(),
		NodeID:  "n1",
		Config:  map[string]any{"message": "this is a test"},
		Outputs: map[string]map[string]any{},
	}

	output, err := n.Execute(ec)
	if err != nil {
		t.Fatalf("execute logger node: %v", err)
	}

	if output["logged_message"] != "this is a test" {
		t.Fatalf("expected logged message to match config, got: %v", output)
	}
}
