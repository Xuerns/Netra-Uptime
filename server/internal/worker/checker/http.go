package checker

import (
	"context"
	"net/http"
	"time"
)

func CheckHTTP(target string, timeout time.Duration, expectedStatus int) (statusCode int, duration time.Duration, err error) {
	client := &http.Client{
		Timeout: timeout,
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
