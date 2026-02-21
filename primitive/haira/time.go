package haira

import (
	"fmt"
	"strings"
	"time"
)

// HairaTime wraps Go's time.Time for use in Haira programs.
// Fields are accessible as .year, .month, .day, .hour, .minute, .second.
type HairaTime struct {
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
	Second int    `json:"second"`
	Unix   int64  `json:"unix"`
	Iso    string `json:"iso"`
	goTime time.Time
}

func newHairaTime(t time.Time) *HairaTime {
	return &HairaTime{
		Year:   t.Year(),
		Month:  int(t.Month()),
		Day:    t.Day(),
		Hour:   t.Hour(),
		Minute: t.Minute(),
		Second: t.Second(),
		Unix:   t.Unix(),
		Iso:    t.Format(time.RFC3339),
		goTime: t,
	}
}

// TimeNow returns the current time as a HairaTime object.
func TimeNow() *HairaTime {
	return newHairaTime(time.Now())
}

// TimeSleep pauses execution for the given number of milliseconds.
func TimeSleep(ms any) {
	switch s := ms.(type) {
	case int:
		time.Sleep(time.Duration(s) * time.Millisecond)
	case float64:
		time.Sleep(time.Duration(s) * time.Millisecond)
	}
}

// TimeFormat formats a HairaTime using a Go time layout string.
func TimeFormat(t any, format string) string {
	switch v := t.(type) {
	case *HairaTime:
		return v.goTime.Format(format)
	case string:
		// If passed a string, try parsing as ISO first
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return v
		}
		return parsed.Format(format)
	}
	return fmt.Sprintf("%v", t)
}

// TimeParse parses a time string using the given Go time layout.
func TimeParse(dateStr string, format string) (*HairaTime, error) {
	t, err := time.Parse(format, dateStr)
	if err != nil {
		return nil, err
	}
	return newHairaTime(t), nil
}

// TimeSince returns the duration in milliseconds since the given time.
func TimeSince(t any) int64 {
	switch v := t.(type) {
	case *HairaTime:
		return time.Since(v.goTime).Milliseconds()
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0
		}
		return time.Since(parsed).Milliseconds()
	}
	return 0
}

// TimeAfter returns a channel that receives after the given milliseconds.
func TimeAfter(ms int) <-chan *HairaTime {
	ch := make(chan *HairaTime, 1)
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		ch <- TimeNow()
		close(ch)
	}()
	return ch
}

// TimeTick returns a channel that fires every ms milliseconds.
func TimeTick(ms int) <-chan *HairaTime {
	ch := make(chan *HairaTime)
	go func() {
		ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			ch <- TimeNow()
		}
	}()
	return ch
}

// TimeSlug returns a filename-safe timestamp slug.
func TimeSlug() string {
	t := time.Now().UTC().Format("2006-01-02T15:04:05")
	return strings.ReplaceAll(strings.ReplaceAll(t, "-", "."), ":", ".")
}

// TimeTimestamp returns the current time as an ISO 8601 string.
func TimeTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
