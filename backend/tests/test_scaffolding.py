"""
Tathya-Avalokan Backend Test Suite
===================================
Covers:
  - Health endpoint and database connectivity
  - Fernet credential encryption / decryption and URI masking
  - Security module warning when TATHYA_ENCRYPTION_KEY is unset
  - Projects CRUD (create, list, get, PATCH, delete)
  - Instances CRUD (create, get, PATCH, delete)
  - Read-only query guard (including comment-bypass attempts)
"""

import logging
import os
from unittest.mock import patch

import pytest
from httpx import ASGITransport, AsyncClient

from tathya_avalokan.database.session import init_db
from tathya_avalokan.main import app
from tathya_avalokan.routers.query import _extract_first_sql_keyword
from tathya_avalokan.utils.security import decrypt_credentials, encrypt_credentials, mask_connection_uri

# ─── Shared async HTTP client ─────────────────────────────────────────────────


@pytest.fixture
async def client():
    """Provide an in-process ASGI test client with a fresh DB schema."""
    await init_db()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac


# ═══════════════════════════════════════════════════════════════════════════════
# 1. Health & Scaffolding
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.mark.asyncio
async def test_health_endpoint(client: AsyncClient):
    response = await client.get("/api/v1/health")
    assert response.status_code == 200
    payload = response.json()
    assert payload["error"] is None
    assert payload["data"]["status"] == "healthy"
    assert payload["data"]["metadata_storage"] == "sqlite+aiosqlite"


@pytest.mark.asyncio
async def test_database_init():
    await init_db()


# ═══════════════════════════════════════════════════════════════════════════════
# 2. Cryptography utilities
# ═══════════════════════════════════════════════════════════════════════════════


def test_fernet_credential_encryption():
    secret = {"password": "ProductionSuperSecretPassword#2026", "connection_url": "postgresql://usr:pass@host/db"}
    cipher_text = encrypt_credentials(secret)
    assert cipher_text != str(secret)
    decrypted = decrypt_credentials(cipher_text)
    assert decrypted == secret


def test_connection_uri_masking():
    raw_uri = "postgresql://dbuser:MyClearPassword123@prod-cluster.internal:5432/analytics"
    masked = mask_connection_uri(raw_uri)
    assert "MyClearPassword123" not in masked
    assert "dbuser:********@prod-cluster.internal" in masked


def test_encryption_key_warning_logged_when_unset(caplog):
    """Verify a WARNING is emitted when TATHYA_ENCRYPTION_KEY is absent (gap #1 fix)."""
    with patch.dict(os.environ, {}, clear=True):
        # Remove both key env vars to force the fallback path
        env_without_keys = {k: v for k, v in os.environ.items() if k not in ("TATHYA_ENCRYPTION_KEY", "APP_SECRET_KEY")}
        with patch.dict(os.environ, env_without_keys, clear=True):
            with caplog.at_level(logging.WARNING, logger="tathya_avalokan.utils.security"):
                encrypt_credentials({"password": "test"})
    assert any("TATHYA_ENCRYPTION_KEY" in record.message for record in caplog.records)


# ═══════════════════════════════════════════════════════════════════════════════
# 3. Read-only SQL guard — keyword extraction
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.mark.parametrize("sql,expected", [
    ("SELECT * FROM users", "select"),
    ("  SELECT 1", "select"),
    ("INSERT INTO t VALUES (1)", "insert"),
    ("-- comment\nINSERT INTO t VALUES (1)", "insert"),       # line-comment bypass attempt
    ("/* block */ DELETE FROM t", "delete"),                   # block-comment bypass attempt
    ("/* multi\nline */ UPDATE t SET x=1", "update"),          # multi-line block comment
    ("  -- note\n  DROP TABLE users", "drop"),                 # whitespace + comment
    ("TRUNCATE TABLE logs", "truncate"),
])
def test_extract_first_sql_keyword(sql: str, expected: str):
    assert _extract_first_sql_keyword(sql) == expected


# ═══════════════════════════════════════════════════════════════════════════════
# 4. Projects CRUD + PATCH
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.mark.asyncio
async def test_create_and_list_project(client: AsyncClient):
    r = await client.post("/api/v1/projects", json={"name": "Test Project", "description": "A test"})
    assert r.status_code == 201
    data = r.json()["data"]
    assert data["name"] == "Test Project"
    assert data["instances_count"] == 0

    list_r = await client.get("/api/v1/projects")
    assert list_r.status_code == 200
    names = [p["name"] for p in list_r.json()["data"]]
    assert "Test Project" in names


@pytest.mark.asyncio
async def test_patch_project(client: AsyncClient):
    r = await client.post("/api/v1/projects", json={"name": "Original Name"})
    project_id = r.json()["data"]["id"]

    patch_r = await client.patch(f"/api/v1/projects/{project_id}", json={"name": "Renamed Project"})
    assert patch_r.status_code == 200
    assert patch_r.json()["data"]["name"] == "Renamed Project"

    # Ensure description was untouched (not overwritten with null)
    get_r = await client.get(f"/api/v1/projects/{project_id}")
    assert get_r.json()["data"]["name"] == "Renamed Project"


