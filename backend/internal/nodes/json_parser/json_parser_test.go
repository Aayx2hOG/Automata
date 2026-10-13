package jsonparser

import (
	"github.com/Aayx2hOG/automata/internal/node"
	"testing"
)

func TestParsesObjectsArraysAndScalars(t *testing.T) {
	for _, input := range []string{`{"customer":{"id":"123"}}`, `[1,2]`, `42`, `null`} {
		output, err := New().Execute(&node.ExecutionContext{Config: map[string]any{"source_node": "request", "source_field": "body"}, Outputs: map[string]map[string]any{"request": {"body": input}}})
		if err != nil {
			t.Fatalf("input=%s err=%v", input, err)
		}
		if _, ok := output["parsed"]; !ok {
			t.Fatal("missing parsed output")
		}
	}
}
