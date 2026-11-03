package consul

import (
	"fmt"
	"net"
	"strings"
)

// registerAddress returns an IPv4/IPv6 address Kong and the Consul agent can dial.
// configured may be empty (pick this host's address on the overlay network),
// a literal IP, or a hostname resolvable from this container.
func registerAddress(configured string, dialTarget string) (string, error) {
	configured = strings.TrimSpace(configured)
	if configured != "" {
		if ip := net.ParseIP(configured); ip != nil {
			return ip.String(), nil
		}
		ips, err := net.LookupIP(configured)
		if err != nil {
			return "", fmt.Errorf("lookup %q: %w", configured, err)
		}
		if len(ips) == 0 {
			return "", fmt.Errorf("lookup %q: no addresses", configured)
		}
		return ips[0].String(), nil
	}

	target := strings.TrimSpace(dialTarget)
	if target == "" {
		target = "consul:8500"
	}
	conn, err := net.Dial("udp", target)
	if err != nil {
		return "", fmt.Errorf("detect local address via %s: %w", target, err)
	}
	defer conn.Close()

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}
	return host, nil
}
