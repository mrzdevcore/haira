package haira

import (
	"os"
	"strconv"
)

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
