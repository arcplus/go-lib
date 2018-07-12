package os

import (
	"os"
)

// Env read env with given key, if empty return defaultValue
func Env(key string, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
