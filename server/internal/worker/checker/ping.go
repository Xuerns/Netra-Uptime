package checker

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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

	start := time.Now()
	ips, err := net.LookupIP(host)
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

func icmpEcho(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if timeout <= 0 {
		return 0, errors.New("ping timed out")
	}

	var (
		conn       *icmp.PacketConn
		privileged = true
	)
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		privileged = false
		conn, err = icmp.ListenPacket("udp4", "0.0.0.0")
		if err != nil {
			return 0, errICMPUnavailable
		}
	}
	defer func() { _ = conn.Close() }()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("netra-uptime")},
	}
	payload, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	var dst net.Addr = &net.UDPAddr{IP: ip}
	if privileged {
		dst = &net.IPAddr{IP: ip}
	}

	start := time.Now()
	deadline := start.Add(timeout)
	_ = conn.SetDeadline(deadline)
	if _, err := conn.WriteTo(payload, dst); err != nil {
		return time.Since(start), err
	}

	buf := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				return time.Since(start), errors.New("ping timed out: no ICMP echo reply received")
			}
			return time.Since(start), err
		}

		var peerIP net.IP
		switch a := peer.(type) {
		case *net.IPAddr:
			peerIP = a.IP
		case *net.UDPAddr:
			peerIP = a.IP
		}
		if peerIP != nil && !peerIP.Equal(ip) {
			continue
		}

		reply, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}
		switch reply.Type {
		case ipv4.ICMPTypeEchoReply:
			return time.Since(start), nil
		case ipv4.ICMPTypeDestinationUnreachable:
			return time.Since(start), errors.New("destination host unreachable")
		}
		if time.Now().After(deadline) {
			return time.Since(start), errors.New("ping timed out: no ICMP echo reply received")
		}
	}
}

// tcpReachable is the fallback probe used when ICMP is not permitted. A completed handshake or an
// explicit connection refusal both prove the host is alive; only timeouts/unreachable mean DOWN.
func tcpReachable(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if timeout <= 0 {
		return 0, errors.New("ping timed out")
	}

	start := time.Now()
	var lastErr error
	for _, port := range []string{"443", "80"} {
		remaining := timeout - time.Since(start)
		if remaining <= 0 {
			break
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port), remaining)
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
