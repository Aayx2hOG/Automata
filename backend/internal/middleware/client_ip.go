package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func trustedAddress(ip netip.Addr, proxies []netip.Prefix) bool {
	for _, proxy := range proxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}

// Walk from the directly connected peer towards the client, stopping at the
// first untrusted hop. Client-supplied prefixes cannot change that identity.
func clientIP(r *http.Request, proxies []netip.Prefix) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	peer = peer.Unmap()
	if !trustedAddress(peer, proxies) {
		return peer.String()
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		hops := strings.Split(forwarded, ",")
		current := peer
		for i := len(hops) - 1; i >= 0; i-- {
			if !trustedAddress(current, proxies) {
				return current.String()
			}
			next, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
			if err != nil {
				return peer.String()
			}
			current = next.Unmap()
		}
		return current.String()
	}
	if real, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("X-Real-IP"))); err == nil {
		return real.Unmap().String()
	}
	return peer.String()
}
