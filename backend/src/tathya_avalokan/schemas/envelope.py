from typing import Any, Generic, TypeVar

from pydantic import BaseModel, Field

T = TypeVar("T")


class ErrorDetail(BaseModel):
    """Standardized API error payload."""

    code: str = Field(..., description="Machine-readable uppercase error code, e.g. NOT_FOUND")
    message: str = Field(..., description="Human-readable description of the error")
    details: dict[str, Any] | None = Field(default=None, description="Additional contextual debug information")


class ResponseEnvelope(BaseModel, Generic[T]):
    """Unified API response envelope consistent with project architectural standards."""

    data: T | None = Field(default=None, description="The payload resource or items, null on failure")
    error: ErrorDetail | None = Field(default=None, description="Error structure, null on success")
    metadata: dict[str, Any] = Field(default_factory=dict, description="Metadata such as pagination or latency")
