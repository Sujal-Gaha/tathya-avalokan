# Tathya-Avalokan REST API Specification

- **Version**: 1.0.0
- **Base URL**: `/api/v1`
- **Protocol**: HTTP/1.1 & HTTP/2 over TLS
- **Data Format**: `application/json`

---

## 1. Unified Response Envelope Standard

Every API response follows the unified envelope pattern:

```typescript
interface ApiResponseEnvelope<T> {
  data: T | null;
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown> | null;
  } | null;
  metadata: Record<string, unknown>;
}
```

### Standard Error Codes
| Code | HTTP Status | Description |
|---|---|---|
| `VALIDATION_ERROR` | 400, 422 | Request body or parameters failed validation schemas |
| `NOT_FOUND` | 404 | Target resource (project, instance) does not exist |
| `CONNECTION_FAILED` | 502 | Target database was unreachable or rejected handshake |
| `QUERY_EXECUTION_ERROR` | 400 | Syntax error, permission error, or constraint violation on target DB |
| `QUERY_TIMEOUT` | 408 | Query execution exceeded configured timeout threshold |
| `READ_ONLY_VIOLATION` | 403 | Attempted mutating query on read-only database instance |
| `INTERNAL_ERROR` | 500 | Unhandled proxy or server exception |

---

## 2. API Endpoints

### 2.1 System & Health

#### `GET /api/v1/health`
Checks server health, proxy capability, and SQLite metadata connectivity.

**Response `200 OK`**:
```json
{
  "data": {
    "status": "healthy",
    "version": "0.1.0",
    "metadata_database": "connected",
    "supported_drivers": ["postgresql", "mysql", "sqlite"]
  },
  "error": null,
  "metadata": {
    "timestamp": "2026-09-16T15:00:00Z"
  }
}
```

---

### 2.2 Projects API

#### `POST /api/v1/projects`
Creates a new project workspace.

**Request Body**:
```json
{
  "name": "E-Commerce Microservices",
  "description": "Primary transactional and reporting databases for retail services"
}
```

**Response `201 Created`**:
```json
{
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "E-Commerce Microservices",
    "description": "Primary transactional and reporting databases for retail services",
    "instances_count": 0,
    "created_at": "2026-09-16T15:00:00Z",
    "updated_at": "2026-09-16T15:00:00Z"
  },
  "error": null,
  "metadata": {}
}
```

---

#### `GET /api/v1/projects`
Lists all registered projects with associated instance summary counts.

**Response `200 OK`**:
```json
{
  "data": [
    {
      "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      "name": "E-Commerce Microservices",
      "description": "Primary transactional and reporting databases for retail services",
      "instances_count": 2,
      "created_at": "2026-09-16T15:00:00Z",
      "updated_at": "2026-09-16T15:00:00Z"
    }
  ],
  "error": null,
  "metadata": {
    "total_count": 1
  }
}
```

---

#### `GET /api/v1/projects/{id}`
Retrieves a specific project along with all nested database instances.

**Response `200 OK`**:
```json
{
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "E-Commerce Microservices",
    "description": "Primary transactional and reporting databases for retail services",
    "instances": [
      {
        "id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
        "project_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
        "name": "Production Orders DB",
        "driver_type": "postgresql",
        "host": "postgres.internal.cloud",
        "port": 5432,
        "database_name": "orders_prod",
        "username": "orders_app",
        "is_password_set": true,
        "ssl_mode": "require",
        "is_read_only": true,
        "created_at": "2026-09-16T15:05:00Z",
        "updated_at": "2026-09-16T15:05:00Z"
      }
    ],
    "created_at": "2026-09-16T15:00:00Z",
    "updated_at": "2026-09-16T15:00:00Z"
  },
  "error": null,
  "metadata": {}
}
```

**Response `404 Not Found`**:
```json
{
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "Project with id '9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d' was not found"
  },
  "metadata": {}
}
```

---

#### `DELETE /api/v1/projects/{id}`
Deletes a project and all associated database instances (cascading delete).

