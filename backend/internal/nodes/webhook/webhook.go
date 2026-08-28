package webhook

import "github.com/Aayx2hOG/automata/internal/node"

type Trigger struct{}

func New() *Trigger {
	return &Trigger{}
}

func (t *Trigger) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	return map[string]any{"triggered_via": "manual_fallback"}, nil
}
