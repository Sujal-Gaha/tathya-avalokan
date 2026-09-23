# Tathya-Avalokan Architecture Specification

## 1. System Overview & Core Purpose

**Tathya-Avalokan (तथ्य अवलोकन)** is an open-source, web-based database management client designed to deliver a desktop-grade (VS Code-style) development experience in the browser. 

Traditional database administration interfaces often suffer from complex installation footprints, rigid desktop bindings, or bloated web interfaces. Tathya-Avalokan addresses this with:
- A hierarchical **Project-to-Instance** organization model.
- A secure **Backend-for-Frontend (BFF)** database proxy layer written in **Go (Chi)**.
- Embedded, zero-maintenance internal metadata persistence via **Pure Go SQLite (`modernc.org/sqlite`)**, requiring zero CGO.
- Authenticated **AES-256-GCM** credential encryption at rest.
- A high-performance, modular React 18 + TypeScript single-page application.

---

## 2. Backend-for-Frontend (BFF) Database Proxy Architecture

### The Need for a BFF Proxy
Web browsers cannot natively establish raw TCP socket connections to database servers (such as PostgreSQL on port `5432` or MySQL on port `3306`) due to standard web security boundaries and protocol constraints. Furthermore, exposing raw database connection credentials (usernames, passwords, hostnames) to the client presents critical security liabilities.

Tathya-Avalokan solves this by establishing a dedicated **Go (Chi) BFF Proxy**:

```text
┌────────────────┐           HTTP REST (/api/v1)           ┌──────────────────┐
│                │ ──────────────────────────────────────> │                  │
│                │   POST /instances/{id}/query            │                  │
│                │   { "sql": "SELECT * FROM users;" }     │                  │
│    Browser     │                                         │   Go (Chi) BFF   │
│ Frontend (SPA) │ <────────────────────────────────────── │      Proxy       │
│                │   Unified Response Envelope             │                  │
│                │   { "data": { "rows": [...] } }         │                  │
└────────────────┘                                         └────────┬─────────┘
                                                                    │
                                            Native Go Drivers       │ (Connection Pool,
                                            (pgxpool / go-sql-mysql)│  AES-256-GCM Decryption,
                                                                    │  Context Timeouts)
                                                                    ▼
                                                           ┌──────────────────┐
                                                           │  Target Database │
                                                           │ (Postgres/MySQL) │
                                                           └──────────────────┘
```

### Proxy & Query Execution Lifecycle
1. **Request Intake**: The frontend transmits a structured query request (`sql`, optional pagination `limit`/`offset`, and `timeout_seconds`) to `POST /api/v1/instances/{id}/query`.
2. **Metadata Lookup & Decryption**: The BFF retrieves the target `DatabaseInstance` record from the internal SQLite database (`app_metadata.db`). The encrypted credentials are decrypted in memory using AES-256-GCM authenticated cipher.
3. **Connection Handshake**: The proxy initializes or reuses a database client connection pool (`database/sql` or `pgxpool`).
4. **Execution & Timeout Guardrails**: The query runs within a `context.WithTimeout(...)` execution wrapper. If the target query exceeds the configured timeout threshold (default: 30 seconds), the proxy aborts the database call, cleans up connections, and issues an HTTP 408 / `QUERY_TIMEOUT` error.
5. **Streaming & Serialization**: Result set metadata (column names, inferred data types) and row tuples are converted into JSON-serializable structures and encapsulated within the unified response envelope.

---

## 3. Frontend State Management Strategy

To ensure seamless performance, zero unnecessary re-renders, and complete separation of concerns, the frontend enforces a strict dual-store state architecture:

```text
┌────────────────────────────────────────────────────────────────────────┐
│                          Frontend State Layers                         │
├───────────────────────────────────┬────────────────────────────────────┤
│           Server State            │             Client State           │
│         (TanStack Query)          │              (Zustand)             │
├───────────────────────────────────┼────────────────────────────────────┤
│ • Projects collection             │ • Active Project Selection         │
│ • Database Instances metadata     │ • Active Instance Selection        │
│ • Cached database schema trees    │ • Open Query Editor Tabs & SQL     │
│ • Historical query logs           │ • Sidebar collapse / Panel layouts │
│ • Table column & index metadata   │ • Active modal visibility          │
└───────────────────────────────────┴────────────────────────────────────┘
```

