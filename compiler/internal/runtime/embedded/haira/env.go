package haira

import (
	"os"
	"strconv"
)

func Env(key string) string { return os.Getenv(key) }

func EnvFloat(key string) float64 {
	f, _ := strconv.ParseFloat(os.Getenv(key), 64)
	return f
}

func EnvInt(key string) int {
	n, _ := strconv.Atoi(os.Getenv(key))
	return n
}
