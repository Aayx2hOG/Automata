package httprequest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Aayx2hOG/automata/internal/credentials"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/google/uuid"
)

type Node struct {
	client *http.Client
}

func New() *Node {
	return &Node{
		client: &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{DialContext: safeDial, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 15 * time.Second, MaxResponseHeaderBytes: 1 << 20}, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return validateURL(req.URL)
		}},
	}
}

func (n *Node) Execute(ec *node.ExecutionContext) (map[string]any, error) {
	url, _ := ec.Config["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("http_request requires a url")
	}

	method, _ := ec.Config["method"].(string)
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	jsonBody := false
	if bodyStr, ok := ec.Config["body"].(string); ok && bodyStr != "" {
		bodyReader = strings.NewReader(bodyStr)
	} else if body := ec.Config["body"]; body != nil && !ok {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
		jsonBody = true
	}

	req, err := http.NewRequestWithContext(ec.Ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %v", err)
	}

	if err := validateURL(req.URL); err != nil {
		return nil, err
	}
	if jsonBody {
		req.Header.Set("Content-Type", "application/json")
	}

	if headers, ok := ec.Config["headers"].(map[string]any); ok {
		for key, val := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", val))
		}
	}
	// A durable node keeps the same key across retries and worker recovery.
	// The destination must implement idempotency for this to prevent duplicates.
	if ec.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", ec.IdempotencyKey)
	}
	secret := ""
	client := *n.client
	if raw, exists := ec.Config["credential_id"]; exists && raw != "" {
		idString, _ := raw.(string)
		id, err := uuid.Parse(idString)
		resolve := credentials.FromContext(ec.Ctx)
		if err != nil || resolve == nil {
			return nil, &node.RetryError{Message: "credential is unavailable or invalid"}
		}
		header, value, err := resolve(ec.Ctx, id, req.URL)
		if err != nil {
			return nil, &node.RetryError{Message: err.Error()}
		}
		secret = value
		req.Header.Set(header, value)
		// Do not forward a stored credential through redirects, including same-host redirects.
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %s", redact(err.Error(), secret))
	}
	defer resp.Body.Close()
	if policy, _ := ec.Config["error_policy"].(string); policy == "fail" && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		retryable := resp.StatusCode == 408 || resp.StatusCode == 429 || resp.StatusCode >= 500
		return nil, &node.RetryError{Message: fmt.Sprintf("HTTP request returned status %d", resp.StatusCode), Retryable: retryable, After: retryAfter(resp.Header.Get("Retry-After"), time.Now())}
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response body: %s", redact(err.Error(), secret))
	}

	if len(bodyBytes) > maxResponseBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}
	return map[string]any{
		"status_code": resp.StatusCode,
		"body":        redact(string(bodyBytes), secret),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}, nil
}

func redact(value, secret string) string {
	if secret == "" {
		return value
	}
	secrets := []string{secret}
	if strings.HasPrefix(secret, "Bearer ") || strings.HasPrefix(secret, "Basic ") {
		token := strings.SplitN(secret, " ", 2)[1]
		if token != "" {
			secrets = append(secrets, token)
		}
	}
	for _, token := range secrets {
		encoded, _ := json.Marshal(token)
		value = strings.ReplaceAll(value, string(encoded[1:len(encoded)-1]), "[REDACTED]")
		value = strings.ReplaceAll(value, token, "[REDACTED]")
	}
	return value
}

func retryAfter(value string, now time.Time) time.Duration {
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
		// A larger wait cannot fit within the execution deadline. Avoid overflow.
		if seconds > 86400 {
			seconds = 86400
		}
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil && date.After(now) {
		return date.Sub(now)
	}
	return 0
}

const maxResponseBytes = 4 << 20

func validateURL(u *url.URL) error {
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return fmt.Errorf("destination must be an HTTP(S) URL without credentials")
	}
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicAddress(ip) {
		return fmt.Errorf("destination address is not public")
	}
	return nil
}

func publicAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, block := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001::/23", "2001:db8::/32", "2002::/16", "64:ff9b::/96", "64:ff9b:1::/48"} {
		if netip.MustParsePrefix(block).Contains(ip) {
			return false
		}
	}
	return ip.Is4() || netip.MustParsePrefix("2000::/3").Contains(ip)
}

func safeDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("destination has no addresses")
	}
	for _, ip := range ips {
		if !publicAddress(ip) {
			return nil, fmt.Errorf("destination resolves to a non-public address")
		}
	}
	var last error
	for _, ip := range ips {
		conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		last = err
	}
	return nil, last
}
