package env

import (
	"os"
	"strconv"
)

func GetEnvString(key string, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intValue, err := strconv.Atoi(val); err == nil {
			return intValue
		}
	}
	return fallback
}

func GetFormattedDsn() string {
	dsn := GetEnvString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable")
	return dsn
}
