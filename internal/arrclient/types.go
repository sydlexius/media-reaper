package arrclient

import "time"

const defaultTimeout = 30 * time.Second

// SystemInfo is the common response from a successful connection test.
type SystemInfo struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}
