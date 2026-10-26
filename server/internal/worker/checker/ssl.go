package checker

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"netra-uptime/pkg/utils"
	"strings"
	"time"
)

type SSLCheckResult struct {
	ExpiresAt *time.Time
	DaysLeft  *int
	Status    string // valid | warning | expired | error
	ErrorMsg  string
}

// ExtractHostPort parses a target URL or host string into hostname and port for SSL dialer
func ExtractHostPort(target string) (host string, port string, err error) {
	raw := strings.TrimSpace(target)
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("invalid target URL: %w", err)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return "", "", fmt.Errorf("could not parse hostname from target")
	}

	p := parsed.Port()
	if p == "" {
		p = "443"
	}

	return hostname, p, nil
}

// CheckSSLExpiry performs a TLS Handshake to extract certificate expiration details
func CheckSSLExpiry(target string, timeout time.Duration) *SSLCheckResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	hostname, port, err := ExtractHostPort(target)
	if err != nil {
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: "Invalid Target: URL format incorrect",
		}
	}

	// Check if target is an IP address (SSL domain check skip/warning)
	if net.ParseIP(hostname) != nil {
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: "Target is IP address (SSL Certificate check requires hostname)",
		}
	}

	// Validate target hostname against SSRF before dialing
	if err := utils.ValidateTargetHost(hostname); err != nil {
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: fmt.Sprintf("Restricted Target: %s", err.Error()),
		}
	}

	addr := net.JoinHostPort(hostname, port)
	rawConn, dialErr := utils.SafeDialTimeout("tcp", addr, timeout)
	if dialErr != nil {
		errMsg := mapSSLError(dialErr)
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: errMsg,
		}
	}
	defer func() {
		_ = rawConn.Close()
	}()

	tlsConfig := &tls.Config{
		ServerName: hostname,
		// #nosec G402 -- InsecureSkipVerify is required to inspect certificates even when expired or untrusted
		InsecureSkipVerify: true,
	}
	conn := tls.Client(rawConn, tlsConfig)
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if err := conn.Handshake(); err != nil {
		errMsg := mapSSLError(err)
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: errMsg,
		}
	}
	defer func() {
		_ = conn.Close()
	}()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return &SSLCheckResult{
			Status:   "error",
			ErrorMsg: "No SSL Certificate found on server",
		}
	}

	leafCert := state.PeerCertificates[0]
	expiry := leafCert.NotAfter
	daysLeft := int(time.Until(expiry).Hours() / 24)

	status := "valid"
	if daysLeft <= 0 {
		status = "expired"
	} else if daysLeft <= 30 {
		status = "warning"
	}

	return &SSLCheckResult{
		ExpiresAt: &expiry,
		DaysLeft:  &daysLeft,
		Status:    status,
		ErrorMsg:  "",
	}
}

func mapSSLError(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "restricted") || strings.Contains(errStr, "prohibited") || strings.Contains(errStr, "SSRF") {
		return fmt.Sprintf("Restricted Target: %s", errStr)
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "failed to resolve") {
		return "DNS Lookup Failed: Hostname could not be resolved"
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "i/o timeout") {
		return "Connection Timeout: Port 443 blocked or unreachable"
	}
	if strings.Contains(errStr, "connection refused") {
		return "Connection Refused: Port 443 closed"
	}
	if strings.Contains(errStr, "certificate has expired") {
		return "SSL Certificate Expired"
	}
	if strings.Contains(errStr, "unknown authority") || strings.Contains(errStr, "self signed") {
		return "Untrusted or Self-Signed SSL Certificate"
	}
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "Forbidden") {
		return "WAF / Firewall Blocked: Access denied by target server"
	}
	return fmt.Sprintf("SSL Handshake Failed: %s", errStr)
}
