package checker

import (
	"context"
	"errors"
	"net/http"
	"netra-uptime/pkg/utils"
	"time"
)

func CheckHTTP(target string, timeout time.Duration, expectedStatus int) (statusCode int, duration time.Duration, err error) {
	if err := utils.ValidateTargetHost(target); err != nil {
		return 0, 0, err
	}

	client := utils.NewSafeHTTPClient(timeout)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("stopped after 5 redirects")
		}
		return utils.ValidateTargetHost(req.URL.String())
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(context.Background(), "GET", target, nil)
	if err != nil {
		return 0, 0, err
	}

	response, err := client.Do(req)
	duration = time.Since(start)

	if err != nil {
		return 0, duration, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	return response.StatusCode, duration, nil
}