**Response `200 OK`**:
```json
{
  "data": {
    "deleted": true,
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
  },
  "error": null,
  "metadata": {}
}
```

---

#### `PATCH /api/v1/projects/{id}`
Partially updates a project's name or description. Only the provided fields are modified.

**Request Body** (all fields optional):
```json
{
  "name": "Updated Platform Name",
  "description": "New description text"
}
```

**Response `200 OK`**:
```json
{
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Updated Platform Name",
    "description": "New description text",
    "instances_count": 2,
    "created_at": "2026-09-16T15:00:00Z",
    "updated_at": "2026-09-17T08:30:00Z"
  },
  "error": null,
  "metadata": {}
}
```

**Response `400 Bad Request`** (empty body):
```json
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "No update fields provided in request body"
  },
  "metadata": {}
}
```

---

### 2.3 Instances API

#### `POST /api/v1/projects/{project_id}/instances`
Registers a new database instance under an existing project. Sensitive credentials (such as password) are symmetrically encrypted before persistence.

**Request Body**:
```json
{
  "name": "Production Orders DB",
  "driver_type": "postgresql",
  "host": "postgres.internal.cloud",
  "port": 5432,
  "database_name": "orders_prod",
  "username": "orders_app",
  "password": "SuperSecretPassword123!",
  "ssl_mode": "require",
  "is_read_only": false
}
```

**Response `201 Created`**:
```json
{
  "data": {
    "id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
    "project_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Production Orders DB",
    "driver_type": "postgresql",
    "host": "postgres.internal.cloud",
    "port": 5432,
    "database_name": "orders_prod",
    "username": "orders_app",
    "is_password_set": true,
    "ssl_mode": "require",
    "is_read_only": false,
    "created_at": "2026-09-16T15:05:00Z",
    "updated_at": "2026-09-16T15:05:00Z"
  },
  "error": null,
  "metadata": {}
}
```

---

#### `GET /api/v1/instances/{id}`
Retrieves details for a single database instance. Sensitive passwords remain masked.

**Response `200 OK`**:
```json
{
  "data": {
    "id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
    "project_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Production Orders DB",
    "driver_type": "postgresql",
    "host": "postgres.internal.cloud",
    "port": 5432,
    "database_name": "orders_prod",
    "username": "orders_app",
    "is_password_set": true,
    "ssl_mode": "require",
    "is_read_only": false,
    "created_at": "2026-09-16T15:05:00Z",
    "updated_at": "2026-09-16T15:05:00Z"
  },
  "error": null,
  "metadata": {}
}
```

---

#### `DELETE /api/v1/instances/{id}`
Removes a database instance configuration.

**Response `200 OK`**:
```json
{
  "data": {
    "deleted": true,
    "id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  },
  "error": null,
  "metadata": {}
}
```

---

### 2.4 Proxy & Query Execution API

#### `POST /api/v1/instances/{id}/test-connection`
Attempts a real-time TCP handshake and authentication ping against the target database instance.

**Response `200 OK` (Successful Connection)**:
```json
{
  "data": {
    "connected": true,
    "latency_ms": 14.2,
    "server_version": "PostgreSQL 16.2 on x86_64-pc-linux-gnu",
    "message": "Connection established successfully"
  },
  "error": null,
  "metadata": {}
}
```

**Response `502 Bad Gateway` (Connection Failure)**:
```json
{
  "data": {
    "connected": false,
    "latency_ms": 1002.5,
    "message": "Password authentication failed for user 'orders_app'"
  },
  "error": {
    "code": "CONNECTION_FAILED",
    "message": "Unable to connect to target database instance: Authentication failed.",
    "details": {
      "host": "postgres.internal.cloud",
      "port": 5432,
      "driver_type": "postgresql"
    }
  },
  "metadata": {}
}
```

---

#### `POST /api/v1/instances/{id}/query`
Executes an arbitrary SQL statement or query against the target database through the proxy engine. Supports timeout limits and pagination constraints.

**Request Body**:
```json
{
  "sql": "SELECT id, order_number, total_amount, created_at FROM orders ORDER BY created_at DESC LIMIT 50;",
  "limit": 50,
  "offset": 0,
  "timeout_seconds": 15
}
```

