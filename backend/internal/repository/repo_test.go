package repository

import (
	"context"
	"testing"

	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/database"
	"tathya-avalokan/backend/internal/models"
)

func setupTestDB(t *testing.T) (*ProjectRepository, *InstanceRepository) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize in-memory DB: %v", err)
	}

	pRepo := NewProjectRepository(db)
	iRepo := NewInstanceRepository(db)
	return pRepo, iRepo
}

func TestProjectAndInstanceLifecycle(t *testing.T) {
	pRepo, iRepo := setupTestDB(t)
	ctx := context.Background()

	// 1. Create Project
	desc := "A sample project"
	p, err := pRepo.CreateProject(ctx, models.ProjectCreate{
		Name:        "E-Commerce",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if p.Name != "E-Commerce" || p.InstancesCount != 0 {
		t.Errorf("Unexpected project state: %+v", p)
	}

	// 2. Create Instance under Project
	host := "localhost"
	port := 5432
	user := "admin"
	testBlob, err := crypto.EncryptCredentials(map[string]string{"password": "secret_password"})
	if err != nil {
		t.Fatalf("EncryptCredentials failed: %v", err)
	}

	inst, err := iRepo.CreateInstance(ctx, p.ID, models.DatabaseInstanceCreate{
		Name:         "Main Postgres",
		DriverType:   models.DriverPostgres,
		Host:         &host,
		Port:         &port,
		DatabaseName: "ecommerce_db",
		Username:     &user,
	}, testBlob)
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}
	if inst.Name != "Main Postgres" || !inst.IsPasswordSet {
		t.Errorf("Unexpected instance state: %+v", inst)
	}

	// 3. List Projects — check instances_count is 1
	list, err := pRepo.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(list) != 1 || list[0].InstancesCount != 1 {
		t.Errorf("Expected 1 project with 1 instance, got: %+v", list)
	}

	// 4. Get Project with nested instances
	fullProj, err := pRepo.GetProjectByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProjectByID failed: %v", err)
	}
	if fullProj == nil || fullProj.Instances == nil || len(*fullProj.Instances) != 1 {
		t.Fatalf("Expected nested instances list with 1 item, got: %+v", fullProj)
	}

	// 5. Update Instance
	newName := "Production Postgres"
	isRO := true
	updatedInst, err := iRepo.UpdateInstance(ctx, inst.ID, models.DatabaseInstanceUpdate{
		Name:       &newName,
		IsReadOnly: &isRO,
	}, nil)
	if err != nil {
		t.Fatalf("UpdateInstance failed: %v", err)
	}
	if updatedInst.Name != newName || !updatedInst.IsReadOnly {
		t.Errorf("Instance update failed: %+v", updatedInst)
	}

	// 6. Delete Project (Cascades to instance)
	deleted, err := pRepo.DeleteProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	if !deleted {
		t.Error("Expected project to be deleted")
	}

	// Confirm child instance was cascade-deleted
	orphan, err := iRepo.GetInstanceByID(ctx, inst.ID)
	if err != nil {
		t.Fatalf("GetInstanceByID after cascade failed: %v", err)
	}
	if orphan != nil {
		t.Errorf("Expected instance to be cascade deleted, but found: %+v", orphan)
	}
}
