package haira

import (
	"os"
	"strconv"
)

// EnvIntOr returns env var as int, falling back to the default if not set or unparseable.
func EnvIntOr(key string, fallback int) int {
	s := os.Getenv(key)
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

// EnvFloatOr returns env var as float64, falling back to the default if not set or unparseable.
func EnvFloatOr(key string, fallback float64) float64 {
	s := os.Getenv(key)
	if s == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return f
}

// Env returns the value of an environment variable.
func Env(key string) string {
	return os.Getenv(key)
}

// EnvFloat returns the value of an environment variable as a float64.
// Returns 0 if the variable is not set or cannot be parsed.
func EnvFloat(key string) float64 {
	s := os.Getenv(key)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// EnvInt returns the value of an environment variable as an int.
// Returns 0 if the variable is not set or cannot be parsed.
func EnvInt(key string) int {
	s := os.Getenv(key)
	n, _ := strconv.Atoi(s)
	return n
}
