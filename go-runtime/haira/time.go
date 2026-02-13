package haira

import (
	"strings"
	"time"
)

// TimeSleep pauses execution for the given number of seconds.
func TimeSleep(seconds any) {
	switch s := seconds.(type) {
	case int:
		time.Sleep(time.Duration(s) * time.Second)
	case float64:
		time.Sleep(time.Duration(s * float64(time.Second)))
	}
}

// TimeNow returns the current UTC time as an ISO 8601 string.
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// TimeSlug returns a filename-safe timestamp slug.
// Example: "2026.02.12T14.30.00"
func TimeSlug() string {
	t := time.Now().UTC().Format("2006-01-02T15:04:05")
	return strings.ReplaceAll(strings.ReplaceAll(t, "-", "."), ":", ".")
}
