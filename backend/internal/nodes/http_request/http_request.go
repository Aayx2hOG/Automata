package httprequest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Aayx2hOG/automata/internal/node"
)

type Node struct {
	client *http.Client
}

func New() *Node {
	return &Node{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (n *Node) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	url, _ := ec.Config["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("http_request not requires a url")
	}

	method, _ := ec.Config["method"].(string)
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if bodyStr, ok := ec.Config["body"].(string); ok && bodyStr != "" {
		bodyReader = strings.NewReader(bodyStr)
	}

	req, err := http.NewRequestWithContext(ec.Ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %v", err)
	}

	if headers, ok := ec.Config["headers"].(map[string]any); ok {
		for key, val := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", val))
		}
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(bodyBytes),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}, nil
}
