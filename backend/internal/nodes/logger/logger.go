package logger

import (
	"fmt"

	"github.com/Aayx2hOG/automata/internal/node"
	"go.uber.org/zap"
)

type Node struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Node {
	return &Node{logger: logger}
}

func (n *Node) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	message, _ := ec.Config["message"].(string)
	if message == "" {
		message = "(no message configured)"
	}

	n.logger.Info("workflow logger node", zap.String("node_id", ec.NodeID), zap.String("message", message))
	return map[string]any{"logged_message": fmt.Sprintf("%v", message)}, nil
}
