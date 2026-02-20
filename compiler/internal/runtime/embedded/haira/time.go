package haira

import (
	"fmt"
	"strings"
	"time"
)

// HairaTime wraps Go's time.Time for use in Haira programs.
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

func TimeNow() *HairaTime { return newHairaTime(time.Now()) }

func TimeSleep(ms any) {
	switch s := ms.(type) {
	case int:
		time.Sleep(time.Duration(s) * time.Millisecond)
	case float64:
		time.Sleep(time.Duration(s) * time.Millisecond)
	}
}

func TimeFormat(t any, format string) string {
	switch v := t.(type) {
	case *HairaTime:
		return v.goTime.Format(format)
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return v
		}
		return parsed.Format(format)
	}
	return fmt.Sprintf("%v", t)
}

func TimeParse(dateStr string, format string) (*HairaTime, error) {
	t, err := time.Parse(format, dateStr)
	if err != nil {
		return nil, err
	}
	return newHairaTime(t), nil
}

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

func TimeAfter(ms int) <-chan *HairaTime {
	ch := make(chan *HairaTime, 1)
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		ch <- TimeNow()
		close(ch)
	}()
	return ch
}

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

func TimeSlug() string {
	t := time.Now().UTC().Format("2006-01-02T15:04:05")
	return strings.ReplaceAll(strings.ReplaceAll(t, "-", "."), ":", ".")
}
