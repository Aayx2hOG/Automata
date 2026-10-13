package workflow

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/Aayx2hOG/automata/internal/models"
)

func resolveConfig(definition models.GraphNode, outputs map[string]map[string]any) (map[string]any, error) {
	config := maps.Clone(definition.Config)
	if config == nil {
		config = make(map[string]any)
	}
	for field, ref := range definition.Inputs {
		var value any
		value, ok := outputs[ref.NodeID]
		if !ok {
			return nil, fmt.Errorf("input %q: source node %q has no output", field, ref.NodeID)
		}
		for _, segment := range ref.Path {
			switch current := value.(type) {
			case map[string]any:
				value, ok = current[segment]
			case []any:
				index, err := strconv.Atoi(segment)
				ok = err == nil && index >= 0 && index < len(current)
				if ok {
					value = current[index]
				}
			default:
				ok = false
			}
			if !ok {
				return nil, fmt.Errorf("input %q: field %q is missing in node %q", field, segment, ref.NodeID)
			}
		}
		config[field] = value
	}
	return config, nil
}

func validateInputs(graph models.WorkflowGraph) []ValidationError {
	parents := make(map[string][]string)
	for _, edge := range graph.Edges {
		parents[edge.ToNodeID] = append(parents[edge.ToNodeID], edge.FromNodeID)
	}
	var errors []ValidationError
	for _, n := range graph.Nodes {
		ancestors := make(map[string]bool)
		var visit func(string)
		visit = func(id string) {
			for _, parent := range parents[id] {
				if !ancestors[parent] {
					ancestors[parent] = true
					visit(parent)
				}
			}
		}
		visit(n.ID)
		for field, ref := range n.Inputs {
			if field == "" || field == "credential_id" || field == "error_policy" || !ancestors[ref.NodeID] || ref.NodeID == n.ID || len(ref.Path) == 0 {
				errors = append(errors, ValidationError{NodeID: n.ID, Message: fmt.Sprintf("input %q must reference an upstream field; credentials and error policy cannot be mapped", field)})
			}
		}
	}
	return errors
}
