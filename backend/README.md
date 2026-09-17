# Tathya-Avalokan Backend

FastAPI Backend-for-Frontend (BFF) and Database Proxy service.

## Structure
- `src/tathya_avalokan/database`: Async SQLite session and engine
- `src/tathya_avalokan/models`: SQLAlchemy ORM models
- `src/tathya_avalokan/schemas`: Pydantic v2 schemas and response envelopes
- `src/tathya_avalokan/routers`: API endpoints
- `src/tathya_avalokan/utils`: Fernet symmetric encryption utilities

## Running the Server
```bash
uvicorn tathya_avalokan.main:app --reload --port 8000
```
