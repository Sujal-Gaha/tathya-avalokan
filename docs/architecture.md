# Tathya-Avalokan Architecture Specification

## 1. System Overview & Core Purpose

**Tathya-Avalokan (तथ्य अवलोकन)** is an open-source, web-based database management client designed to deliver a desktop-grade (VS Code-style) development experience in the browser. 

Traditional database administration interfaces often suffer from complex installation footprints, rigid desktop bindings, or bloated web interfaces. Tathya-Avalokan addresses this with:
- A hierarchical **Project-to-Instance** organization model.
- A secure **Backend-for-Frontend (BFF)** database proxy layer.
- Embedded, zero-maintenance internal metadata persistence via **SQLite (`aiosqlite`)**.
- A high-performance, modular React 18 + TypeScript single-page application.

---

## 2. Backend-for-Frontend (BFF) Database Proxy Architecture

### The Need for a BFF Proxy
Web browsers cannot natively establish raw TCP socket connections to database servers (such as PostgreSQL on port `5432` or MySQL on port `3306`) due to standard web security boundaries and protocol constraints. Furthermore, exposing raw database connection credentials (usernames, passwords, hostnames) to the client presents critical security liabilities.

Tathya-Avalokan solves this by establishing a dedicated **FastAPI BFF Proxy**:

```text
┌────────────────┐           HTTP REST (/api/v1)           ┌──────────────────┐
│                │ ──────────────────────────────────────> │                  │
│                │   POST /instances/{id}/query            │                  │
│                │   { "sql": "SELECT * FROM users;" }     │                  │
│    Browser     │                                         │   FastAPI BFF    │
│ Frontend (SPA) │ <────────────────────────────────────── │     Proxy        │
│                │   Unified Response Envelope             │                  │
│                │   { "data": { "rows": [...] } }         │                  │
└────────────────┘                                         └────────┬─────────┘
                                                                    │
                                            Native Async Drivers    │ (Connection Pool,
                                            (asyncpg / aiomysql)    │  Fernet Credential
                                                                    │  Decryption, Timeouts)
                                                                    ▼
                                                           ┌──────────────────┐
                                                           │  Target Database │
                                                           │ (Postgres/MySQL) │
                                                           └──────────────────┘
```

### Proxy & Query Execution Lifecycle
1. **Request Intake**: The frontend transmits a structured query request (`sql`, optional pagination `limit`/`offset`, and `timeout_seconds`) to `POST /api/v1/instances/{id}/query`.
2. **Metadata Lookup & Decryption**: The BFF retrieves the target `DatabaseInstance` record from the internal SQLite database (`app_metadata.db`). The encrypted credentials are decrypted in memory using `cryptography.fernet.Fernet`.
3. **Connection Handshake**: The proxy initializes or reuses an asynchronous database client connection (`asyncpg` for PostgreSQL, `aiomysql` for MySQL, or `aiosqlite` for SQLite).
4. **Execution & Timeout Guardrails**: The query runs within an `asyncio.wait_for(...)` execution wrapper. If the target query exceeds the configured timeout threshold (default: 30 seconds), the proxy aborts the database call, cleans up connections, and issues an HTTP 408 / `QUERY_TIMEOUT` error.
5. **Streaming & Serialization**: Result set metadata (column names, inferred data types) and row tuples are converted into JSON-serializable dictionaries and encapsulated within the unified response envelope.

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

The application persists internal state (projects, configurations, encrypted credentials) in an embedded SQLite database using SQLAlchemy 2.0 (async) and `aiosqlite`.

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
│ encrypted_credentials (TEXT) (Fernet ciphertext)
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
| `encrypted_credentials` | `TEXT` | NOT NULL | Fernet-encrypted JSON (password, token, connection options) |
| `ssl_mode` | `VARCHAR(30)` | NOT NULL, DEFAULT 'prefer' | Target TLS / SSL verification policy |
| `is_read_only` | `BOOLEAN` | NOT NULL, DEFAULT FALSE | Enforces read-only query guardrails |
| `created_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Creation timestamp |
| `updated_at` | `DATETIME` | NOT NULL, DEFAULT UTC | Modification timestamp |

---

## 5. Security Model & Credential Protection

### Symmetric Encryption at Rest (Fernet)
Database passwords, authentication tokens, and sensitive connection strings MUST NEVER be stored in plain text.
- Tathya-Avalokan uses Python's standard `cryptography.fernet.Fernet` (AES-128 in CBC mode with PKCS7 padding and HMAC-SHA256 authentication).
- Encryption keys are retrieved from the `TATHYA_ENCRYPTION_KEY` environment variable.
- During development, if no key is provided, a deterministic development key is derived or generated with an explicit warning in application logs.

```python
# Conceptual Flow
cipher = Fernet(ENCRYPTION_KEY)
# Storage
encrypted_blob = cipher.encrypt(json.dumps({"password": raw_password}).encode()).decode()
# Retrieval in Proxy
decrypted_data = json.loads(cipher.decrypt(encrypted_blob.encode()).decode())
```

### Credential Masking in API Responses
- Passwords and raw encrypted blobs are **strictly excluded** from API responses.
- Responses return `is_password_set: true` instead of exposing sensitive secrets.
- Any generated connection string displays are masked (e.g., `postgresql://dbuser:********@prod-db.internal:5432/orders`).

