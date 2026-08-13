package node

import "context"

type ExecutionContext struct {
	Ctx     context.Context
	NodeID  string
	Config  map[string]any
	Outputs map[string]map[string]any
}

func (ec *ExecutionContext) Output(nodeID, field string) (any, bool) {
	nodeOutputs, ok := ec.Outputs[nodeID]
	if !ok {
		return nil, false
	}
	val, ok := nodeOutputs[field]
	return val, ok
}

type Node interface {
	Execute(ec *ExecutionContext) (map[string]any, error)
}
