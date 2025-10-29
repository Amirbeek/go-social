package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(val); err == nil {
			return v
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if v, err := strconv.ParseBool(val); err == nil {
			return v
		}
	}
	return fallback
}
