package workflow

import (
	"fmt"

	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/nodes/condition"
	"github.com/Aayx2hOG/automata/internal/nodes/delay"
	httprequest "github.com/Aayx2hOG/automata/internal/nodes/http_request"
	jsonparser "github.com/Aayx2hOG/automata/internal/nodes/json_parser"
	"github.com/Aayx2hOG/automata/internal/nodes/logger"
	"github.com/Aayx2hOG/automata/internal/nodes/manual"
	"github.com/Aayx2hOG/automata/internal/nodes/webhook"
	"go.uber.org/zap"
)

type NodeConstructor func() node.Node

type Registry struct {
	constructors map[string]NodeConstructor
}

func NewRegistry(appLogger *zap.Logger) *Registry {
	r := &Registry{constructors: make(map[string]NodeConstructor)}

	r.constructors["manual_trigger"] = func() node.Node { return manual.New() }
	r.constructors["webhook_trigger"] = func() node.Node { return webhook.New() }
	r.constructors["logger"] = func() node.Node { return logger.New(appLogger) }
	r.constructors["delay"] = func() node.Node { return delay.New() }
	r.constructors["condition"] = func() node.Node { return condition.New() }
	r.constructors["http_request"] = func() node.Node { return httprequest.New() }
	r.constructors["json_parser"] = func() node.Node { return jsonparser.New() }

	return r
}

func (r *Registry) Build(nodeType string) (node.Node, error) {
	constructor, ok := r.constructors[nodeType]
	if !ok {
		return nil, fmt.Errorf("unknown node type: %q", nodeType)
	}
	return constructor(), nil
}
