package checker

import (
	"netra-uptime/pkg/utils"
	"time"
)

func CheckTCP(target string, timeout time.Duration) (duration time.Duration, err error) {
	if err := utils.ValidateTargetHost(target); err != nil {
		return 0, err
	}

	start := time.Now()
	conn, err := utils.SafeDialTimeout("tcp", target, timeout)
	duration = time.Since(start)

	if err != nil {
		return duration, err
	}
	_ = conn.Close()
	return duration, nil
}
