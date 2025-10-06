package shared

import "os"

// GetNodename returns the node name if the NODE_NAME environment variable is set.
// If NODE_NAME is not set, it returns the hostname of the machine.
// If both are not available, it returns "unknown".
func GetNodename() string {
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) > 0 {
		return nodeName
	}

	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		return "unknown"
	}
	return hostname
}
