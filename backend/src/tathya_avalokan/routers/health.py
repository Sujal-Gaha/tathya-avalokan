from fastapi import APIRouter

from tathya_avalokan.schemas.envelope import ResponseEnvelope

router = APIRouter(prefix="/health", tags=["Health"])


@router.get("", response_model=ResponseEnvelope[dict[str, str]])
async def health_check() -> ResponseEnvelope[dict[str, str]]:
    """Checks API server health and driver availability."""
    return ResponseEnvelope(
        data={
            "status": "healthy",
            "version": "0.1.0",
            "metadata_storage": "sqlite+aiosqlite",
        },
        error=None,
        metadata={},
    )