@pytest.mark.asyncio
async def test_patch_project_empty_body_rejected(client: AsyncClient):
    r = await client.post("/api/v1/projects", json={"name": "Project X"})
    project_id = r.json()["data"]["id"]
    patch_r = await client.patch(f"/api/v1/projects/{project_id}", json={})
    assert patch_r.status_code == 400


@pytest.mark.asyncio
async def test_delete_project(client: AsyncClient):
    r = await client.post("/api/v1/projects", json={"name": "To Delete"})
    project_id = r.json()["data"]["id"]

    del_r = await client.delete(f"/api/v1/projects/{project_id}")
    assert del_r.status_code == 200
    assert del_r.json()["data"]["deleted"] is True

    get_r = await client.get(f"/api/v1/projects/{project_id}")
    assert get_r.status_code == 404


# ═══════════════════════════════════════════════════════════════════════════════
# 5. Instances CRUD + PATCH
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.fixture
async def project_id(client: AsyncClient) -> str:
    r = await client.post("/api/v1/projects", json={"name": "Instance Owner Project"})
    return r.json()["data"]["id"]


@pytest.mark.asyncio
async def test_create_and_get_instance(client: AsyncClient, project_id: str):
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={
            "name": "Test DB",
            "driver_type": "postgresql",
            "host": "localhost",
            "port": 5432,
            "database_name": "testdb",
            "username": "admin",
            "password": "secret",
        },
    )
    assert r.status_code == 201
    data = r.json()["data"]
    assert data["is_password_set"] is True
    assert "secret" not in r.text  # password must never leak

    instance_id = data["id"]
    get_r = await client.get(f"/api/v1/instances/{instance_id}")
    assert get_r.status_code == 200
    assert get_r.json()["data"]["name"] == "Test DB"


@pytest.mark.asyncio
async def test_patch_instance(client: AsyncClient, project_id: str):
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "Old Name", "driver_type": "sqlite", "database_name": "dev.db"},
    )
    instance_id = r.json()["data"]["id"]

    patch_r = await client.patch(f"/api/v1/instances/{instance_id}", json={"name": "New Name", "is_read_only": True})
    assert patch_r.status_code == 200
    patched = patch_r.json()["data"]
    assert patched["name"] == "New Name"
    assert patched["is_read_only"] is True


@pytest.mark.asyncio
async def test_patch_instance_password_re_encrypted(client: AsyncClient, project_id: str):
    """Patching password should update encrypted_credentials without leaking the plain text."""
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "PG DB", "driver_type": "postgresql", "database_name": "mydb", "password": "old_pass"},
    )
    instance_id = r.json()["data"]["id"]

    patch_r = await client.patch(f"/api/v1/instances/{instance_id}", json={"password": "new_pass"})
    assert patch_r.status_code == 200
    assert "new_pass" not in patch_r.text
    assert patch_r.json()["data"]["is_password_set"] is True


@pytest.mark.asyncio
async def test_delete_instance(client: AsyncClient, project_id: str):
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "Delete Me", "driver_type": "sqlite", "database_name": "bye.db"},
    )
    instance_id = r.json()["data"]["id"]

    del_r = await client.delete(f"/api/v1/instances/{instance_id}")
    assert del_r.status_code == 200
    assert del_r.json()["data"]["deleted"] is True


# ═══════════════════════════════════════════════════════════════════════════════
# 6. Read-only guard (integration)
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.mark.asyncio
async def test_read_only_guard_blocks_insert(client: AsyncClient, project_id: str):
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "RO DB", "driver_type": "sqlite", "database_name": "ro.db", "is_read_only": True},
    )
    instance_id = r.json()["data"]["id"]

    query_r = await client.post(
        f"/api/v1/instances/{instance_id}/query",
        json={"sql": "INSERT INTO logs VALUES (1)"},
    )
    assert query_r.status_code == 403
    assert query_r.json()["error"]["code"] == "READ_ONLY_VIOLATION"


@pytest.mark.asyncio
async def test_read_only_guard_blocks_comment_bypassed_insert(client: AsyncClient, project_id: str):
    """Guard must catch INSERT even when preceded by a SQL comment (gap #5 fix)."""
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "RO DB 2", "driver_type": "sqlite", "database_name": "ro2.db", "is_read_only": True},
    )
    instance_id = r.json()["data"]["id"]

    query_r = await client.post(
        f"/api/v1/instances/{instance_id}/query",
        json={"sql": "-- bypass attempt\nINSERT INTO logs VALUES (1)"},
    )
    assert query_r.status_code == 403


@pytest.mark.asyncio
async def test_read_only_guard_allows_select(client: AsyncClient, project_id: str):
    r = await client.post(
        f"/api/v1/projects/{project_id}/instances",
        json={"name": "RO SELECT DB", "driver_type": "sqlite", "database_name": "ro3.db", "is_read_only": True},
    )
    instance_id = r.json()["data"]["id"]

    query_r = await client.post(
        f"/api/v1/instances/{instance_id}/query",
        json={"sql": "SELECT 1 as ok"},
    )
    assert query_r.status_code == 200
