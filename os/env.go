package os

import (
	"os"
)

// Getenv try to read ENV with optional default value.
func Getenv(key string, defaultValue ...string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if len(defaultValue) != 0 {
		return defaultValue[0]
	}
	return ""
}

// Deprecated, using Getenv instead.
func Env(key string, defaultValue ...string) string {
	return Getenv(key, defaultValue...)
}
