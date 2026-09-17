from datetime import datetime
from enum import StrEnum

from pydantic import BaseModel, ConfigDict, Field


class DriverType(StrEnum):
    POSTGRESQL = "postgresql"
    MYSQL = "mysql"
    SQLITE = "sqlite"


class DatabaseInstanceBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100, description="Friendly instance identifier")
    driver_type: DriverType = Field(default=DriverType.POSTGRESQL, description="Database driver engine")
    host: str | None = Field(default=None, max_length=255, description="Host address or hostname")
    port: int | None = Field(default=None, ge=1, le=65535, description="Port number")
    database_name: str = Field(..., min_length=1, max_length=100, description="Database or schema name")
    username: str | None = Field(default=None, max_length=100, description="Database user")
    ssl_mode: str = Field(default="prefer", description="SSL mode policy")
    is_read_only: bool = Field(default=False, description="Enforce read-only query guardrails")


class DatabaseInstanceCreate(DatabaseInstanceBase):
    password: str | None = Field(default=None, description="Raw password (will be Fernet-encrypted at rest)")
    connection_url: str | None = Field(default=None, description="Direct connection string (optional alternative)")


class DatabaseInstanceUpdate(BaseModel):
    """Partial update schema for PATCH /instances/{id}. All fields are optional."""

    name: str | None = Field(default=None, min_length=1, max_length=100, description="Friendly instance identifier")
    host: str | None = Field(default=None, max_length=255, description="Target hostname or IP")
    port: int | None = Field(default=None, ge=1, le=65535, description="Port number")
    database_name: str | None = Field(default=None, min_length=1, max_length=100, description="Database or schema name")
    username: str | None = Field(default=None, max_length=100, description="Database user")
    password: str | None = Field(default=None, description="New password — will be re-encrypted at rest if provided")
    connection_url: str | None = Field(default=None, description="New direct connection string")
    ssl_mode: str | None = Field(default=None, description="SSL mode policy")
    is_read_only: bool | None = Field(default=None, description="Toggle read-only query guardrails")


class DatabaseInstanceResponse(DatabaseInstanceBase):
    id: str
    project_id: str
    is_password_set: bool = Field(
        default=False,
        description="Indicates whether credentials are saved without exposing them",
    )
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)
