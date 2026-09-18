import logging
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from tathya_avalokan.database.session import init_db
from tathya_avalokan.routers import (
    health_router,
    instances_router,
    projects_router,
    query_router,
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("tathya_avalokan")


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan: initialize metadata database schema on startup."""
    logger.info("Initializing Tathya-Avalokan SQLite metadata schema...")
    await init_db()
    logger.info("Internal metadata schema ready.")
    yield


app = FastAPI(
    title="Tathya-Avalokan API",
    version="0.1.0",
    description="Backend-for-Frontend & Database Proxy API for Tathya-Avalokan",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS Configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Global Unified Response Envelope for HTTP Exceptions
@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException) -> JSONResponse:
    code_map = {
        400: "QUERY_EXECUTION_ERROR",
        403: "READ_ONLY_VIOLATION",
        404: "NOT_FOUND",
        408: "QUERY_TIMEOUT",
        422: "VALIDATION_ERROR",
        502: "CONNECTION_FAILED",
    }
    error_code = code_map.get(exc.status_code, "INTERNAL_ERROR")
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "data": None,
            "error": {
                "code": error_code,
                "message": str(exc.detail),
                "details": None,
            },
            "metadata": {},
        },
    )


# Global Unified Response Envelope for Unhandled Exceptions
@app.exception_handler(Exception)
async def unhandled_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    logger.exception("Unhandled server error processing %s: %s", request.url.path, exc)
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "data": None,
            "error": {
                "code": "INTERNAL_ERROR",
                "message": "An unexpected server error occurred. Check backend logs.",
                "details": {"error_type": type(exc).__name__},
            },
            "metadata": {},
        },
    )


# Mount API v1 Routes
API_V1_PREFIX = "/api/v1"
app.include_router(health_router, prefix=API_V1_PREFIX)
app.include_router(projects_router, prefix=API_V1_PREFIX)
app.include_router(instances_router, prefix=API_V1_PREFIX)
app.include_router(query_router, prefix=API_V1_PREFIX)
