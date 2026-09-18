# Tathya-Avalokan Feature Roadmap

This document tracks the planned features and strategic direction of the project.

## 🟢 Active & Immediate (Phase 3 - 5)

These features are integrated into the core development path to ensure a professional-grade experience.

### 1. The Proxy Engine (Core Phase 3)

- **BFF Hard-Guardrails**: SQL AST parsing in the proxy to block mutating queries (`UPDATE`, `DELETE`, `DROP`) on read-only instances, regardless of DB permissions.
- **Connection Pool Registry**: High-performance async connection pooling per `DatabaseInstance`.
- **Audit Logging**: Proxy-level logging of every query, user, and execution time.

### 2. Power-User Toolset

- **Visual Execution Plans**: Converting `EXPLAIN` output into a color-coded Directed Acyclic Graph (DAG) to identify bottlenecks.
- **Environment-Aware Theming**: Aggressive visual cues (Deep Red for Prod, Yellow for Staging) across the entire UI.
- **Query A/B Profiler**: Side-by-side execution and delta analysis of two query versions.

### 3. UX "Magic Moments"

- **Command Center (Cmd+K)**: Universal spotlight search for navigation and command execution.
- **Stateful Snapshots**: Deep-link URLs that preserve the query, the specific result set, and filters.
- **PWA & Metadata Caching**: Instant load times via IndexedDB schema caching.

---

## 🟡 The "Moat" (Long-term / P2)

Strategic features designed to make the tool indispensable for teams and enterprises.

- **Collaborative Notebooks**: Multiplayer cursors and shared SQL notebooks for team-based data investigation.
- **Interactive Schema Graph**: Zoomable node-graph ERD with glowing foreign key edges (Dependent on Phase 6 Introspection API).
- **Plugin Ecosystem**:
  - **Frontend Slots**: Community-driven UI components (Visualizers, Formatters).
  - **Proxy Middleware**: Community-driven logic (Custom Guardrails, Caching).
- **K8s Auto-Discovery**: Automatic detection of database services within a Kubernetes namespace.
- **Public-Readable Playbooks**: Publishing read-only notebooks as living documentation.

---

## 🔴 Backlog (Future / Deferred)

Features deferred to later stages to focus on core stability and professional utility.

- **AI-Native Integration**:
  - Model Context Protocol (MCP) Bridge.
  - AI-augmented result analysis.
  - Ghost-text SQL autocomplete.
- **Cross-Instance Execution**: Parallel query broadcasting across multiple shards.
- **Transformation Export Pipelines**: Pre-export data masking and formatting in the proxy.