**Response `200 OK` (SELECT Query)**:
```json
{
  "data": {
    "columns": [
      { "name": "id", "type": "uuid" },
      { "name": "order_number", "type": "varchar" },
      { "name": "total_amount", "type": "numeric" },
      { "name": "created_at", "type": "timestamptz" }
    ],
    "rows": [
      {
        "id": "8c0a37e5-1a2f-48e2-a0b4-3a5e8c1d3b2a",
        "order_number": "ORD-2026-90412",
        "total_amount": 129.99,
        "created_at": "2026-09-16T14:30:10Z"
      },
      {
        "id": "5f1b29d4-8c3e-42a1-b9e7-2b4a7d0c2e1f",
        "order_number": "ORD-2026-90411",
        "total_amount": 45.00,
        "created_at": "2026-09-16T14:28:45Z"
      }
    ],
    "rows_affected": 2,
    "execution_time_ms": 18.7,
    "has_more": false
  },
  "error": null,
  "metadata": {
    "executed_at": "2026-09-16T15:10:00Z"
  }
}
```

**Response `408 Request Timeout`**:
```json
{
  "data": null,
  "error": {
    "code": "QUERY_TIMEOUT",
    "message": "Query execution exceeded the timeout threshold of 15 seconds."
  },
  "metadata": {}
}
```

**Response `400 Bad Request` (SQL Syntax Error)**:
```json
{
  "data": null,
  "error": {
    "code": "QUERY_EXECUTION_ERROR",
    "message": "syntax error at or near 'FORM'",
    "details": {
      "position": "15"
    }
  },
  "metadata": {}
}
```

---

#### `PATCH /api/v1/instances/{id}`
Partially updates a database instance configuration. All fields are optional. If `password` or `connection_url` is provided, credentials are re-encrypted at rest.

**Request Body** (all fields optional):
```json
{
  "name": "Renamed Staging DB",
  "host": "new-host.internal",
  "port": 5433,
  "database_name": "orders_staging_v2",
  "username": "new_user",
  "password": "NewSecret456!",
  "ssl_mode": "require",
  "is_read_only": true
}
```

**Response `200 OK`**:
```json
{
  "data": {
    "id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
    "project_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Renamed Staging DB",
    "driver_type": "postgresql",
    "host": "new-host.internal",
    "port": 5433,
    "database_name": "orders_staging_v2",
    "username": "new_user",
    "is_password_set": true,
    "ssl_mode": "require",
    "is_read_only": true,
    "created_at": "2026-09-16T15:05:00Z",
    "updated_at": "2026-09-17T09:00:00Z"
  },
  "error": null,
  "metadata": {}
}
```

---

### 2.5 Schema Introspection API *(Phase 6 — Planned)*

> **Status**: Not yet implemented. Planned for Phase 6 (Schema Visualizer & Autocomplete).

#### `GET /api/v1/instances/{id}/schema`
Introspects the target database and returns a structural tree of all tables, columns, types, indexes, and foreign key relationships.

**Response `200 OK`** *(planned shape)*:
```json
{
  "data": {
    "tables": [
      {
        "name": "orders",
        "schema": "public",
        "columns": [
          { "name": "id", "type": "uuid", "nullable": false, "primary_key": true },
          { "name": "user_id", "type": "uuid", "nullable": false },
          { "name": "total_amount", "type": "numeric", "nullable": false },
          { "name": "created_at", "type": "timestamptz", "nullable": false }
        ],
        "indexes": [
          { "name": "orders_pkey", "columns": ["id"], "unique": true }
        ],
        "foreign_keys": [
          { "column": "user_id", "references_table": "users", "references_column": "id" }
        ]
      }
    ],
    "views": [],
    "total_tables": 1
  },
  "error": null,
  "metadata": {
    "introspected_at": "2026-09-17T10:00:00Z",
    "cached": false
  }
}
```

TanStack Query cache key: `['schema', instanceId]` with a 5-minute stale time.
