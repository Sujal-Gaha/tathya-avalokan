package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// InitDB opens a SQLite database using the modernc.org/sqlite driver,
// enforces foreign keys and WAL mode in the connection DSN, and executes the embedded schema.
func InitDB(dsn string) (*sql.DB, error) {
	connDSN := dsn
	if dsn == ":memory:" {
		connDSN = ":memory:?_pragma=foreign_keys(1)"
	} else if !strings.Contains(dsn, "_pragma") {
		delim := "?"
		if strings.Contains(dsn, "?") {
			delim = "&"
		}
		connDSN = fmt.Sprintf("%s%s_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dsn, delim)
	}

	db, err := sql.Open("sqlite", connDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Limit pool connections appropriately for SQLite
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	// Enable Foreign Key support explicitly
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Execute schema migration
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to execute schema DDL: %w", err)
	}

	return db, nil
}
