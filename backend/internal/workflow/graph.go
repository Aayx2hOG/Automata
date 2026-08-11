package workflow

import "github.com/Aayx2hOG/automata/internal/models"

func BuildAdjacencyList(graph models.WorkflowGraph) map[string][]string {
	adjacency := make(map[string][]string)

	for _, node := range graph.Nodes {
		if _, exists := adjacency[node.ID]; !exists {
			adjacency[node.ID] = []string{}
		}
	}

	for _, edge := range graph.Edges {
		adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], edge.ToNodeID)
	}
	return adjacency
}
