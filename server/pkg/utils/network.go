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

// IsRestrictedIP checks whether an IP address is private, loopback, or cloud metadata.
// Private network ranges (RFC 1918) are blocked unless ALLOW_PRIVATE_NETWORK_MONITORING=true.
// Loopback, unspecified, link-local, and cloud metadata (169.254.169.254) are ALWAYS blocked.
func IsRestrictedIP(ip net.IP) error {
	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")
	return isRestrictedIP(ip, allowPrivate)
}

// isRestrictedIP returns an error if an IP address is private, loopback, or metadata.
func isRestrictedIP(ip net.IP, allowPrivate bool) error {
	if ip == nil {
		return errors.New("nil IP address")
	}

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
// It fails closed: if DNS lookup fails or yields no IPs, an error is returned.
func ValidateTargetHost(target string) error {
	host, err := ExtractHostname(target)
	if err != nil {
		return err
	}

	hostLower := strings.ToLower(host)
	if blockedHostnames[hostLower] {
		return fmt.Errorf("access to internal hostname '%s' is prohibited", host)
	}

	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")

	// 1. Direct IP Check
	if directIP := net.ParseIP(host); directIP != nil {
		return isRestrictedIP(directIP, allowPrivate)
	}

	// 2. DNS Lookup Check (Fail-Closed)
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host '%s': %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for host '%s'", host)
	}
	for _, ip := range ips {
		if err := isRestrictedIP(ip, allowPrivate); err != nil {
			return err
		}
	}

	return nil
}

// ResolveSafeIPs resolves a target/host to a list of safe IP addresses,
// ensuring fail-closed resolution and filtering out any restricted IP addresses.
func ResolveSafeIPs(ctx context.Context, target string) ([]net.IP, error) {
	host, err := ExtractHostname(target)
	if err != nil {
		return nil, err
	}

	hostLower := strings.ToLower(host)
	if blockedHostnames[hostLower] {
		return nil, fmt.Errorf("access to internal hostname '%s' is prohibited (SSRF protection)", host)
	}

	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")

	// 1. Direct IP Check
	if directIP := net.ParseIP(host); directIP != nil {
		if err := isRestrictedIP(directIP, allowPrivate); err != nil {
			return nil, err
		}
		return []net.IP{directIP}, nil
	}

	// 2. DNS Lookup Check (Fail-Closed)
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host '%s': %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for host '%s'", host)
	}

	var safeIPs []net.IP
	for _, ip := range ips {
		if err := isRestrictedIP(ip, allowPrivate); err == nil {
			safeIPs = append(safeIPs, ip)
		}
	}

	if len(safeIPs) == 0 {
		return nil, fmt.Errorf("all resolved IPs for host '%s' are restricted (SSRF protection)", host)
	}

	return safeIPs, nil
}

// SafeDialContext dials the target address while enforcing SSRF protection at dial time,
// preventing DNS rebinding and connections to internal/restricted IP ranges.
func SafeDialContext(ctx context.Context, network, addr string, timeout time.Duration) (net.Conn, error) {
	targetAddr := addr
	if strings.Contains(targetAddr, "://") {
		if u, err := url.Parse(targetAddr); err == nil && u.Host != "" {
			targetAddr = u.Host
		}
	}

	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid address '%s': %w", addr, err)
	}

	hostLower := strings.ToLower(host)
	if blockedHostnames[hostLower] {
		return nil, fmt.Errorf("connection to internal host '%s' is prohibited (SSRF protection)", host)
	}

	allowPrivate := strings.EqualFold(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"), "true")
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	// Validate IP if it's already an IP literal
	if ip := net.ParseIP(host); ip != nil {
		if err := isRestrictedIP(ip, allowPrivate); err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	// Resolve host to IP
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host '%s': %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for host '%s'", host)
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
}

// SafeDialTimeout connects to addr on the named network with a timeout, validating
// that resolved addresses are not restricted (SSRF protection).
func SafeDialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return SafeDialContext(ctx, network, addr, timeout)
}

// NewSafeHTTPClient returns an *http.Client whose transport verifies resolved IPs
// during TCP connection time, preventing SSRF and DNS rebinding (TOCTOU).
// It also limits HTTP redirects to 5 and validates redirect destination hosts.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return SafeDialContext(ctx, network, addr, timeout)
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 5 redirects")
			}
			return ValidateTargetHost(req.URL.String())
		},
	}
}
