package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tathya-avalokan/backend/internal/database"
	"tathya-avalokan/backend/internal/proxy"
	"tathya-avalokan/backend/internal/repository"

	"github.com/go-chi/chi/v5"
)

func setupTestServer(t *testing.T) *httptest.Server {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}

	pRepo := repository.NewProjectRepository(db)
	iRepo := repository.NewInstanceRepository(db)
	proxyEngine := proxy.NewProxyEngine()

	healthH := NewHealthHandler()
	projH := NewProjectsHandler(pRepo, proxyEngine)
	instH := NewInstancesHandler(pRepo, iRepo, proxyEngine)
	queryH := NewQueryHandler(iRepo, proxyEngine)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthH.HealthCheck)

		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projH.CreateProject)
			r.Get("/", projH.ListProjects)
			r.Get("/{id}", projH.GetProject)
			r.Patch("/{id}", projH.UpdateProject)
			r.Delete("/{id}", projH.DeleteProject)

			r.Post("/{project_id}/instances", instH.CreateInstance)
		})

		r.Route("/instances", func(r chi.Router) {
			r.Get("/{id}", instH.GetInstance)
			r.Patch("/{id}", instH.UpdateInstance)
			r.Delete("/{id}", instH.DeleteInstance)

			r.Post("/{id}/test-connection", queryH.TestConnection)
			r.Post("/{id}/query", queryH.ExecuteQuery)
		})
	})

	return httptest.NewServer(r)
}

func doReq(t *testing.T, method, url string, body any) (*http.Response, map[string]any, string) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	resp.Body.Close()

	var payload map[string]any
	_ = json.Unmarshal(respBytes, &payload)
	return resp, payload, string(respBytes)
}

func TestHealthEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, payload, _ := doReq(t, http.MethodGet, ts.URL+"/api/v1/health", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if payload["error"] != nil {
		t.Errorf("Expected nil error, got %v", payload["error"])
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("Missing data object: %v", payload)
	}
	if data["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %v", data["status"])
	}
	if data["metadata_storage"] != "sqlite+modernc" {
		t.Errorf("Expected sqlite+modernc metadata_storage, got %v", data["metadata_storage"])
	}
}

func TestProjectsCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// 1. Create Project
	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name":        "Test Project",
		"description": "A test",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d: %v", resp.StatusCode, payload)
	}
	data := payload["data"].(map[string]any)
	if data["name"] != "Test Project" || data["instances_count"].(float64) != 0 {
		t.Errorf("Unexpected created project: %v", data)
	}
	projectID := data["id"].(string)

	// 2. List Projects
	resp, payload, _ = doReq(t, http.MethodGet, ts.URL+"/api/v1/projects", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	list := payload["data"].([]any)
	if len(list) == 0 {
		t.Fatal("Expected projects list not to be empty")
	}

	// 3. Patch Project
	resp, payload, _ = doReq(t, http.MethodPatch, ts.URL+"/api/v1/projects/"+projectID, map[string]any{
		"name": "Renamed Project",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["name"] != "Renamed Project" {
		t.Errorf("Expected renamed project, got %v", data["name"])
	}

	// 4. Patch Project with empty body rejected
	resp, _, _ = doReq(t, http.MethodPatch, ts.URL+"/api/v1/projects/"+projectID, map[string]any{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty patch body, got %d", resp.StatusCode)
	}

	// 5. Delete Project
	resp, payload, _ = doReq(t, http.MethodDelete, ts.URL+"/api/v1/projects/"+projectID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["deleted"] != true {
		t.Errorf("Expected deleted: true, got %v", data["deleted"])
	}

	// 6. Verify 404 after deletion
	resp, _, _ = doReq(t, http.MethodGet, ts.URL+"/api/v1/projects/"+projectID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", resp.StatusCode)
	}
}

func TestInstancesCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create project first
	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "Owner Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	// 1. Create Instance
	resp, payload, rawText := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "Test DB",
		"driver_type":   "postgresql",
		"host":          "localhost",
		"port":          5432,
		"database_name": "testdb",
		"username":      "admin",
		"password":      "secret",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d: %v", resp.StatusCode, payload)
	}
	data := payload["data"].(map[string]any)
	if data["is_password_set"] != true {
		t.Errorf("Expected is_password_set: true")
	}
	if strings.Contains(rawText, "secret") {
		t.Errorf("Plaintext password leaked in response text!")
	}
	instanceID := data["id"].(string)

	// 2. Get Instance
	resp, payload, _ = doReq(t, http.MethodGet, ts.URL+"/api/v1/instances/"+instanceID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["name"] != "Test DB" {
		t.Errorf("Expected name 'Test DB', got %v", data["name"])
	}

	// 3. Patch Instance
	resp, payload, _ = doReq(t, http.MethodPatch, ts.URL+"/api/v1/instances/"+instanceID, map[string]any{
		"name":         "New Name",
		"is_read_only": true,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["name"] != "New Name" || data["is_read_only"] != true {
		t.Errorf("Patch instance failed: %v", data)
	}

	// 4. Patch Instance Password (re-encryption, no leak)
	resp, payload, rawText = doReq(t, http.MethodPatch, ts.URL+"/api/v1/instances/"+instanceID, map[string]any{
		"password": "new_pass",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if strings.Contains(rawText, "new_pass") {
		t.Errorf("New password leaked in patch response!")
	}
	data = payload["data"].(map[string]any)
	if data["is_password_set"] != true {
		t.Errorf("Expected is_password_set: true after password patch")
	}

	// 5. Delete Instance
	resp, payload, _ = doReq(t, http.MethodDelete, ts.URL+"/api/v1/instances/"+instanceID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["deleted"] != true {
		t.Errorf("Expected deleted: true")
	}
}

func TestReadOnlyGuardEnforcement(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create project
	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "RO Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	// Create read-only instance
	_, iPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "RO DB",
		"driver_type":   "sqlite",
		"database_name": ":memory:",
		"is_read_only":  true,
	})
	instanceID := iPayload["data"].(map[string]any)["id"].(string)

	// Direct INSERT attempt
	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "INSERT INTO logs VALUES (1)",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", resp.StatusCode)
	}
	errObj := payload["error"].(map[string]any)
	if errObj["code"] != "READ_ONLY_VIOLATION" {
		t.Errorf("Expected READ_ONLY_VIOLATION, got %v", errObj["code"])
	}

	// Comment-bypassed INSERT attempt
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "-- bypass attempt\nINSERT INTO logs VALUES (1)",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for comment bypass, got %d", resp.StatusCode)
	}

	// Block-comment bypassed UPDATE attempt
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "/* block comment */ UPDATE logs SET x = 1",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for block comment bypass, got %d", resp.StatusCode)
	}

	// Parenthesis-wrapped INSERT attempt
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "(INSERT INTO logs VALUES (1))",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for parenthesis-wrapped INSERT, got %d", resp.StatusCode)
	}

	// Mutating CTE attempt
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "WITH d AS (DELETE FROM logs WHERE id=1) SELECT 1",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for mutating CTE, got %d", resp.StatusCode)
	}

	// Multi-statement query with mutating keyword
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "SELECT 1; DROP TABLE logs",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for multi-statement DROP, got %d", resp.StatusCode)
	}

	// SELECT query allowed
	resp, payload, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "SELECT 1 as ok",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for SELECT, got %d: %v", resp.StatusCode, payload)
	}
	data := payload["data"].(map[string]any)
	if data["rows_affected"].(float64) < 1 {
		t.Errorf("Expected rows_affected >= 1, got %v", data["rows_affected"])
	}
}

func TestInstanceWithoutPasswordHasIsPasswordSetFalse(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "SQLite Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "SQLite No Password DB",
		"driver_type":   "sqlite",
		"database_name": "local.db",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d", resp.StatusCode)
	}
	data := payload["data"].(map[string]any)
	if data["is_password_set"] != false {
		t.Errorf("Expected is_password_set: false for instance without credentials, got %v", data["is_password_set"])
	}
	instanceID := data["id"].(string)

	resp, payload, _ = doReq(t, http.MethodGet, ts.URL+"/api/v1/instances/"+instanceID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data = payload["data"].(map[string]any)
	if data["is_password_set"] != false {
		t.Errorf("Expected GET is_password_set: false, got %v", data["is_password_set"])
	}
}

