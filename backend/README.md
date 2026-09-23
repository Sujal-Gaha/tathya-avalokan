# Tathya-Avalokan Backend (Go)

Go + Chi Backend-for-Frontend (BFF) and Database Proxy service.

## Architecture & Layout
- `cmd/server/main.go`: Application entrypoint, Chi router mounting, graceful shutdown
- `internal/config`: Environment variable loading and defaults
- `internal/database`: Pure Go SQLite engine (`modernc.org/sqlite`) with embedded schema migration
- `internal/crypto`: AES-256-GCM authenticated credential encryption, key derivation, and URI masking
- `internal/guard`: Comment-stripping SQL keyword extractor and read-only guardrails
- `internal/models`: Domain structs, requests, and unified response DTOs
- `internal/repository`: Clean SQLite data access layer for projects and database instances
- `internal/handlers`: HTTP REST controllers implementing `/api/v1`
- `internal/middleware`: CORS and request handling

## Getting Started

### Prerequisites
- Go 1.22+ (Zero CGO required, `CGO_ENABLED=0`)

### Running Locally
```bash
# Copy sample environment configuration
cp .env.example .env

# Run development server
go run ./cmd/server
```

### Running Tests
```bash
go test -v -race ./...
```

### Building Binary
```bash
CGO_ENABLED=0 go build -o tathya-server ./cmd/server
```
