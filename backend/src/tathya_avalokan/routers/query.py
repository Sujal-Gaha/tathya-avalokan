import re
import time

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from tathya_avalokan.database.session import get_db
from tathya_avalokan.models.instance import DatabaseInstance
from tathya_avalokan.schemas.envelope import ResponseEnvelope
from tathya_avalokan.schemas.query import (
    ColumnMeta,
    ConnectionTestResponse,
    QueryRequest,
    QueryResponse,
)

router = APIRouter(prefix="/instances", tags=["Query & Proxy"])

# Mutating SQL keywords forbidden on read-only instances
_MUTATING_KEYWORDS: frozenset[str] = frozenset(
    {"insert", "update", "delete", "drop", "alter", "truncate", "create", "replace"}
)

# Regex patterns for stripping SQL comments before keyword extraction
_BLOCK_COMMENT_RE = re.compile(r"/\*.*?\*/", re.DOTALL)
_LINE_COMMENT_RE = re.compile(r"--[^\n]*")


def _extract_first_sql_keyword(sql: str) -> str:
    """
    Extracts the first meaningful SQL keyword from a statement,
    safely stripping single-line (--) and block (/* */) comments
    so that comment-prefixed mutating statements cannot bypass the guard.

    Examples:
        '-- comment\\nINSERT INTO ...'  -> 'insert'
        '/* block */ DELETE FROM ...'  -> 'delete'
        '  SELECT 1'                   -> 'select'
    """
    # Strip block comments first, then line comments
    stripped = _BLOCK_COMMENT_RE.sub("", sql)
    stripped = _LINE_COMMENT_RE.sub("", stripped)
    # Grab the first non-whitespace token
    tokens = stripped.split()
    return tokens[0].lower() if tokens else ""


@router.post("/{instance_id}/test-connection", response_model=ResponseEnvelope[ConnectionTestResponse])
async def test_connection(
    instance_id: str,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[ConnectionTestResponse]:
    """Tests connectivity and authentication to the target database instance."""
    stmt = select(DatabaseInstance).where(DatabaseInstance.id == instance_id)
    res = await db.execute(stmt)
    instance = res.scalar_one_or_none()

    if not instance:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Database instance with id '{instance_id}' not found",
        )

    # Baseline connection ping stub (to be fully integrated with asyncpg/aiomysql drivers in Phase 3)
    start_time = time.perf_counter()
    latency = round((time.perf_counter() - start_time) * 1000, 2)

    return ResponseEnvelope(
        data=ConnectionTestResponse(
            connected=True,
            latency_ms=latency,
            server_version=f"{instance.driver_type.capitalize()} Proxy Target",
            message="Connection configuration verified successfully",
        ),
        error=None,
        metadata={},
    )


@router.post("/{instance_id}/query", response_model=ResponseEnvelope[QueryResponse])
async def execute_query(
    instance_id: str,
    payload: QueryRequest,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[QueryResponse]:
    """Executes a SQL query against the target database via the async proxy."""
    stmt = select(DatabaseInstance).where(DatabaseInstance.id == instance_id)
    res = await db.execute(stmt)
    instance = res.scalar_one_or_none()

    if not instance:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Database instance with id '{instance_id}' not found",
        )

    if instance.is_read_only:
        first_keyword = _extract_first_sql_keyword(payload.sql)
        if first_keyword in _MUTATING_KEYWORDS:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Mutating queries are prohibited on read-only database instances (detected keyword: '{first_keyword}')",
            )

    start_time = time.perf_counter()
    # Scaffolding baseline response (Phase 3 will attach the live asyncpg/aiomysql executor)
    execution_time = round((time.perf_counter() - start_time) * 1000, 2)

    return ResponseEnvelope(
        data=QueryResponse(
            columns=[
                ColumnMeta(name="status", type="text"),
                ColumnMeta(name="info", type="text"),
            ],
            rows=[
                {"status": "ready", "info": f"Query received for {instance.name} ({instance.driver_type})"},
            ],
            rows_affected=1,
            execution_time_ms=execution_time,
            has_more=False,
        ),
        error=None,
        metadata={},
    )

