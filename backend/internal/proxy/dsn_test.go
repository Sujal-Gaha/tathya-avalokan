package proxy

import (
	"strings"
	"testing"

	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/models"
)

func TestBuildDSN_Postgres(t *testing.T) {
	crypto.InitCipherKey("test-key-32-bytes-long-12345678", "")

	host := "db.internal"
	port := 5433
	user := "app_user"
	creds, err := crypto.EncryptCredentials(map[string]string{
		"password": "s3cr3t@password:with#special",
	})
	if err != nil {
		t.Fatalf("Failed to encrypt test creds: %v", err)
	}

	inst := &models.DatabaseInstance{
		ID:                   "inst-pg-1",
		Name:                 "PG Prod",
		DriverType:           models.DriverPostgres,
		Host:                 &host,
		Port:                 &port,
		DatabaseName:         "analytics",
		Username:             &user,
		EncryptedCredentials: creds,
		SSLMode:              "require",
	}

	driver, dsn, err := BuildDSN(inst)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}

	if driver != "pgx" {
		t.Errorf("Expected driver 'pgx', got: %s", driver)
	}

	if !strings.HasPrefix(dsn, "postgres://") {
		t.Errorf("Expected dsn to start with postgres://, got: %s", dsn)
	}

	if !strings.Contains(dsn, "db.internal:5433") {
		t.Errorf("Expected host:port in dsn, got: %s", dsn)
	}

	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("Expected sslmode=require in dsn, got: %s", dsn)
	}

	// Password with '@' should be properly URL escaped
	if !strings.Contains(dsn, "s3cr3t%40password%3Awith%23special") {
		t.Errorf("Expected URL-encoded password in dsn, got: %s", dsn)
	}
}

func TestBuildDSN_MySQL(t *testing.T) {
	crypto.InitCipherKey("test-key-32-bytes-long-12345678", "")

	host := "mysql.host"
	port := 3307
	user := "root"
	creds, err := crypto.EncryptCredentials(map[string]string{
		"password": "mypass@word:123#",
	})
	if err != nil {
		t.Fatalf("Failed to encrypt test creds: %v", err)
	}

	inst := &models.DatabaseInstance{
		ID:                   "inst-my-1",
		Name:                 "MySQL DB",
		DriverType:           models.DriverMySQL,
		Host:                 &host,
		Port:                 &port,
		DatabaseName:         "users_db",
		Username:             &user,
		EncryptedCredentials: creds,
		SSLMode:              "disable",
	}

	driver, dsn, err := BuildDSN(inst)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}

	if driver != "mysql" {
		t.Errorf("Expected driver 'mysql', got: %s", driver)
	}

	if !strings.Contains(dsn, "tcp(mysql.host:3307)/users_db") {
		t.Errorf("Expected host and db in MySQL DSN, got: %s", dsn)
	}

	if !strings.Contains(dsn, "tls=false") {
		t.Errorf("Expected tls=false for disabled SSL mode, got: %s", dsn)
	}
}

func TestBuildDSN_SQLite(t *testing.T) {
	instMemory := &models.DatabaseInstance{
		ID:           "inst-sq-1",
		Name:         "In-Memory SQLite",
		DriverType:   models.DriverSQLite,
		DatabaseName: ":memory:",
	}

	driver, dsn, err := BuildDSN(instMemory)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if driver != "sqlite" {
		t.Errorf("Expected driver 'sqlite', got: %s", driver)
	}
	if !strings.Contains(dsn, ":memory:") || !strings.Contains(dsn, "_pragma=foreign_keys(1)") {
		t.Errorf("Expected memory dsn with foreign_keys pragma, got: %s", dsn)
	}

	instFile := &models.DatabaseInstance{
		ID:           "inst-sq-2",
		Name:         "File SQLite",
		DriverType:   models.DriverSQLite,
		DatabaseName: "./test_target.db",
	}

	_, dsnFile, err := BuildDSN(instFile)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if !strings.Contains(dsnFile, "journal_mode(WAL)") {
		t.Errorf("Expected WAL mode pragma in file sqlite dsn, got: %s", dsnFile)
	}
}

func TestBuildDSN_Unsupported(t *testing.T) {
	inst := &models.DatabaseInstance{
		ID:         "inst-err",
		DriverType: "oracle",
	}
	_, _, err := BuildDSN(inst)
	if err == nil {
		t.Errorf("Expected error for unsupported driver, got nil")
	}
}
