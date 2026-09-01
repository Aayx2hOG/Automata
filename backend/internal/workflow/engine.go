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

type seed struct {
	nodeType string
	data     map[string]any
}

func NewEngine(registry *Registry) *Engine {
	return &Engine{registry: registry}
}

func (e *Engine) Run(ctx *node.ExecutionContext, graph models.WorkflowGraph) *ExecutionResult {
	return e.run(ctx, graph, nil)
}

func (e *Engine) RunWithSeededOutput(ctx *node.ExecutionContext, graph models.WorkflowGraph, seedNodeType string, seedData map[string]any) *ExecutionResult {
	return e.run(ctx, graph, &seed{nodeType: seedNodeType, data: seedData})

}

func (e *Engine) run(ctx *node.ExecutionContext, graph models.WorkflowGraph, s *seed) *ExecutionResult {
	validation := ValidatorGraph(graph)
	if !validation.Valid {
		return &ExecutionResult{
			Error: fmt.Errorf("invalid workflow graph: %v", validation.Errors),
		}
	}

	nodesByID := make(map[string]models.GraphNode, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodesByID[n.ID] = n
	}

	outgoing := make(map[string][]models.GraphEdge)
	hasIncoming := make(map[string]bool, len(graph.Nodes))
	for _, edge := range graph.Edges {
		outgoing[edge.FromNodeID] = append(outgoing[edge.FromNodeID], edge)
		hasIncoming[edge.ToNodeID] = true
	}

	roots := []string{}
	for _, n := range graph.Nodes {
		if !hasIncoming[n.ID] {
			roots = append(roots, n.ID)
		}
	}

	outputs := make(map[string]map[string]any)
	visited := make(map[string]bool, len(graph.Nodes))
	queue := append([]string{}, roots...)

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if visited[nodeID] {
			continue
		}
		visited[nodeID] = true

		graphNode, ok := nodesByID[nodeID]
		if !ok {
			return &ExecutionResult{
				Outputs:      outputs,
				Error:        fmt.Errorf("node %q referenced but not defined in graph", nodeID),
				FailedNodeID: nodeID,
			}
		}

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

		for _, edge := range outgoing[nodeID] {
			if edgeIsActive(edge, result) && !visited[edge.ToNodeID] {
				queue = append(queue, edge.ToNodeID)
			}
		}
	}
	return &ExecutionResult{Outputs: outputs}
}

func edgeIsActive(edge models.GraphEdge, sourceOutput map[string]any) bool {
	if edge.Condition == "" {
		return true
	}
	result, ok := sourceOutput["result"]
	if !ok {
		return false
	}

	return fmt.Sprintf("%v", result) == edge.Condition
}