### Execution Guardrails
- **Statement Timeouts**: Every proxy query has a hard timeout limit enforced by `asyncio.wait_for` (configurable, default 30s).
- **Read-Only Mode Enforcement**: Instances configured as `is_read_only=True` reject mutating statements (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`).
- **CORS Protection**: Restricted to trusted frontend origins.

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

### Recommended Hardening for Shared / Network Deployment

If the application is deployed in a shared or team context, the following authentication path is recommended:

| Concern | Recommended Approach |
|---|---|
| API Authentication | Bearer token middleware (e.g. `python-jose` JWT or `authlib`) |
| Identity Provider | Any OIDC-compatible IdP (Keycloak, Auth0, Google) |
| Per-User Data Isolation | Add `user_id` FK to the `projects` table |
| Frontend Token Storage | `sessionStorage` only (never `localStorage` for sensitive tokens) |
| CORS | Restrict `allow_origins` to the exact frontend hostname |

This authentication layer is deferred to a post-Phase 6 hardening milestone and is **explicitly out of scope** for local developer use cases.

---

## 8. Connection Pool Strategy (Phase 3 Design)

When Phase 3 introduces live database driver execution (`asyncpg`, `aiomysql`), the proxy layer will manage async connection pools per `DatabaseInstance`. The following design is planned:

### Per-Instance Pool Design

```text
FastAPI Request
    │
    ▼
ConnectionPoolRegistry (in-memory dict, instance_id → Pool)
    │
    ├── asyncpg.Pool (PostgreSQL)  — min_size=1, max_size=10
    ├── aiomysql.Pool (MySQL)      — minsize=1, maxsize=10
    └── aiosqlite conn (SQLite)    — single connection, serialized
```

### Pool Lifecycle

| Event | Action |
|---|---|
| First query on an instance | Pool created and cached in registry |
| Subsequent queries | Existing pool connection acquired |
| Instance deleted | Pool evicted from registry, connections drained |
| Server shutdown | All pools gracefully closed in lifespan teardown |
| Query timeout | Connection returned/closed; not poisoned in pool |

### Configuration Parameters (Phase 3)

```python
# Planned environment variable controls per driver
PROXY_POOL_MAX_SIZE = 10      # Maximum concurrent connections per instance
PROXY_POOL_MIN_SIZE = 1       # Minimum idle connections maintained
PROXY_POOL_MAX_IDLE_SEC = 300 # Evict connections idle longer than this
PROXY_QUERY_TIMEOUT_SEC = 30  # Default asyncio.wait_for timeout
```

---

## 9. Result Pagination Design

The `QueryRequest` and `QueryResponse` schemas are **already fully defined** to support pagination:

- `QueryRequest.limit` — Maximum rows to return in a single batch (default: 100, max: 5000)
- `QueryRequest.offset` — Row offset for keyset/offset pagination
- `QueryResponse.has_more` — Whether rows beyond `limit` exist

### Implementation Status

| Component | Status |
|---|---|
| Pydantic schemas | ✅ Defined |
| API contract | ✅ Documented in `api_spec.md` |
| Backend executor wiring | ⏳ Phase 3 — will apply `LIMIT`/`OFFSET` in the proxy SQL wrapper |
| Frontend UI controls | ⏳ Phase 5 — TanStack Table pagination will pass `offset` as the user pages |

The pagination contract is fixed and will not change shape; only the executor wiring is deferred.

---

## 10. `instances_count` — Live Aggregate vs. Stored Counter

The `instances_count` field returned by `GET /api/v1/projects` and `PATCH /api/v1/projects/{id}` is a **live aggregate**, not a stored denormalized counter.

### How it works

The `Project` ORM model uses `lazy="selectin"` on the `instances` relationship:

```python
instances: Mapped[list["DatabaseInstance"]] = relationship(
    "DatabaseInstance",
    back_populates="project",
    cascade="all, delete-orphan",
    lazy="selectin",  # SQLAlchemy fires a SELECT IN query alongside the parent
)
```

The router then computes: `instances_count = len(project.instances)`.

### Guarantees

- **Always accurate** — derived from live FK rows, no separate counter to drift
- **Cascade safe** — deleting a project cascades at the DB level; the count naturally hits 0
- **No race condition** — single SQLite database with serialized writes eliminates counter inconsistency

This design is appropriate for the current SQLite persistence layer. For a high-throughput PostgreSQL metadata store, a denormalized counter with triggers or a `SELECT COUNT(*)` subquery may perform better.
