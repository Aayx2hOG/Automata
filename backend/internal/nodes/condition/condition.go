package condition

import (
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
	operator, _ := ec.Config["operator"].(string)

	if sourceNode == "" || sourceField == "" {
		return nil, fmt.Errorf("condition node requires source_code and source_field")
	}

	actual, found := ec.Output(sourceNode, sourceField)

	var result bool
	switch operator {
	case "exists":
		result = found
	case "not_equal":
		result = found && fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", ec.Config["value"])
	case "equals", "":
		result = found && fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", ec.Config["value"])
	default:
		return nil, fmt.Errorf("condition node: unknown operator %q", operator)
	}
	return map[string]any{"result": result}, nil
}
