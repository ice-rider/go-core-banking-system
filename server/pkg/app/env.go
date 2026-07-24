package app

import "os"

// GetEnv returns the value of the environment variable named by key,
// or fallback if it is not set.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
