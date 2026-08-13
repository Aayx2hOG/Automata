package delay

import (
	"time"

	"github.com/Aayx2hOG/automata/internal/node"
)

type Node struct{}

func New() *Node {
	return &Node{}
}

func (n *Node) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	seconds, _ := ec.Config["seconds"].(float64)
	if seconds <= 0 {
		seconds = 1
	}

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
		return map[string]any{"delayed_seconds": seconds}, nil
	case <-ec.Ctx.Done():
		return nil, ec.Ctx.Err()
	}
}
