package workflow

import (
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
)

type ValidationError struct {
	NodeID  string `json:"node_id,omitempty"`
	Message string `json:"message"`
}

type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

type visitState int

const MaxNodes = 100
const MaxEdges = 500

const (
	white visitState = iota
	gray
	black
)

func ValidatorGraph(graph models.WorkflowGraph) ValidationResult {
	errs := []ValidationError{}
	if len(graph.Nodes) > MaxNodes || len(graph.Edges) > MaxEdges {
		return ValidationResult{Errors: []ValidationError{{Message: fmt.Sprintf("workflow exceeds limits of %d nodes or %d edges", MaxNodes, MaxEdges)}}}
	}

	if len(graph.Nodes) == 0 {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Message: "workflow must contain atleast one node."}},
		}
	}

	nodeIDs := make(map[string]bool, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if node.Type == "http_request" {
			if policy, exists := node.Config["error_policy"]; exists && policy != "output" && policy != "fail" {
				errs = append(errs, ValidationError{NodeID: node.ID, Message: "HTTP error_policy must be output or fail"})
			}
		}
		if len(node.ID) > 255 {
			errs = append(errs, ValidationError{Message: "node id must not exceed 255 bytes"})
		}
		if node.Retry != nil && (node.Retry.MaxAttempts < 1 || node.Retry.MaxAttempts > 5 || node.Retry.BackoffSeconds < 0 || node.Retry.BackoffSeconds > 30) {
			errs = append(errs, ValidationError{NodeID: node.ID, Message: "retry requires 1–5 attempts and 0–30 seconds backoff"})
		}
		if node.ID == "" {
			errs = append(errs, ValidationError{Message: "node found with an empty id"})
			continue
		}
		if nodeIDs[node.ID] {
			errs = append(errs, ValidationError{NodeID: node.ID, Message: "duplicate node id"})
			continue
		}
		nodeIDs[node.ID] = true
	}

	for _, edge := range graph.Edges {
		if !nodeIDs[edge.FromNodeID] {
			errs = append(errs, ValidationError{
				NodeID:  edge.FromNodeID,
				Message: fmt.Sprintf("edge references unknown source node %q", edge.FromNodeID),
			})
		}
		if !nodeIDs[edge.ToNodeID] {
			errs = append(errs, ValidationError{
				NodeID:  edge.FromNodeID,
				Message: fmt.Sprintf("edge references unknown source node %q", edge.FromNodeID),
			})
		}
	}

	if len(errs) == 0 {
		if cycleNodeID, found := detectCycle(BuildAdjacencyList(graph)); found {
			errs = append(errs, ValidationError{
				NodeID:  cycleNodeID,
				Message: "workflow graph contains a cycle",
			})
		}
	}
	if len(errs) == 0 {
		errs = append(errs, validateInputs(graph)...)
	}
	return ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

func detectCycle(adjacency map[string][]string) (string, bool) {
	state := make(map[string]visitState, len(adjacency))
	for id := range adjacency {
		state[id] = white
	}

	var visit func(nodeID string) (string, bool)
	visit = func(nodeID string) (string, bool) {
		state[nodeID] = gray
		for _, next := range adjacency[nodeID] {
			switch state[next] {
			case gray:
				return next, true
			case white:
				if cycleID, found := visit(next); found {
					return cycleID, found
				}
			}
		}
		state[nodeID] = black
		return "", false
	}

	for id, s := range state {
		if s == white {
			if cycleID, found := visit(id); found {
				return cycleID, true
			}
		}
	}
	return "", false
}
