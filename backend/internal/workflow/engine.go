package workflow

import (
	"fmt"
	"maps"

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
	remaining := make(map[string]int, len(graph.Nodes))
	active := make(map[string]bool, len(graph.Nodes))
	for _, edge := range graph.Edges {
		outgoing[edge.FromNodeID] = append(outgoing[edge.FromNodeID], edge)
		remaining[edge.ToNodeID]++
	}

	ready := []string{}
	for _, n := range graph.Nodes {
		if remaining[n.ID] == 0 {
			ready = append(ready, n.ID)
			active[n.ID] = true
		}
	}
	outputs := make(map[string]map[string]any)
	for len(ready) > 0 {
		nodeID := ready[0]
		ready = ready[1:]
		if err := ctx.Ctx.Err(); err != nil {
			return &ExecutionResult{Outputs: outputs, Error: err, FailedNodeID: nodeID}
		}
		graphNode := nodesByID[nodeID]
		var result map[string]any
		if active[nodeID] {
			if s != nil && graphNode.Type == s.nodeType {
				// Persisted trigger data is the trigger's output, not its config.
				result = maps.Clone(s.data)
			} else {
				impl, err := e.registry.Build(graphNode.Type)
				if err != nil {
					return &ExecutionResult{Outputs: outputs, Error: fmt.Errorf("node %q: %w", nodeID, err), FailedNodeID: nodeID}
				}
				result, err = impl.Execute(&node.ExecutionContext{
					Ctx: ctx.Ctx, NodeID: nodeID, Config: graphNode.Config, Outputs: outputs,
				})
				if err != nil {
					return &ExecutionResult{Outputs: outputs, Error: fmt.Errorf("node %q failed: %w", nodeID, err), FailedNodeID: nodeID}
				}
			}
			outputs[nodeID] = result
		}
		// Every edge must resolve, including edges from skipped branches. A join
		// executes once all predecessors resolve and at least one edge is active.
		for _, edge := range outgoing[nodeID] {
			if active[nodeID] && edgeIsActive(edge, result) {
				active[edge.ToNodeID] = true
			}
			remaining[edge.ToNodeID]--
			if remaining[edge.ToNodeID] == 0 {
				ready = append(ready, edge.ToNodeID)
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
