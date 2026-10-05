package checker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

type DomainCheckResult struct {
	ExpiresAt *time.Time
	DaysLeft  *int
	Status    string // valid | warning | expired | error | none
	ErrorMsg  string
}

type rdapResponse struct {
	Events []struct {
		Action string    `json:"eventAction"`
		Date   time.Time `json:"eventDate"`
	} `json:"events"`
}

var sharedPlatformSuffixes = []string{
	".workers.dev",
	".vercel.app",
	".pages.dev",
	".github.io",
	".gitlab.io",
	".netlify.app",
	".herokuapp.com",
	".railway.app",
	".render.com",
	".onrender.com",
	".fly.dev",
	".glitch.me",
	".supabase.co",
	".ngrok-free.app",
	".ngrok.io",
	".loca.lt",
}

func isSharedPlatformSubdomain(hostname string) bool {
	h := strings.ToLower(strings.TrimSpace(hostname))
	for _, suffix := range sharedPlatformSuffixes {
		if strings.HasSuffix(h, suffix) && h != strings.TrimPrefix(suffix, ".") {
			return true
		}
	}
	return false
}

// CheckDomainExpiry checks domain expiration date via RDAP JSON API with WHOIS TCP fallback
func CheckDomainExpiry(target string, timeout time.Duration) *DomainCheckResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	hostname, _, err := ExtractHostPort(target)
	if err != nil {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "Invalid Target: URL format incorrect",
		}
	}

	// Subdomains on shared platforms (e.g. *.workers.dev, *.vercel.app) do not have independent domain registrations
	if isSharedPlatformSubdomain(hostname) {
		return &DomainCheckResult{
			Status:   "none",
			ErrorMsg: "Domain expiration not applicable for shared platform subdomains",
		}
	}

	// Remove www. or subdomains to get apex domain for WHOIS/RDAP lookup
	domain := getApexDomain(hostname)
	if domain == "" || net.ParseIP(domain) != nil {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "Target is IP address or local domain",
		}
	}

	// 1. Try RDAP API first (HTTP REST JSON)
	rdapRes := checkRDAP(domain, timeout)
	if rdapRes != nil {
		return rdapRes
	}

	// 2. Fallback to WHOIS Port 43 TCP Query
	whoisRes := checkWHOIS(domain, timeout)
	return whoisRes
}

func checkRDAP(domain string, timeout time.Duration) *DomainCheckResult {
	client := &http.Client{Timeout: timeout}
	url := fmt.Sprintf("https://rdap.org/domain/%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "PulseWatch-UptimeBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "Registry Rate Limit: RDAP rate limit reached, re-checking later",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil // Fallback to WHOIS
	}

	var data rdapResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	for _, event := range data.Events {
		if strings.EqualFold(event.Action, "expiration") || strings.EqualFold(event.Action, "registration expiration") {
			expiry := event.Date
			daysLeft := int(time.Until(expiry).Hours() / 24)

			status := "valid"
			if daysLeft <= 0 {
				status = "expired"
			} else if daysLeft <= 30 {
				status = "warning"
			}

			return &DomainCheckResult{
				ExpiresAt: &expiry,
				DaysLeft:  &daysLeft,
				Status:    status,
				ErrorMsg:  "",
			}
		}
	}

	return nil
}

func checkWHOIS(domain string, timeout time.Duration) *DomainCheckResult {
	whoisServer := "whois.iana.org:43"
	if strings.HasSuffix(domain, ".com") || strings.HasSuffix(domain, ".net") {
		whoisServer = "whois.verisign-grs.com:43"
	} else if strings.HasSuffix(domain, ".org") {
		whoisServer = "whois.pir.org:43"
	} else if strings.HasSuffix(domain, ".id") {
		whoisServer = "whois.id:43"
	}

	conn, err := net.DialTimeout("tcp", whoisServer, timeout)
	if err != nil {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "WHOIS Server Connection Timeout",
		}
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	_, _ = fmt.Fprintf(conn, "%s\r\n", domain)

	scanner := bufio.NewScanner(conn)
	var rawOutput strings.Builder
	for scanner.Scan() {
		rawOutput.WriteString(scanner.Text() + "\n")
	}

	output := rawOutput.String()
	if strings.Contains(output, "LIMIT EXCEEDED") || strings.Contains(output, "Too Many Connections") {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "Registry Rate Limit: WHOIS rate limit exceeded",
		}
	}

	expiryDate := parseWHOISExpiry(output)
	if expiryDate.IsZero() {
		return &DomainCheckResult{
			Status:   "error",
			ErrorMsg: "Registrar Privacy Protected or Expiry Date format unrecognized",
		}
	}

	daysLeft := int(time.Until(expiryDate).Hours() / 24)
	status := "valid"
	if daysLeft <= 0 {
		status = "expired"
	} else if daysLeft <= 30 {
		status = "warning"
	}

	return &DomainCheckResult{
		ExpiresAt: &expiryDate,
		DaysLeft:  &daysLeft,
		Status:    status,
		ErrorMsg:  "",
	}
}

var expiryRegexes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)Registry Expiry Date:\s*(.+)`),
	regexp.MustCompile(`(?i)Expiration Date:\s*(.+)`),
	regexp.MustCompile(`(?i)paid-till:\s*(.+)`),
	regexp.MustCompile(`(?i)Expiry Date:\s*(.+)`),
	regexp.MustCompile(`(?i)expires:\s*(.+)`),
}

func parseWHOISExpiry(whoisOutput string) time.Time {
	for _, re := range expiryRegexes {
		matches := re.FindStringSubmatch(whoisOutput)
		if len(matches) > 1 {
			dateStr := strings.TrimSpace(matches[1])
			formats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05Z",
				"2006-01-02 15:04:05",
				"2006-01-02",
				"02-Jan-2006",
				"2006.01.02",
			}
			for _, fmtStr := range formats {
				if t, err := time.Parse(fmtStr, dateStr); err == nil {
					return t
				}
			}
		}
	}
	return time.Time{}
}

func getApexDomain(hostname string) string {
	h := strings.ToLower(strings.TrimSpace(hostname))
	if apex, err := publicsuffix.EffectiveTLDPlusOne(h); err == nil && apex != "" {
		return apex
	}

	parts := strings.Split(h, ".")
	if len(parts) < 2 {
		return h
	}
	if len(parts) >= 3 {
		secondToLast := parts[len(parts)-2]
		knownSLDs := map[string]bool{
			"co": true, "com": true, "org": true, "go": true, "gov": true,
			"edu": true, "ac": true, "sch": true, "mil": true, "net": true,
			"web": true, "my": true, "biz": true, "or": true, "desa": true,
			"ponpes": true,
		}
		if knownSLDs[secondToLast] {
			return strings.Join(parts[len(parts)-3:], ".")
		}
	}
	return strings.Join(parts[len(parts)-2:], ".")
}
