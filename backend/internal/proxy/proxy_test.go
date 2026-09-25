package proxy

import (
	"context"
	"errors"
	"testing"

	"tathya-avalokan/backend/internal/models"
)

func TestProxyEngine_TestConnection(t *testing.T) {
	engine := NewProxyEngine()
	defer engine.CloseAll()

	inst := &models.DatabaseInstance{
		ID:           "test-conn-inst",
		Name:         "SQLite Test",
		DriverType:   models.DriverSQLite,
		DatabaseName: ":memory:",
	}

	res, err := engine.TestConnection(context.Background(), inst)
	if err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}

	if !res.Connected {
		t.Errorf("Expected connected: true")
	}
	if res.LatencyMs < 0 {
		t.Errorf("Expected positive latency, got %f", res.LatencyMs)
	}
	if res.ServerVersion == "" {
		t.Errorf("Expected non-empty ServerVersion")
	}
}

func TestProxyEngine_TestConnection_Failure(t *testing.T) {
	engine := NewProxyEngine()
	defer engine.CloseAll()

	badHost := "127.0.0.1"
	badPort := 59999
	inst := &models.DatabaseInstance{
		ID:           "test-bad-conn",
		Name:         "Bad Postgres",
		DriverType:   models.DriverPostgres,
		Host:         &badHost,
		Port:         &badPort,
		DatabaseName: "nonexistent",
		SSLMode:      "disable",
	}

	_, err := engine.TestConnection(context.Background(), inst)
	if err == nil {
		t.Errorf("Expected connection error for unreachable host, got nil")
	}

	// Verify broken pool was evicted from registry
	if engine.Registry().PoolCount() != 0 {
		t.Errorf("Expected broken pool to be evicted, got %d pools in registry", engine.Registry().PoolCount())
	}
}

func TestProxyEngine_ExecuteQuery_Lifecycle(t *testing.T) {
	engine := NewProxyEngine()
	defer engine.CloseAll()

	// Share an in-memory SQLite target DB using file or shared cache
	inst := &models.DatabaseInstance{
		ID:           "test-exec-inst",
		Name:         "SQLite Exec Target",
		DriverType:   models.DriverSQLite,
		DatabaseName: "file:test_proxy_target?mode=memory&cache=shared",
	}

	ctx := context.Background()

	// 1. DDL: Create table
	ddlRes, err := engine.ExecuteQuery(ctx, inst, models.QueryRequest{
		SQL: "CREATE TABLE employees (id INT, name TEXT, salary REAL);",
	})
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	if len(ddlRes.Columns) != 0 || len(ddlRes.Rows) != 0 {
		t.Errorf("Expected empty columns/rows for DDL, got cols: %d, rows: %d", len(ddlRes.Columns), len(ddlRes.Rows))
	}

	// 2. DML: Insert records
	dmlRes, err := engine.ExecuteQuery(ctx, inst, models.QueryRequest{
		SQL: "INSERT INTO employees VALUES (1, 'Alice', 95000.50), (2, 'Bob', 82000.00);",
	})
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	if dmlRes.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", dmlRes.RowsAffected)
	}

	// 3. SELECT: Query records
	selRes, err := engine.ExecuteQuery(ctx, inst, models.QueryRequest{
		SQL:   "SELECT id, name, salary FROM employees ORDER BY id ASC;",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if len(selRes.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(selRes.Columns))
	}
	if len(selRes.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(selRes.Rows))
	}
	if selRes.RowsAffected != 2 {
		t.Errorf("Expected rows_affected = 2, got %d", selRes.RowsAffected)
	}
	if selRes.Rows[0]["name"] != "Alice" || selRes.Rows[1]["name"] != "Bob" {
		t.Errorf("Unexpected row content: %v", selRes.Rows)
	}

	// 4. Read-only guardrail test
	readOnlyInst := &models.DatabaseInstance{
		ID:           "test-exec-ro",
		Name:         "SQLite RO Target",
		DriverType:   models.DriverSQLite,
		DatabaseName: "file:test_proxy_target?mode=memory&cache=shared",
		IsReadOnly:   true,
	}

	// Attempt mutating query on read-only instance
	_, err = engine.ExecuteQuery(ctx, readOnlyInst, models.QueryRequest{
		SQL: "DROP TABLE employees;",
	})
	if err == nil {
		t.Fatalf("Expected error for DROP TABLE on read-only instance, got nil")
	}
	var roErr *ReadOnlyViolationError
	if !errors.As(err, &roErr) {
		t.Errorf("Expected ReadOnlyViolationError, got: %v", err)
	}

	// SELECT on read-only instance should succeed
	roSelRes, err := engine.ExecuteQuery(ctx, readOnlyInst, models.QueryRequest{
		SQL: "SELECT count(*) AS total FROM employees;",
	})
	if err != nil {
		t.Fatalf("SELECT on read-only instance failed: %v", err)
	}
	if len(roSelRes.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(roSelRes.Rows))
	}
}
