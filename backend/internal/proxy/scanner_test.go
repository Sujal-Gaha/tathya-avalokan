package proxy

import (
	"database/sql"
	"math"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestIsRowReturningStatement(t *testing.T) {
	tests := []struct {
		sql      string
		expected bool
	}{
		{"SELECT * FROM users", true},
		{"SHOW TABLES", true},
		{"EXPLAIN SELECT 1", true},
		{"PRAGMA foreign_keys", true},
		{"CREATE TABLE foo (id INT)", false},
		{"DROP TABLE foo", false},
		{"ALTER TABLE foo ADD COLUMN x INT", false},
		{"TRUNCATE TABLE foo", false},
		{"INSERT INTO logs (data) VALUES ('customer_returning_item')", false},
		{"INSERT INTO users (name) VALUES ('Alice') RETURNING id", true},
		{"UPDATE users SET name = 'Bob' RETURNING *", true},
		{"DELETE FROM users WHERE id = 1 RETURNING id", true},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", true},
	}

	for _, tc := range tests {
		got := IsRowReturningStatement(tc.sql)
		if got != tc.expected {
			t.Errorf("IsRowReturningStatement(%q) = %v; want %v", tc.sql, got, tc.expected)
		}
	}
}

func TestNormalizeValue(t *testing.T) {
	// 1. Valid UTF-8 bytes -> string
	b := []byte("hello world")
	if norm := NormalizeValue(b); norm != "hello world" {
		t.Errorf("Expected 'hello world', got %v", norm)
	}

	// 2. Binary bytes -> hex string
	bin := []byte{0xff, 0xfe, 0x00}
	if norm := NormalizeValue(bin); norm != "\\xfffe00" {
		t.Errorf("Expected '\\xfffe00', got %v", norm)
	}

	// 3. Float NaN
	nan := math.NaN()
	if norm := NormalizeValue(nan); norm != "NaN" {
		t.Errorf("Expected 'NaN', got %v", norm)
	}

	// 4. Float Infinity
	inf := math.Inf(1)
	if norm := NormalizeValue(inf); norm != "+Infinity" {
		t.Errorf("Expected '+Infinity', got %v", norm)
	}

	// 5. Time
	now := time.Now().UTC()
	if norm := NormalizeValue(now); norm != now.Format(time.RFC3339Nano) {
		t.Errorf("Expected RFC3339Nano timestamp, got %v", norm)
	}

	// 6. 16-byte UUID
	uuidBytes := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	if norm := NormalizeValue(uuidBytes); norm != "01020304-0506-0708-090a-0b0c0d0e0f10" {
		t.Errorf("Expected UUID string, got %v", norm)
	}

	// 7. Nil
	if norm := NormalizeValue(nil); norm != nil {
		t.Errorf("Expected nil, got %v", norm)
	}
}

func TestScanPaginatedRows(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE items (id INT, name TEXT);
		INSERT INTO items VALUES (1, 'One'), (2, 'Two'), (3, 'Three'), (4, 'Four'), (5, 'Five');
	`)
	if err != nil {
		t.Fatalf("Failed to seed items: %v", err)
	}

	// Case 1: Limit 2, Offset 0 -> 2 rows, hasMore: true
	rows, err := db.Query("SELECT id, name FROM items ORDER BY id ASC")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	cols, res, hasMore, err := ScanPaginatedRows(rows, 2, 0)
	if err != nil {
		t.Fatalf("ScanPaginatedRows failed: %v", err)
	}
	if len(cols) != 2 || len(res) != 2 || !hasMore {
		t.Errorf("Expected 2 rows and hasMore=true, got %d rows, hasMore=%v", len(res), hasMore)
	}
	if res[0]["name"] != "One" || res[1]["name"] != "Two" {
		t.Errorf("Unexpected row data: %v", res)
	}

	// Case 2: Limit 2, Offset 2 -> 2 rows, hasMore: true (items 3, 4)
	rows, _ = db.Query("SELECT id, name FROM items ORDER BY id ASC")
	_, res, hasMore, _ = ScanPaginatedRows(rows, 2, 2)
	if len(res) != 2 || !hasMore {
		t.Errorf("Expected 2 rows and hasMore=true, got %d rows, hasMore=%v", len(res), hasMore)
	}
	if res[0]["name"] != "Three" || res[1]["name"] != "Four" {
		t.Errorf("Unexpected row data: %v", res)
	}

	// Case 3: Limit 2, Offset 4 -> 1 row, hasMore: false (item 5)
	rows, _ = db.Query("SELECT id, name FROM items ORDER BY id ASC")
	_, res, hasMore, _ = ScanPaginatedRows(rows, 2, 4)
	if len(res) != 1 || hasMore {
		t.Errorf("Expected 1 row and hasMore=false, got %d rows, hasMore=%v", len(res), hasMore)
	}
	if res[0]["name"] != "Five" {
		t.Errorf("Unexpected row data: %v", res)
	}

	// Case 4: Limit 2, Offset 10 -> 0 rows, hasMore: false
	rows, _ = db.Query("SELECT id, name FROM items ORDER BY id ASC")
	_, res, hasMore, _ = ScanPaginatedRows(rows, 2, 10)
	if len(res) != 0 || hasMore {
		t.Errorf("Expected 0 rows and hasMore=false, got %d rows, hasMore=%v", len(res), hasMore)
	}
}

func TestDuplicateColumnDisambiguation(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Query with duplicate column names
	rows, err := db.Query("SELECT 1 AS id, 2 AS id, 3 AS id")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	cols, res, _, err := ScanPaginatedRows(rows, 10, 0)
	if err != nil {
		t.Fatalf("ScanPaginatedRows failed: %v", err)
	}

	if len(cols) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(cols))
	}
	if cols[0].Name != "id" || cols[1].Name != "id_1" || cols[2].Name != "id_2" {
		t.Errorf("Expected id, id_1, id_2; got %s, %s, %s", cols[0].Name, cols[1].Name, cols[2].Name)
	}

	if res[0]["id"] == res[0]["id_1"] || res[0]["id_1"] == res[0]["id_2"] {
		t.Errorf("Expected disambiguated values to be distinct: %v", res[0])
	}
}
