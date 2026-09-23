package guard

import (
	"testing"
)

func TestExtractFirstSQLKeyword(t *testing.T) {
	tests := []struct {
		sql      string
		expected string
	}{
		{"SELECT * FROM users", "select"},
		{"  SELECT 1", "select"},
		{"INSERT INTO t VALUES (1)", "insert"},
		{"(INSERT INTO t VALUES (1))", "insert"},
		{"-- comment\nINSERT INTO t VALUES (1)", "insert"},
		{"/* block */ DELETE FROM t", "delete"},
		{"/* multi\nline */ UPDATE t SET x=1", "update"},
		{"  -- note\n  DROP TABLE users", "drop"},
		{"TRUNCATE TABLE logs", "truncate"},
		{"", ""},
		{"   ", ""},
		{"-- only comment", ""},
		{"/* only block comment */", ""},
		{"();;", ""},
	}

	for _, tt := range tests {
		got := ExtractFirstSQLKeyword(tt.sql)
		if got != tt.expected {
			t.Errorf("ExtractFirstSQLKeyword(%q) = %q, expected %q", tt.sql, got, tt.expected)
		}
	}
}

func TestIsMutating(t *testing.T) {
	mutatingTests := []string{
		"INSERT INTO t VALUES (1)",
		"(INSERT INTO t VALUES (1))",
		"update users set active=1",
		"DELETE from users",
		"DROP table logs",
		"ALTER table logs add col int",
		"truncate table cache",
		"CREATE table test (id int)",
		"replace into t values (1)",
		"WITH d AS (DELETE FROM logs WHERE id=1) SELECT 1",
		"WITH u AS (UPDATE users SET active=0) SELECT 1",
		"WITH ins AS (INSERT INTO users VALUES (1)) SELECT 1",
		"SELECT 1; DROP TABLE users",
		"SELECT 1; DELETE FROM users",
	}

	for _, sql := range mutatingTests {
		isMut, kw := IsMutating(sql)
		if !isMut {
			t.Errorf("Expected IsMutating(%q) to be true, got false (kw: %s)", sql, kw)
		}
	}

	readOnlyTests := []string{
		"SELECT 1",
		"  select * from test",
		"WITH cte AS (SELECT 1 as col) SELECT * FROM cte",
		"EXPLAIN SELECT 1",
	}

	for _, sql := range readOnlyTests {
		isMut, kw := IsMutating(sql)
		if isMut {
			t.Errorf("Expected IsMutating(%q) to be false, got true (kw: %s)", sql, kw)
		}
	}
}
