package httprequest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/Aayx2hOG/automata/internal/node"
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

	if err := validateURL(req.URL); err != nil {
		return nil, err
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

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if len(bodyBytes) > maxResponseBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}
	return map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(bodyBytes),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}, nil
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
