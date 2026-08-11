package checker

import (
	"net"
	"time"
)

func CheckPing(target string, timeout time.Duration) (duration time.Duration, err error) {
	start := time.Now()

	ips, err := net.LookupIP(target)
	duration = time.Since(start)
	if err != nil || len(ips) == 0 {
		return duration, err
	}

	return duration, err
}
