package workflow

import (
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
)

func topologicalOrder(graph models.WorkflowGraph) ([]string, error) {
	adjacency := BuildAdjacencyList(graph)

	inDegree := make(map[string]int, len(graph.Nodes))
	for _, n := range graph.Nodes {
		inDegree[n.ID] = 0
	}
	for _, targets := range adjacency {
		for _, target := range targets {
			inDegree[target]++
		}
	}

	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	order := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, next := range adjacency[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(graph.Nodes) {
		return nil, fmt.Errorf("unable to compute execution order: graph may contain a cycle")
	}
	return order, nil
}
