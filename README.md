# Tathya-Avalokan (तथ्य अवलोकन)

> **A modern, full-stack, browser-based database client and workspace.**
> Manage projects, connect to diverse database engines, inspect schemas, and execute queries in a sleek, developer-first interface.

[![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8)](backend/)
[![Chi Router](https://img.shields.io/badge/router-Chi%20v5-00ADD8)](backend/)
[![TypeScript](https://img.shields.io/badge/typescript-5.4%2B-blue)](frontend/)
[![React](https://img.shields.io/badge/React-18.3%2B-61DAFB)](frontend/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 🧭 Project Philosophy

Modern developers work across multiple projects, microservices, and database instances. Switching between command-line tools or bulky desktop clients can fragment the workflow. **Tathya-Avalokan** brings the power and familiarity of a VS Code-like database workspace directly to the browser:

1. **Project-to-Instance Hierarchy**: Organize databases by logical projects (e.g. _"E-commerce Platform"_, _"Analytics Engine"_) rather than an unorganized flat list of connection strings.
2. **Backend-for-Frontend (BFF) Security**: Database credentials never leak to the client. The browser communicates exclusively with an authenticated Go (Chi) BFF proxy.
3. **Encrypted Credentials at Rest**: Connection strings and credentials stored in the internal metadata database are encrypted using authenticated **AES-256-GCM** symmetric cryptography.
4. **Lightweight Internal Metadata**: Application configurations and project/instance state are persisted in an embedded, zero-maintenance pure Go SQLite database (`modernc.org/sqlite`, zero CGO).
5. **Modern Developer UX**: Monaco code editor with SQL syntax highlighting, virtualized TanStack data grids, schema inspection trees, and low-latency proxy query execution with timeout guardrails.

---

## 🏛️ System Architecture

Tathya-Avalokan adopts a clean **Backend-for-Frontend (BFF)** database proxy architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (SPA)                           │
│  React 18 + TypeScript + Vite + Tailwind CSS                │
│  • Server State: TanStack Query (caching & mutations)       │
│  • Client State: Zustand (active project, tabs, UI state)   │
│  • Workspace: Monaco SQL Editor + TanStack Data Grid        │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTP REST (/api/v1)
                               │ JSON Unified Envelope
┌──────────────────────────────▼──────────────────────────────┐
│                    Go (Chi) BFF Proxy                       │
│  • Project & Instance Metadata CRUD                         │
│  • Credential Encryption / Decryption (AES-256-GCM)         │
│  • Target Database Connector & Proxy Worker                 │
│  • Query Timeout & Guardrails Enforcement                   │
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
    database/  │ modernc.org/sqlite            │ Driver Proxy (Phase 3)
           sql │ (CGO-free Pure Go)            │ pgx / go-sql-driver/mysql / sqlite
               ▼                               ▼
┌──────────────────────────────┐ ┌────────────────────────────┐
│   App Metadata (SQLite)      │ │   Target User Databases    │
│   • projects                 │ │   • PostgreSQL             │
│   • database_instances       │ │   • MySQL                  │
│   (app_metadata.db)          │ │   • SQLite                 │
└──────────────────────────────┘ └────────────────────────────┘
```

For in-depth architectural details, refer to [`docs/architecture.md`](docs/architecture.md).
For the comprehensive API contract, refer to [`docs/api_spec.md`](docs/api_spec.md).

---

## 🗂️ Monorepo Structure

```text
tathya-avalokan/
├── docs/                           # Architecture specs, API guidelines, security models
│   ├── architecture.md
│   └── api_spec.md
├── backend/                        # Go (Chi) BFF & database proxy server
│   ├── go.mod                      # Go module dependencies
│   ├── go.sum                      # Go checksums
│   ├── .env.example                # Sample environment configuration
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # Application entrypoint & HTTP server
│   └── internal/
│       ├── config/                 # Environment configuration
│       ├── database/               # SQLite connection & embedded schema
│       ├── models/                 # Domain structs & DTOs
│       ├── crypto/                 # AES-256-GCM encryption & URI masking
│       ├── response/               # Standardized JSON response envelope
│       ├── repository/             # SQLite data access layer
│       ├── guard/                  # SQL keyword extractor & read-only guard
│       ├── handlers/               # HTTP REST handlers (/api/v1)
│       └── middleware/             # CORS and request middleware
└── frontend/                       # React + TypeScript single-page application
    ├── package.json                # Dependencies (Monaco, TanStack Table/Query, Zustand)
    ├── vite.config.ts              # Vite configuration & dev proxy
    ├── tsconfig.json               # Strict TypeScript configuration
    └── src/
        ├── components/             # Reusable UI primitives & layout shells
        ├── features/               # Domain feature modules (projects, instances, query)
        ├── services/               # Strongly typed API client & services
        ├── store/                  # Global Zustand UI stores
        └── types/                  # Shared TypeScript interfaces
```

---

## 🚀 Quick Start (Development)

### Prerequisites

- Go 1.22+ (CGO-free, `CGO_ENABLED=0`)
- Node.js 18+ and `pnpm` (or `npm`)

### 1. Backend Setup

```bash
cd backend
cp .env.example .env

# Run the API server (default: http://localhost:8000)
go run ./cmd/server
```

### 2. Frontend Setup

```bash
cd frontend
pnpm install

# Start the Vite development server (default: http://localhost:5173)
pnpm run dev
```

---

## 📅 Roadmap

- [x] **Phase 1: Project Scaffolding & Architecture Documentation** (Foundational setup, specifications, and structure)
- [x] **Phase 2: Backend Metadata & Projects/Instances API in Go** (Chi router, AES-256-GCM encryption, modernc SQLite persistence)
- [ ] **Phase 3: Database Proxy & Driver Engine** (Connection testing and query execution for PostgreSQL and MySQL)
- [ ] **Phase 4: Frontend Workspace & Monaco Editor** (Project sidebar, schema explorer, query execution workbench)
- [ ] **Phase 5: Data Grid & Result Export** (Virtualized table viewer, pagination, CSV/JSON export)
- [ ] **Phase 6: Schema Visualizer & Autocomplete** (Table columns, indexes, foreign key, SQL schema-aware autocompletion)
