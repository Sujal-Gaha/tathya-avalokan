package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	MetadataDatabaseURL string
	EncryptionKey       string
	AppSecretKey        string
	CORSOrigins         []string
}

// LoadConfig loads environment variables from .env if present and populates Config with sensible defaults.
func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	rawDBURL := os.Getenv("METADATA_DATABASE_URL")
	if rawDBURL == "" {
		rawDBURL = "./app_metadata.db"
	}
	dbPath := cleanSQLiteDSN(rawDBURL)

	encryptionKey := os.Getenv("TATHYA_ENCRYPTION_KEY")

	fallbackKey := os.Getenv("APP_SECRET_KEY")
	if fallbackKey == "" {
		fallbackKey = "tathya-avalokan-default-dev-secret-key-32b"
	}

	rawCORS := os.Getenv("CORS_ORIGINS")
	var origins []string
	if rawCORS == "" || rawCORS == "*" {
		origins = []string{"*"}
	} else {
		for p := range strings.SplitSeq(rawCORS, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return &Config{
		Port:                port,
		MetadataDatabaseURL: dbPath,
		EncryptionKey:       encryptionKey,
		AppSecretKey:        fallbackKey,
		CORSOrigins:         origins,
	}
}

// cleanSQLiteDSN strips URL prefixes like sqlite:// or sqlite+aiosqlite:/// leaving a valid SQLite path or DSN.
func cleanSQLiteDSN(dsn string) string {
	prefixes := []string{
		"sqlite+aiosqlite:///",
		"sqlite+aiosqlite://",
		"sqlite:///",
		"sqlite://",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(dsn, prefix) {
			return strings.TrimPrefix(dsn, prefix)
		}
	}
	return dsn
}
