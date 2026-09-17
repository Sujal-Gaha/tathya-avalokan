import os
from collections.abc import AsyncGenerator

from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

# Internal metadata database location (SQLite via aiosqlite)
METADATA_DATABASE_URL = os.getenv("METADATA_DATABASE_URL", "sqlite+aiosqlite:///./app_metadata.db")

# SQLite concurrency configuration
connect_args = {}
if "sqlite" in METADATA_DATABASE_URL:
    connect_args["check_same_thread"] = False

async_engine = create_async_engine(
    METADATA_DATABASE_URL,
    echo=os.getenv("SQL_DEBUG", "False").lower() in ("true", "1"),
    connect_args=connect_args,
    future=True,
)

AsyncSessionLocal = async_sessionmaker(
    bind=async_engine,
    class_=AsyncSession,
    expire_on_commit=False,
    autocommit=False,
    autoflush=False,
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """FastAPI dependency to provide an isolated async database session per request."""
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()


async def init_db() -> None:
    """Initializes the internal SQLite metadata schema on application startup."""
    from tathya_avalokan.models.base import Base

    async with async_engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
