package httprequest

import (
	"context"
	"github.com/Aayx2hOG/automata/internal/node"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestBlockedDestinations(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "100.100.100.200", "::1", "::ffff:127.0.0.1", "fc00::1", "2002:7f00:1::", "0.0.0.0"} {
		if publicAddress(netip.MustParseAddr(ip)) {
			t.Errorf("allowed %s", ip)
		}
	}
	for _, url := range []string{"file:///etc/passwd", "http://127.0.0.1", "http://[::1]", "http://user:pass@example.com"} {
		if _, err := New().Execute(&node.ExecutionContext{Ctx: context.Background(), Config: map[string]any{"url": url}}); err == nil {
			t.Errorf("allowed %s", url)
		}
	}
	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	if err := New().client.CheckRedirect(req, nil); err == nil {
		t.Fatal("allowed internal redirect")
	}
	if c, err := safeDial(context.Background(), "tcp", "127.0.0.1:80"); err == nil {
		c.Close()
		t.Fatal("dialed loopback")
	}
}
func TestResponseLimit(t *testing.T) {
	for _, size := range []int{maxResponseBytes, maxResponseBytes + 1} {
		n := New()
		n.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat("a", size)))}, nil
		})
		_, err := n.Execute(&node.ExecutionContext{Ctx: context.Background(), Config: map[string]any{"url": "https://example.com"}})
		if (err != nil) != (size > maxResponseBytes) {
			t.Fatalf("size %d: %v", size, err)
		}
	}
}
