package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var blockedHostnames = map[string]bool{
	"localhost":                true,
	"postgres":                 true,
	"redis":                    true,
	"metadata.google.internal": true,
	"metadata":                 true,
}

// ExtractHostname parses a URL, host:port, or hostname/IP string into just the hostname.
func ExtractHostname(rawTarget string) (string, error) {
	trimmed := strings.TrimSpace(rawTarget)
	if trimmed == "" {
		return "", errors.New("target cannot be empty")
	}

	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("invalid target URL format: %w", err)
		}
		host := u.Hostname()
		if host == "" {
			return "", errors.New("target URL missing hostname")
		}
		return strings.ToLower(host), nil
	}

	// Try host:port
	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		return strings.ToLower(host), nil
	}

	// Pure hostname or IP
	return strings.ToLower(trimmed), nil
}

// isRestrictedIP returns an error if an IP address is private, loopback, or metadata.
func isRestrictedIP(ip net.IP, allowPrivate bool) error {
	// Cloud metadata (169.254.169.254) is ALWAYS blocked
	if ip.String() == "169.254.169.254" {
		return errors.New("access to cloud metadata service (169.254.169.254) is prohibited (SSRF protection)")
	}

	// Loopback (127.0.0.0/8, ::1) is ALWAYS blocked
	if ip.IsLoopback() {
		return fmt.Errorf("target resolves to loopback address '%s' (SSRF protection)", ip.String())
	}

	// Unspecified (0.0.0.0, ::) is ALWAYS blocked
	if ip.IsUnspecified() {
		return fmt.Errorf("target resolves to unspecified address '%s' (SSRF protection)", ip.String())
	}

	// Link-local unicast / multicast is ALWAYS blocked
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("target resolves to link-local address '%s' (SSRF protection)", ip.String())
	}

	// Private IP ranges (RFC 1918: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
	if ip.IsPrivate() && !allowPrivate {
		return fmt.Errorf("target resolves to private network address '%s' (SSRF protection)", ip.String())
	}

	return nil
}

// ValidateTargetHost checks that the target does not resolve to restricted private, loopback, or metadata addresses.
func ValidateTargetHost(target string) error {
	host, err := ExtractHostname(target)
	if err != nil {
		return err
	}

	if blockedHostnames[host] {
		return fmt.Errorf("access to internal hostname '%s' is prohibited", host)
	}

	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")

	// 1. Direct IP Check
	if directIP := net.ParseIP(host); directIP != nil {
		return isRestrictedIP(directIP, allowPrivate)
	}

	// 2. DNS Lookup Check
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			if err := isRestrictedIP(ip, allowPrivate); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewSafeHTTPClient returns an *http.Client whose transport verifies resolved IPs
// during TCP connection time, preventing SSRF and DNS rebinding (TOCTOU).
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			if blockedHostnames[strings.ToLower(host)] {
				return nil, fmt.Errorf("connection to internal host '%s' is prohibited", host)
			}

			// Validate IP if it's already an IP literal
			if ip := net.ParseIP(host); ip != nil {
				if err := isRestrictedIP(ip, allowPrivate); err != nil {
					return nil, err
				}
				return dialer.DialContext(ctx, network, addr)
			}

			// Resolve host to IP
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			var chosenIP net.IP
			for _, ip := range ips {
				if err := isRestrictedIP(ip, allowPrivate); err == nil {
					chosenIP = ip
					break
				}
			}

			if chosenIP == nil {
				return nil, fmt.Errorf("all resolved IPs for host '%s' are restricted (SSRF protection)", host)
			}

			return dialer.DialContext(ctx, network, net.JoinHostPort(chosenIP.String(), port))
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
