package checker

import (
	"net"
	"time"
)

func CheckTCP(target string, timeout time.Duration) (duration time.Duration, err error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", target, timeout)
	duration = time.Since(start)

	if err != nil {
		return duration, err
	}
	conn.Close()
	return duration, nil
}
