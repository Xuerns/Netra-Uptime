package cache

import "fmt"

func MonitorDetailKey(orgID, monID string) string {
	return fmt.Sprintf("monitor:org:%s:%s", orgID, monID)
}

func MonitorListKey(orgID string) string {
	return fmt.Sprintf("monitors:org:%s", orgID)
}
