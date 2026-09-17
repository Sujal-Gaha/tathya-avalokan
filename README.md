# Tathya-Avalokan (तथ्य अवलोकन)

> **A modern, full-stack, browser-based database client and workspace.**
> Manage projects, connect to diverse database engines, inspect schemas, and execute queries in a sleek, developer-first interface.

[![Python](https://img.shields.io/badge/python-3.12%2B-blue)](backend/)
[![TypeScript](https://img.shields.io/badge/typescript-5.4%2B-blue)](frontend/)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.115%2B-009688)](backend/)
[![React](https://img.shields.io/badge/React-18.3%2B-61DAFB)](frontend/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 🧭 Project Philosophy

Modern developers work across multiple projects, microservices, and database instances. Switching between command-line tools or bulky desktop clients can fragment the workflow. **Tathya-Avalokan** brings the power and familiarity of a VS Code-like database workspace directly to the browser:

1. **Project-to-Instance Hierarchy**: Organize databases by logical projects (e.g. *"E-commerce Platform"*, *"Analytics Engine"*) rather than an unorganized flat list of connection strings.
2. **Backend-for-Frontend (BFF) Security**: Database credentials never leak to the client. The browser communicates exclusively with an authenticated FastAPI BFF proxy.
3. **Encrypted Credentials at Rest**: Connection strings and credentials stored in the internal metadata database are encrypted using Fernet symmetric cryptography (`cryptography.fernet.Fernet`).
4. **Lightweight Internal Metadata**: Application configurations and project/instance state are persisted in an embedded, zero-maintenance SQLite database (`sqlite+aiosqlite:///./app_metadata.db`).
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
│                    FastAPI BFF Proxy                        │
│  • Project & Instance Metadata CRUD                         │
│  • Credential Encryption / Decryption (Fernet)              │
│  • Target Database Connector & Proxy Worker                 │
│  • Query Timeout & Guardrails Enforcement                   │
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
    SQLAlchemy │ Async (aiosqlite)             │ asyncpg / aiomysql / sqlite
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
├── backend/                        # FastAPI BFF & database proxy server
│   ├── pyproject.toml              # Python packaging & dependencies
│   ├── requirements.txt            # Fallback pip requirements
│   ├── alembic.ini                 # Metadata migration configuration
│   └── src/tathya_avalokan/
│       ├── main.py                 # Application factory & CORS
│       ├── database/               # Async SQLAlchemy engine & sessionmaker
│       ├── models/                 # Metadata ORM entities (Project, DatabaseInstance)
│       ├── schemas/                # Pydantic v2 validation models & response envelopes
│       ├── routers/                # API endpoints (health, projects, instances, query)
│       └── utils/                  # Cryptography & Fernet credential encryption
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
- Python 3.12+ (with `pip` or `uv`)
- Node.js 18+ and `pnpm` (or `npm`)

### 1. Backend Setup
```bash
cd backend
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Run the API server with auto-reload (default: http://localhost:8000)
uvicorn tathya_avalokan.main:app --reload --port 8000
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
- [ ] **Phase 2: Backend Metadata & Projects/Instances API** (CRUD operations, Fernet encryption, SQLite session persistence)
- [ ] **Phase 3: Database Proxy & Driver Engine** (Async connection testing and query execution for PostgreSQL and MySQL)
- [ ] **Phase 4: Frontend Workspace & Monaco Editor** (Project sidebar, schema explorer, query execution workbench)
- [ ] **Phase 5: Data Grid & Result Export** (Virtualized table viewer, pagination, CSV/JSON export)
- [ ] **Phase 6: Schema Visualizer & Autocomplete** (Table columns, indexes, foreign keys, SQL schema-aware autocompletion)
