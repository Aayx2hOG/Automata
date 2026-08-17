package workflow

import (
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
)

type ExecutionResult struct {
	Outputs      map[string]map[string]any
	Error        error
	FailedNodeID string
}

type Engine struct {
	registry *Registry
}

func NewEngine(registry *Registry) *Engine {
	return &Engine{registry: registry}
}

func (e *Engine) Run(ctx *node.ExecutionContext, graph models.WorkflowGraph) *ExecutionResult {
	validation := ValidatorGraph(graph)
	if !validation.Valid {
		return &ExecutionResult{Error: fmt.Errorf("invalid workflow graph: %v", validation.Errors)}
	}

	order, err := topologicalOrder(graph)
	if err != nil {
		return &ExecutionResult{Error: err}
	}

	nodesByID := make(map[string]models.GraphNode, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodesByID[n.ID] = n
	}

	outputs := make(map[string]map[string]any)
	for _, nodeID := range order {
		graphNode := nodesByID[nodeID]

		impl, err := e.registry.Build(graphNode.Type)
		if err != nil {
			return &ExecutionResult{
				Outputs:      outputs,
				Error:        fmt.Errorf("node %q: %w", nodeID, err),
				FailedNodeID: nodeID,
			}
		}

		nodeCtx := &node.ExecutionContext{
			Ctx:     ctx.Ctx,
			NodeID:  nodeID,
			Config:  graphNode.Config,
			Outputs: outputs,
		}

		result, err := impl.Execute(nodeCtx)
		if err != nil {
			return &ExecutionResult{
				Outputs:      outputs,
				Error:        fmt.Errorf("node %q failed: %w", nodeID, err),
				FailedNodeID: nodeID,
			}
		}
		outputs[nodeID] = result
	}
	return &ExecutionResult{Outputs: outputs}
}
