package checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"netra-uptime/pkg/utils"
	"strings"
	"syscall"
	"time"
)

// CheckPing verifies that a host is reachable. It resolves the target and sends an ICMP echo
// request. When ICMP sockets are unavailable (no CAP_NET_RAW / ping_group_range), it falls back to
// a TCP connection attempt on common ports, so a host that only resolves in DNS is never reported UP.
func CheckPing(target string, timeout time.Duration) (duration time.Duration, err error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	host := strings.TrimSpace(target)
	if h, _, splitErr := net.SplitHostPort(host); splitErr == nil {
		host = h
	}

	if err := utils.ValidateTargetHost(host); err != nil {
		return 0, err
	}

	start := time.Now()
	ips, err := utils.ResolveSafeIPs(context.Background(), host)
	if err != nil {
		return time.Since(start), err
	}
	if len(ips) == 0 {
		return time.Since(start), fmt.Errorf("no IP addresses found for %s", host)
	}

	var ip4 net.IP
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			ip4 = v4
			break
		}
	}
	if ip4 == nil {
		return tcpReachable(ips[0], timeout-time.Since(start))
	}

	rtt, icmpErr := icmpEcho(ip4, timeout-time.Since(start))
	if icmpErr == nil {
		return rtt, nil
	}
	if errors.Is(icmpErr, errICMPUnavailable) {
		return tcpReachable(ip4, timeout-time.Since(start))
	}
	return rtt, icmpErr
}

var errICMPUnavailable = errors.New("icmp sockets unavailable")

// tcpReachable is the fallback probe used when ICMP is not permitted. A completed handshake or an
// explicit connection refusal both prove the host is alive; only timeouts/unreachable mean DOWN.
func tcpReachable(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if timeout <= 0 {
		return 0, errors.New("ping timed out")
	}

	if err := utils.IsRestrictedIP(ip); err != nil {
		return 0, err
	}

	start := time.Now()
	var lastErr error
	for _, port := range []string{"443", "80"} {
		remaining := timeout - time.Since(start)
		if remaining <= 0 {
			break
		}
		conn, err := utils.SafeDialTimeout("tcp", net.JoinHostPort(ip.String(), port), remaining)
		if err == nil {
			_ = conn.Close()
			return time.Since(start), nil
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			return time.Since(start), nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("ping timed out")
	}
	return time.Since(start), fmt.Errorf("host unreachable: %w", lastErr)
}