func TestTestConnection(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "Conn Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	_, iPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "Conn DB",
		"driver_type":   "sqlite",
		"database_name": ":memory:",
	})
	instanceID := iPayload["data"].(map[string]any)["id"].(string)

	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/test-connection", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	data := payload["data"].(map[string]any)
	if data["connected"] != true {
		t.Errorf("Expected connected: true, got %v", data["connected"])
	}
}

func TestTestConnection_Failure(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "Fail Conn Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	badPort := 59999
	_, iPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "Unreachable DB",
		"driver_type":   "postgresql",
		"host":          "127.0.0.1",
		"port":          badPort,
		"database_name": "none",
		"ssl_mode":      "disable",
	})
	instanceID := iPayload["data"].(map[string]any)["id"].(string)

	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/test-connection", nil)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected 502 Bad Gateway for unreachable host, got %d", resp.StatusCode)
	}
	errObj := payload["error"].(map[string]any)
	if errObj["code"] != "CONNECTION_FAILED" {
		t.Errorf("Expected CONNECTION_FAILED, got %v", errObj["code"])
	}
}

func TestExecuteQuery_EndToEnd(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// 1. Create project & SQLite instance with shared in-memory DB
	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "Query Exec Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	_, iPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "Query Exec DB",
		"driver_type":   "sqlite",
		"database_name": "file:test_e2e_query?mode=memory&cache=shared",
	})
	instanceID := iPayload["data"].(map[string]any)["id"].(string)

	// 2. Execute DDL: CREATE TABLE
	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "CREATE TABLE products (id INT, name TEXT, price REAL);",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for CREATE TABLE, got %d: %v", resp.StatusCode, payload)
	}

	// 3. Execute DML: INSERT
	resp, payload, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "INSERT INTO products VALUES (1, 'Book', 19.99), (2, 'Pen', 2.50), (3, 'Notebook', 7.00);",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for INSERT, got %d: %v", resp.StatusCode, payload)
	}
	data := payload["data"].(map[string]any)
	if data["rows_affected"].(float64) != 3 {
		t.Errorf("Expected 3 rows affected, got %v", data["rows_affected"])
	}

	// 4. Execute SELECT with pagination
	resp, payload, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql":    "SELECT id, name, price FROM products ORDER BY id ASC;",
		"limit":  2,
		"offset": 0,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for SELECT, got %d: %v", resp.StatusCode, payload)
	}
	data = payload["data"].(map[string]any)
	rows := data["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if data["has_more"] != true {
		t.Errorf("Expected has_more: true, got %v", data["has_more"])
	}
	if data["rows_affected"].(float64) != 2 {
		t.Errorf("Expected rows_affected: 2, got %v", data["rows_affected"])
	}

	// 5. Execute Syntax Error: Expect 400 Bad Request QUERY_EXECUTION_ERROR
	resp, payload, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "SELECT * FORM products;",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for syntax error, got %d", resp.StatusCode)
	}
	errObj := payload["error"].(map[string]any)
	if errObj["code"] != "QUERY_EXECUTION_ERROR" {
		t.Errorf("Expected QUERY_EXECUTION_ERROR, got %v", errObj["code"])
	}
}

func TestPoolEviction_OnCascadeDelete(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// 1. Create project
	_, pPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects", map[string]any{
		"name": "Cascade Eviction Project",
	})
	projectID := pPayload["data"].(map[string]any)["id"].(string)

	// 2. Create instance
	_, iPayload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/projects/"+projectID+"/instances", map[string]any{
		"name":          "Cascade DB",
		"driver_type":   "sqlite",
		"database_name": ":memory:",
	})
	instanceID := iPayload["data"].(map[string]any)["id"].(string)

	// 3. Test connection to instantiate connection pool in registry
	resp, payload, _ := doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/test-connection", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %v", resp.StatusCode, payload)
	}

	// 4. Delete Project (which cascades instance deletion and evicts pools)
	resp, _, _ = doReq(t, http.MethodDelete, ts.URL+"/api/v1/projects/"+projectID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on DeleteProject, got %d", resp.StatusCode)
	}

	// 5. Subsequent query on deleted instance should return 404 NOT_FOUND
	resp, _, _ = doReq(t, http.MethodPost, ts.URL+"/api/v1/instances/"+instanceID+"/query", map[string]any{
		"sql": "SELECT 1;",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found for deleted instance, got %d", resp.StatusCode)
	}
}
