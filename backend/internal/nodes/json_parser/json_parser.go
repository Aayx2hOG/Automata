package jsonparser

import (
	"encoding/json"
	"fmt"

	"github.com/Aayx2hOG/automata/internal/node"
)

type Node struct{}

func New() *Node {
	return &Node{}
}

func (n *Node) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	sourceNode, _ := ec.Config["source_node"].(string)
	sourceField, _ := ec.Config["source_field"].(string)

	if sourceNode == "" || sourceField == "" {
		return nil, fmt.Errorf("json_parse node requires source_node and source_field")
	}

	raw, found := ec.Output(sourceNode, sourceField)
	if !found {
		return nil, fmt.Errorf("json_parser: no output found at %s.%s", sourceNode, sourceField)
	}

	rawStr, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("json_parser: %s.%s is not a string", sourceNode, sourceField)
	}

	parsed := []any{}
	if err := json.Unmarshal([]byte(rawStr), &parsed); err != nil {
		return nil, fmt.Errorf("json_parser: invalid JSON in %s.%s: %w", sourceNode, sourceField, err)
	}

	return map[string]any{"parsed": parsed}, nil
}
