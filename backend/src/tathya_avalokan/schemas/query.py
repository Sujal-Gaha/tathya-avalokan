from typing import Any

from pydantic import BaseModel, Field


class QueryRequest(BaseModel):
    sql: str = Field(..., min_length=1, description="SQL statement to execute")
    limit: int = Field(default=100, ge=1, le=5000, description="Maximum rows to return in a single batch")
    offset: int = Field(default=0, ge=0, description="Offset for paginated query results")
    timeout_seconds: int = Field(default=30, ge=1, le=120, description="Hard timeout limit for statement execution")


class ColumnMeta(BaseModel):
    name: str = Field(..., description="Column name or alias")
    type: str = Field(default="text", description="Data type representation")


class QueryResponse(BaseModel):
    columns: list[ColumnMeta] = Field(default_factory=list, description="Column definitions")
    rows: list[dict[str, Any]] = Field(default_factory=list, description="Row records formatted as key-value objects")
    rows_affected: int = Field(default=0, description="Number of affected or returned rows")
    execution_time_ms: float = Field(default=0.0, description="Execution duration in milliseconds")
    has_more: bool = Field(default=False, description="Whether additional records exist beyond limit")


class ConnectionTestResponse(BaseModel):
    connected: bool = Field(..., description="Whether connection handshake succeeded")
    latency_ms: float = Field(default=0.0, description="Roundtrip ping latency in milliseconds")
    server_version: str | None = Field(default=None, description="Target database engine and version banner")
    message: str = Field(default="", description="Status or connection feedback message")
