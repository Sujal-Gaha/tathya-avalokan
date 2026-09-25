package proxy

import (
	"context"
	"sync"
	"testing"

	"tathya-avalokan/backend/internal/models"
)

func TestPoolRegistry_Lifecycle(t *testing.T) {
	registry := NewPoolRegistry()
	ctx := context.Background()

	inst1 := &models.DatabaseInstance{
		ID:           "test-inst-1",
		Name:         "SQLite 1",
		DriverType:   models.DriverSQLite,
		DatabaseName: ":memory:",
	}
	inst2 := &models.DatabaseInstance{
		ID:           "test-inst-2",
		Name:         "SQLite 2",
		DriverType:   models.DriverSQLite,
		DatabaseName: ":memory:",
	}

	// 1. GetOrCreate creates first pool
	db1, err := registry.GetOrCreate(ctx, inst1)
	if err != nil {
		t.Fatalf("Failed to create db1: %v", err)
	}
	if registry.PoolCount() != 1 {
		t.Errorf("Expected 1 pool in registry, got %d", registry.PoolCount())
	}

	// 2. Subsequent call returns same instance
	db1Again, err := registry.GetOrCreate(ctx, inst1)
	if err != nil {
		t.Fatalf("Failed to get db1 again: %v", err)
	}
	if db1 != db1Again {
		t.Errorf("Expected identical *sql.DB instance on reuse")
	}

	// 3. Create second instance pool
	db2, err := registry.GetOrCreate(ctx, inst2)
	if err != nil {
		t.Fatalf("Failed to create db2: %v", err)
	}
	if db1 == db2 {
		t.Errorf("Expected distinct *sql.DB for different instances")
	}
	if registry.PoolCount() != 2 {
		t.Errorf("Expected 2 pools in registry, got %d", registry.PoolCount())
	}

	// 4. Evict single instance
	if err := registry.Evict(inst1.ID); err != nil {
		t.Fatalf("Failed to evict inst1: %v", err)
	}
	if registry.PoolCount() != 1 {
		t.Errorf("Expected 1 pool remaining after evicting inst1, got %d", registry.PoolCount())
	}

	// 5. Re-create and EvictMultiple
	_, _ = registry.GetOrCreate(ctx, inst1)
	registry.EvictMultiple([]string{inst1.ID, inst2.ID})
	if registry.PoolCount() != 0 {
		t.Errorf("Expected 0 pools after EvictMultiple, got %d", registry.PoolCount())
	}

	// 6. CloseAll
	_, _ = registry.GetOrCreate(ctx, inst1)
	if err := registry.CloseAll(); err != nil {
		t.Fatalf("Failed CloseAll: %v", err)
	}
	if registry.PoolCount() != 0 {
		t.Errorf("Expected 0 pools after CloseAll, got %d", registry.PoolCount())
	}
}

func TestPoolRegistry_ConcurrentDoubleCheck(t *testing.T) {
	registry := NewPoolRegistry()
	ctx := context.Background()

	inst := &models.DatabaseInstance{
		ID:           "concurrent-inst",
		Name:         "SQLite Concurrent",
		DriverType:   models.DriverSQLite,
		DatabaseName: ":memory:",
	}

	var wg sync.WaitGroup
	const goroutines = 20
	results := make([]any, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			db, err := registry.GetOrCreate(ctx, inst)
			if err != nil {
				results[idx] = err
				return
			}
			results[idx] = db
		}(i)
	}

	wg.Wait()

	if registry.PoolCount() != 1 {
		t.Errorf("Expected exactly 1 pool created under concurrency, got %d", registry.PoolCount())
	}

	firstDB := results[0]
	for i, res := range results {
		if res != firstDB {
			t.Errorf("Goroutine %d got a different pool or error: %v", i, res)
		}
	}

	_ = registry.CloseAll()
}
