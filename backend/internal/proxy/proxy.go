package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tathya-avalokan/backend/internal/guard"
	"tathya-avalokan/backend/internal/models"
)

// ReadOnlyViolationError represents a rejected mutating statement on a read-only instance.
type ReadOnlyViolationError struct {
	Keyword string
}

func (e *ReadOnlyViolationError) Error() string {
	return fmt.Sprintf("Mutating queries are prohibited on read-only database instances (detected keyword: '%s')", e.Keyword)
}

// ProxyEngine manages the execution of database operations against target database instances.
type ProxyEngine struct {
	registry *PoolRegistry
}

// NewProxyEngine initializes a new ProxyEngine with an internal connection pool registry.
func NewProxyEngine() *ProxyEngine {
	return &ProxyEngine{
		registry: NewPoolRegistry(),
	}
}

// Registry returns the underlying PoolRegistry.
func (p *ProxyEngine) Registry() *PoolRegistry {
	return p.registry
}

// EvictPool closes and evicts a cached connection pool for the given instance ID.
func (p *ProxyEngine) EvictPool(instanceID string) error {
	return p.registry.Evict(instanceID)
}

// EvictMultiple closes and evicts cached connection pools for multiple instance IDs.
func (p *ProxyEngine) EvictMultiple(instanceIDs []string) {
	p.registry.EvictMultiple(instanceIDs)
}

// CloseAll gracefully closes all cached connection pools.
func (p *ProxyEngine) CloseAll() error {
	return p.registry.CloseAll()
}

// TestConnection verifies target database connectivity, server version, and ping latency.
func (p *ProxyEngine) TestConnection(ctx context.Context, inst *models.DatabaseInstance) (*models.ConnectionTestResponse, error) {
	if inst == nil {
		return nil, fmt.Errorf("database instance is nil")
	}

	start := time.Now()
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := p.registry.GetOrCreate(testCtx, inst)
	if err != nil {
		p.registry.Evict(inst.ID)
		return nil, err
	}

	if err := db.PingContext(testCtx); err != nil {
		// Do not retain poisoned pools in registry on failure
		p.registry.Evict(inst.ID)
		return nil, err
	}

	// Query server version depending on driver
	var versionQuery string
	switch inst.DriverType {
	case models.DriverSQLite:
		versionQuery = "SELECT sqlite_version();"
	default:
		versionQuery = "SELECT version();"
	}

	var version string
	if err := db.QueryRowContext(testCtx, versionQuery).Scan(&version); err != nil {
		version = fmt.Sprintf("%s Database Target", strings.ToUpper(string(inst.DriverType)))
	}

	latencyMs := float64(time.Since(start).Microseconds()) / 1000.0

	return &models.ConnectionTestResponse{
		Connected:     true,
		LatencyMs:     latencyMs,
		ServerVersion: version,
		Message:       "Connection established successfully",
	}, nil
}

// ExecuteQuery runs an arbitrary SQL statement with timeout and read-only protections,
// returning paginated result sets or affected row counts.
func (p *ProxyEngine) ExecuteQuery(ctx context.Context, inst *models.DatabaseInstance, req models.QueryRequest) (*models.QueryResponse, error) {
	if inst == nil {
		return nil, fmt.Errorf("database instance is nil")
	}

	sqlStr := strings.TrimSpace(req.SQL)
	if sqlStr == "" {
		return nil, fmt.Errorf("SQL statement cannot be empty")
	}

	// 1. Read-only guard check
	if inst.IsReadOnly {
		isMutating, kw := guard.IsMutating(sqlStr)
		if isMutating {
			return nil, &ReadOnlyViolationError{Keyword: kw}
		}
	}

	// 2. Setup execution context timeout
	timeoutSeconds := 30
	if req.TimeoutSeconds > 0 {
		timeoutSeconds = req.TimeoutSeconds
		if timeoutSeconds > 120 {
			timeoutSeconds = 120 // Capped safeguard
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()

	db, err := p.registry.GetOrCreate(execCtx, inst)
	if err != nil {
		return nil, err
	}

	// 3. Dispatch: row-returning query vs DDL/DML execution
	if !IsRowReturningStatement(sqlStr) {
		res, err := db.ExecContext(execCtx, sqlStr)
		if err != nil {
			return nil, err
		}

		rowsAffected, _ := res.RowsAffected()
		execTime := float64(time.Since(start).Microseconds()) / 1000.0

		return &models.QueryResponse{
			Columns:         []models.ColumnMeta{},
			Rows:            []map[string]any{},
			RowsAffected:    int(rowsAffected),
			ExecutionTimeMs: execTime,
			HasMore:         false,
		}, nil
	}

	// Row-producing query
	rows, err := db.QueryContext(execCtx, sqlStr)
	if err != nil {
		return nil, err
	}

	columns, rowsList, hasMore, err := ScanPaginatedRows(rows, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	execTime := float64(time.Since(start).Microseconds()) / 1000.0

	return &models.QueryResponse{
		Columns:         columns,
		Rows:            rowsList,
		RowsAffected:    len(rowsList), // Reflects row count for status bar
		ExecutionTimeMs: execTime,
		HasMore:         hasMore,
	}, nil
}