### Server State: TanStack Query
- **Single Source of Truth for Remote Data**: All data originating from or persisted to the backend is managed by `@tanstack/react-query`.
- **Cache Invalidation**: Mutations (e.g. creating an instance or renaming a project) automatically invalidate the relevant query keys (`['projects']`, `['instances', projectId]`, `['schema', instanceId]`).
- **Optimistic Updates**: Fast UI response for non-destructive operations.

### Client State: Zustand
- **Transient UI & Workspace Session**: Managed through lightweight, modular Zustand stores (`useUIStore`).
- **Tab State**: Manages multi-tab query editor sessions (tab IDs, title, SQL buffer content, execution state, active data grid pagination offset).
- **Rule of Thumb**: If data does not live in the backend database and represents the developer's immediate UI interaction state, it belongs in Zustand.

---

## 4. Internal Metadata SQLite Schema

The application persists internal state (projects, configurations, encrypted credentials) in an embedded pure Go SQLite database using standard library `database/sql` and `modernc.org/sqlite` (requiring zero CGO).

### Entity-Relationship Diagram

```text
┌───────────────────────────┐
│          Project          │
├───────────────────────────┤
│ id (PK, UUID)             │ <──┐
│ name (VARCHAR, Indexed)   │    │ 1-to-Many
│ description (TEXT)        │    │ Cascade Delete
│ created_at (DATETIME)     │    │
│ updated_at (DATETIME)     │    │
└───────────────────────────┘    │
                                 │
┌───────────────────────────┐    │
│     DatabaseInstance      │    │
├───────────────────────────┤    │
│ id (PK, UUID)             │    │
│ project_id (FK, UUID)     │ ───┘
│ name (VARCHAR)            │
│ driver_type (ENUM)        │  (postgresql, mysql, sqlite)
│ host (VARCHAR, Nullable)  │
│ port (INTEGER, Nullable)  │
│ database_name (VARCHAR)   │
│ username (VARCHAR, Nullable)
│ encrypted_credentials (TEXT) (AES-256-GCM ciphertext)
│ ssl_mode (VARCHAR)        │  (disable, require, verify-ca, verify-full)
│ is_read_only (BOOLEAN)    │
│ created_at (DATETIME)     │
│ updated_at (DATETIME)     │
└───────────────────────────┘
```

### Table Definitions & Constraints

#### 1. `projects` Table
| Column Name | Type | Constraints | Description |
|---|---|---|---|
| `id` | `VARCHAR(36)` | PRIMARY KEY | RFC 4122 UUID v4 string |
| `name` | `VARCHAR(100)` | NOT NULL, INDEXED | Logical display name for the project |
| `description` | `TEXT` | NULLABLE | Detailed project context or notes |
| `created_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Timestamp of project creation |
| `updated_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Timestamp of last modification |

#### 2. `database_instances` Table
| Column Name | Type | Constraints | Description |
|---|---|---|---|
| `id` | `VARCHAR(36)` | PRIMARY KEY | RFC 4122 UUID v4 string |
| `project_id` | `VARCHAR(36)` | NOT NULL, FK -> `projects.id` | Cascades on project deletion |
| `name` | `VARCHAR(100)` | NOT NULL | Friendly instance identifier (e.g. "Staging DB") |
| `driver_type` | `VARCHAR(20)` | NOT NULL | Driver identifier (`postgresql`, `mysql`, `sqlite`) |
| `host` | `VARCHAR(255)` | NULLABLE | Target hostname or IP (NULL for local SQLite) |
| `port` | `INTEGER` | NULLABLE | Port number (e.g. 5432, 3306) |
| `database_name` | `VARCHAR(100)` | NOT NULL | Target database or catalog name |
| `username` | `VARCHAR(100)` | NULLABLE | Authentication username |
| `encrypted_credentials` | `TEXT` | NOT NULL | AES-256-GCM-encrypted JSON (password, token, connection options) |
| `ssl_mode` | `VARCHAR(30)` | NOT NULL, DEFAULT 'prefer' | Target TLS / SSL verification policy |
| `is_read_only` | `BOOLEAN` | NOT NULL, DEFAULT FALSE | Enforces read-only query guardrails |
| `created_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Creation timestamp |
| `updated_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Modification timestamp |

