package proxy

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"tathya-avalokan/backend/internal/models"
)

// PoolRegistry manages active connection pools (*sql.DB) for database instances.
// It is fully thread-safe and enforces double-checked locking to avoid duplicate pools.
type PoolRegistry struct {
	mu    sync.RWMutex
	pools map[string]*sql.DB
}

// NewPoolRegistry initializes an empty connection pool registry.
func NewPoolRegistry() *PoolRegistry {
	return &PoolRegistry{
		pools: make(map[string]*sql.DB),
	}
}

// GetOrCreate retrieves an existing connection pool for an instance or initializes a new one.
// Implements double-checked locking under sync.RWMutex to prevent race conditions.
func (r *PoolRegistry) GetOrCreate(ctx context.Context, inst *models.DatabaseInstance) (*sql.DB, error) {
	if inst == nil {
		return nil, fmt.Errorf("database instance is nil")
	}

	// 1. Fast read lock check
	r.mu.RLock()
	if db, exists := r.pools[inst.ID]; exists {
		r.mu.RUnlock()
		return db, nil
	}
	r.mu.RUnlock()

	// 2. Acquire write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check to prevent concurrent initialization race
	if db, exists := r.pools[inst.ID]; exists {
		return db, nil
	}

	driverName, dsn, err := BuildDSN(inst)
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN for instance '%s': %w", inst.ID, err)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure pool limits appropriately per driver
	switch inst.DriverType {
	case models.DriverSQLite:
		// Serialized single connection for SQLite to eliminate locking conflicts
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)

	default: // Postgres and MySQL
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(1 * time.Hour)
		db.SetConnMaxIdleTime(15 * time.Minute)
	}

	r.pools[inst.ID] = db
	return db, nil
}

// Evict gracefully closes and evicts a cached connection pool for a given instance ID.
func (r *PoolRegistry) Evict(instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	db, exists := r.pools[instanceID]
	if !exists {
		return nil
	}

	delete(r.pools, instanceID)
	return db.Close()
}

// EvictMultiple closes and evicts connection pools for a slice of instance IDs.
func (r *PoolRegistry) EvictMultiple(instanceIDs []string) {
	if len(instanceIDs) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range instanceIDs {
		if db, exists := r.pools[id]; exists {
			delete(r.pools, id)
			_ = db.Close()
		}
	}
}

// CloseAll closes all cached connection pools. Called during server graceful shutdown.
func (r *PoolRegistry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for id, db := range r.pools {
		if err := db.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(r.pools, id)
	}

	return firstErr
}

// PoolCount returns the number of active connection pools in the registry (primarily for testing/observability).
func (r *PoolRegistry) PoolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.pools)
}
