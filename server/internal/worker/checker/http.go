package checker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"netra-uptime/pkg/utils"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxResponseSnippetBytes = 512

type HTTPCheckOptions struct {
	Method         string
	HeadersJSON    string
	Body           string
	ExpectedStatus int
}

func CheckHTTP(target string, timeout time.Duration, expectedStatus int) (statusCode int, duration time.Duration, err error) {
	code, dur, _, err := CheckHTTPWithOptions(target, timeout, HTTPCheckOptions{
		Method:         "GET",
		ExpectedStatus: expectedStatus,
	})
	return code, dur, err
}

func CheckHTTPWithOptions(target string, timeout time.Duration, opts HTTPCheckOptions) (statusCode int, duration time.Duration, responseSnippet string, err error) {
	if err := utils.ValidateTargetHost(target); err != nil {
		return 0, 0, "", err
	}

	method := strings.ToUpper(strings.TrimSpace(opts.Method))
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if opts.Body != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(opts.Body)
	}

	client := utils.NewSafeHTTPClient(timeout)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("stopped after 5 redirects")
		}
		return utils.ValidateTargetHost(req.URL.String())
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(context.Background(), method, target, bodyReader)
	if err != nil {
		return 0, 0, "", err
	}

	req.Header.Set("User-Agent", "Netra-Uptime/1.0")

	if strings.TrimSpace(opts.HeadersJSON) != "" {
		var hdrMap map[string]string
		if err := json.Unmarshal([]byte(opts.HeadersJSON), &hdrMap); err == nil {
			for k, v := range hdrMap {
				cleanKey := strings.TrimSpace(k)
				if cleanKey == "" || strings.ContainsAny(cleanKey, "\r\n") || strings.ContainsAny(v, "\r\n") {
					continue
				}
				if strings.EqualFold(cleanKey, "Host") {
					continue
				}
				req.Header.Set(cleanKey, v)
			}
		}
	}

	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(req)
	duration = time.Since(start)

	if err != nil {
		return 0, duration, "", err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= 400 && response.Body != nil {
		rawBytes, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseSnippetBytes))
		if readErr == nil && len(rawBytes) > 0 {
			responseSnippet = SanitizeResponseSnippet(rawBytes)
		}
	}

	return response.StatusCode, duration, responseSnippet, nil
}

// SanitizeResponseSnippet cleans raw HTTP response bytes into a safe, single-line printable string capped at 512 bytes.
func SanitizeResponseSnippet(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if len(raw) > maxResponseSnippetBytes {
		raw = raw[:maxResponseSnippetBytes]
	}

	var b strings.Builder
	b.Grow(len(raw))
	lastSpace := false

	for i := 0; i < len(raw); {
		r, width := utf8.DecodeRune(raw[i:])
		if r == utf8.RuneError && width == 1 {
			i++
			continue
		}
		i += width

		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			if !lastSpace && b.Len() > 0 {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}

		if unicode.IsPrint(r) {
			b.WriteRune(r)
			lastSpace = false
		}
	}

	res := strings.TrimSpace(b.String())
	if len(res) > maxResponseSnippetBytes {
		res = res[:maxResponseSnippetBytes]
	}
	return res
}
