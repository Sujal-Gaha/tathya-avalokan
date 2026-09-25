package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/models"

	"github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// BuildDSN generates the appropriate sql.Open driver name and connection string (DSN)
// for a given database instance, safely decrypting any stored credentials.
func BuildDSN(inst *models.DatabaseInstance) (string, string, error) {
	if inst == nil {
		return "", "", fmt.Errorf("database instance is nil")
	}

	var password, connURL string
	if inst.EncryptedCredentials != "" {
		creds, err := crypto.DecryptCredentials(inst.EncryptedCredentials)
		if err == nil && creds != nil {
			password = creds["password"]
			connURL = creds["connection_url"]
		}
	}

	switch inst.DriverType {
	case models.DriverPostgres:
		return buildPostgresDSN(inst, password, connURL)

	case models.DriverMySQL:
		return buildMySQLDSN(inst, password, connURL)

	case models.DriverSQLite:
		return buildSQLiteDSN(inst, connURL)

	default:
		return "", "", fmt.Errorf("unsupported driver type: '%s'", inst.DriverType)
	}
}

func buildPostgresDSN(inst *models.DatabaseInstance, password, connURL string) (string, string, error) {
	// If explicit connection URL is provided in credentials, prioritize it
	if strings.TrimSpace(connURL) != "" {
		return "pgx", connURL, nil
	}

	host := "localhost"
	if inst.Host != nil && strings.TrimSpace(*inst.Host) != "" {
		host = strings.TrimSpace(*inst.Host)
	}

	port := 5432
	if inst.Port != nil && *inst.Port > 0 {
		port = *inst.Port
	}

	username := "postgres"
	if inst.Username != nil && strings.TrimSpace(*inst.Username) != "" {
		username = strings.TrimSpace(*inst.Username)
	}

	sslMode := "prefer"
	if strings.TrimSpace(inst.SSLMode) != "" {
		sslMode = strings.TrimSpace(inst.SSLMode)
	}

	// Build userinfo
	var userInfo *url.Userinfo
	if password != "" {
		userInfo = url.UserPassword(username, password)
	} else if username != "" {
		userInfo = url.User(username)
	}

	u := url.URL{
		Scheme: "postgres",
		User:   userInfo,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   "/" + strings.TrimPrefix(inst.DatabaseName, "/"),
	}

	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return "pgx", u.String(), nil
}

func buildMySQLDSN(inst *models.DatabaseInstance, password, connURL string) (string, string, error) {
	if strings.TrimSpace(connURL) != "" {
		return "mysql", connURL, nil
	}

	host := "127.0.0.1"
	if inst.Host != nil && strings.TrimSpace(*inst.Host) != "" {
		host = strings.TrimSpace(*inst.Host)
	}

	port := 3306
	if inst.Port != nil && *inst.Port > 0 {
		port = *inst.Port
	}

	username := ""
	if inst.Username != nil {
		username = strings.TrimSpace(*inst.Username)
	}

	cfg := mysql.NewConfig()
	cfg.User = username
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	cfg.DBName = strings.TrimPrefix(inst.DatabaseName, "/")
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Timeout = 10 * time.Second
	cfg.ReadTimeout = 30 * time.Second
	cfg.WriteTimeout = 30 * time.Second
	cfg.InterpolateParams = true

	// SSL/TLS mode mapping
	switch strings.ToLower(strings.TrimSpace(inst.SSLMode)) {
	case "disable":
		cfg.TLSConfig = "false"
	case "require":
		cfg.TLSConfig = "skip-verify"
	case "verify-ca", "verify-full":
		cfg.TLSConfig = "true"
	default:
		cfg.TLSConfig = "preferred"
	}

	return "mysql", cfg.FormatDSN(), nil
}

func buildSQLiteDSN(inst *models.DatabaseInstance, connURL string) (string, string, error) {
	dbPath := strings.TrimSpace(inst.DatabaseName)
	if strings.TrimSpace(connURL) != "" {
		dbPath = strings.TrimSpace(connURL)
	}
	if dbPath == "" {
		dbPath = ":memory:"
	}

	if dbPath == ":memory:" {
		return "sqlite", ":memory:?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", nil
	}

	delim := "?"
	if strings.Contains(dbPath, "?") {
		delim = "&"
	}

	dsn := fmt.Sprintf("%s%s_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath, delim)
	return "sqlite", dsn, nil
}
