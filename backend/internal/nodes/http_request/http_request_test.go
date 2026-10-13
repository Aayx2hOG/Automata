package httprequest

import (
	"context"
	"errors"
	"github.com/Aayx2hOG/automata/internal/credentials"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"
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

func TestHTTPFailurePolicies(t *testing.T) {
	for _, status := range []int{200, 302, 400, 401, 408, 429, 500, 503} {
		for _, policy := range []string{"output", "fail"} {
			n := New()
			n.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{"3"}}, Body: io.NopCloser(strings.NewReader("private response"))}, nil
			})
			output, err := n.Execute(&node.ExecutionContext{Ctx: context.Background(), Config: map[string]any{"url": "https://example.com", "error_policy": policy}})
			if policy == "output" || status == 200 {
				if err != nil || output["status_code"] != status {
					t.Fatalf("status=%d policy=%s output=%v err=%v", status, policy, output, err)
				}
				continue
			}
			var classified *node.RetryError
			if !errors.As(err, &classified) || classified.Retryable != (status == 408 || status == 429 || status >= 500) || classified.After != 3*time.Second {
				t.Fatalf("status=%d error=%v", status, err)
			}
			if strings.Contains(err.Error(), "private") {
				t.Fatal("response leaked into error")
			}
		}
	}
}

func TestCredentialHeaderAndOutputRedaction(t *testing.T) {
	n := New()
	n.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer top-secret" {
			t.Fatal("missing stored credential")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"auth":"Bearer top-secret","token":"top-secret"}`))}, nil
	})
	ctx := credentials.WithResolver(context.Background(), func(context.Context, uuid.UUID, *url.URL) (string, string, error) {
		return "Authorization", "Bearer top-secret", nil
	})
	output, err := n.Execute(&node.ExecutionContext{Ctx: ctx, Config: map[string]any{"url": "https://example.com", "credential_id": uuid.NewString()}})
	if err != nil || strings.Contains(output["body"].(string), "top-secret") {
		t.Fatalf("output redaction failed: %v", err)
	}
}

func TestCredentialRedirectIsNotFollowed(t *testing.T) {
	n := New()
	calls := 0
	n.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://other.example.com"}}, Body: io.NopCloser(strings.NewReader("redirect")), Request: r}, nil
	})
	ctx := credentials.WithResolver(context.Background(), func(context.Context, uuid.UUID, *url.URL) (string, string, error) {
		return "Authorization", "secret", nil
	})
	_, err := n.Execute(&node.ExecutionContext{Ctx: ctx, Config: map[string]any{"url": "https://example.com", "credential_id": uuid.NewString()}})
	if err != nil || calls != 1 {
		t.Fatalf("redirect followed: calls=%d err=%v", calls, err)
	}
}

func TestRetryAfterDate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	if got := retryAfter(now.Add(5*time.Second).Format(http.TimeFormat), now); got != 5*time.Second {
		t.Fatal(got)
	}
	for _, value := range []string{"invalid", "-5", "0"} {
		if retryAfter(value, now) != 0 {
			t.Fatalf("accepted %q", value)
		}
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

func TestDurableIdempotencyKeyOverridesConfiguredHeader(t *testing.T) {
	n := New()
	calls := 0
	n.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if got := r.Header.Get("Idempotency-Key"); got != "automata-stable-run-node-key" {
			t.Fatalf("idempotency key = %q", got)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})
	ec := &node.ExecutionContext{Ctx: context.Background(), IdempotencyKey: "automata-stable-run-node-key", Config: map[string]any{
		"url": "https://example.com", "method": "POST", "headers": map[string]any{"Idempotency-Key": "unsafe-shared-value"},
	}}
	for i := 0; i < 2; i++ {
		if _, err := n.Execute(ec); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}
