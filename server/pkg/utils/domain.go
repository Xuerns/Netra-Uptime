package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ExtractRootDomain extracts the effective TLD+1 (eTLD+1) or IP address from a target URL/host string
func ExtractRootDomain(rawTarget string) (string, error) {
	trimmed := strings.TrimSpace(rawTarget)
	if trimmed == "" {
		return "", fmt.Errorf("empty target")
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}

	u, err := url.Parse(trimmed)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("invalid target URL format")
	}

	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))

	// If hostname is an IP literal, return host IP
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}

	// Extract eTLD+1 (e.g. sub.api.example.com -> example.com)
	etldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		// Fallback for internal / local / unqualified hostnames
		return host, nil
	}

	return strings.ToLower(etldPlusOne), nil
}