---

## 5. Security Model & Credential Protection

### Symmetric Encryption at Rest (AES-256-GCM)
Database passwords, authentication tokens, and sensitive connection strings MUST NEVER be stored in plain text.
- Tathya-Avalokan uses Go's standard library `crypto/cipher` and `crypto/aes` for authenticated AES-256-GCM encryption.
- 96-bit random nonces are generated per encryption operation via `crypto/rand` and prepended to the ciphertext.
- Master 256-bit keys are derived using `crypto/sha256` from the `TATHYA_ENCRYPTION_KEY` environment variable.
- During development, if no key is provided, a deterministic fallback key is used alongside an explicit warning in application logs.

### Credential Masking in API Responses
- Passwords and raw encrypted blobs are **strictly excluded** from API responses.
- Responses return `is_password_set: true` instead of exposing sensitive secrets.
- Any generated connection string displays are masked (e.g., `postgresql://dbuser:********@prod-db.internal:5432/orders`).

### Execution Guardrails
- **Statement Timeouts**: Every proxy query has a hard timeout limit enforced by `context.WithTimeout` (configurable, default 30s).
- **Read-Only Mode Enforcement**: Instances configured as `is_read_only=true` reject mutating statements (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`, `REPLACE`). Comment bypass attempts (`--` and `/* */`) are stripped prior to keyword extraction.
- **CORS Protection**: Restricted to trusted frontend origins via `go-chi/cors`.

---

## 6. Response Envelope & Architectural Conventions

Following the project's strict conventions (as established across the monorepo), all API responses follow a unified envelope:

```json
{
  "data": { ... } | [ ... ] | null,
  "error": {
    "code": "ERROR_CODE_STRING",
    "message": "Human readable explanation",
    "details": { ... }
  } | null,
  "metadata": {
    "timestamp": "2026-09-16T15:00:00Z"
  }
}
```

---

## 7. Authentication & Authorization Design

### Current Posture: Single-User Local Tool

Tathya-Avalokan is currently designed as a **single-user, local-first development tool**. There is no user model, session system, or authentication layer in Phase 1 or 2.

> ⚠️ **Important**: The CORS policy defaults to `allow_origins=["*"]` in development. This is intentional for localhost developer use, but **must** be tightened before any deployment reachable over a network. Configure `CORS_ORIGINS` in `.env` with explicit trusted origins.

---

## 8. Connection Pool Strategy (Phase 3 Design)

When Phase 3 introduces live database driver execution, the proxy layer will manage connection pools per `DatabaseInstance`:

### Per-Instance Pool Design

```text
Chi HTTP Request
    │
    ▼
ConnectionPoolRegistry (sync.Map, instance_id → *sql.DB / *pgxpool.Pool)
    │
    ├── pgxpool.Pool (PostgreSQL)  — min_conns=1, max_conns=10
    ├── database/sql (MySQL)       — SetMaxOpenConns(10), SetMaxIdleConns(1)
    └── database/sql (SQLite)      — single connection, serialized
```

### Pool Lifecycle

| Event | Action |
|---|---|
| First query on an instance | Pool created and cached in registry |
| Subsequent queries | Existing pool connection acquired |
| Instance deleted | Pool evicted from registry, connections drained |
| Server shutdown | All pools gracefully closed |
| Query timeout | Context cancellation aborts statement cleanly |

---

## 9. Result Pagination Design

The `QueryRequest` and `QueryResponse` schemas are **already fully defined** to support pagination:
- `QueryRequest.limit` — Maximum rows to return in a single batch (default: 100, max: 5000)
- `QueryRequest.offset` — Row offset for keyset/offset pagination
- `QueryResponse.has_more` — Whether rows beyond `limit` exist

---

## 10. `instances_count` — Live Aggregate vs. Stored Counter

The `instances_count` field returned by `GET /api/v1/projects` is a **live aggregate** computed via `LEFT JOIN` and `COUNT(d.id)` grouped by `p.id`.

### Guarantees
- **Always accurate** — derived from live FK rows, no separate counter to drift
- **Cascade safe** — deleting a project cascades at the DB level; count naturally zeroes
- **Zero synchronization latency** — always strictly consistent
