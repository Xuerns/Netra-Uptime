//go:build !windows

package checker

import (
	"errors"
	"net"
	"netra-uptime/pkg/utils"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func icmpEcho(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if timeout <= 0 {
		return 0, errors.New("ping timed out")
	}

	if err := utils.IsRestrictedIP(ip); err != nil {
		return 0, err
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
