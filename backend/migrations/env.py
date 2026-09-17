"""
Alembic migration environment for Tathya-Avalokan.

Configured to:
  - Use the same SQLAlchemy metadata as the application models
  - Support async SQLite via aiosqlite in online mode
  - Read the database URL from the METADATA_DATABASE_URL environment variable
    (matching the app's session.py configuration)
"""

import asyncio
import os
import sys
from logging.config import fileConfig

from sqlalchemy import pool
from sqlalchemy.engine import Connection
from sqlalchemy.ext.asyncio import async_engine_from_config

from alembic import context

# ----------------------------------------------------------------
# Ensure the src/ directory is on sys.path so the app package
# can be imported cleanly during migration runs.
# ----------------------------------------------------------------
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

# Import application Base and all models so Alembic sees the full metadata graph
from tathya_avalokan.models.base import Base  # noqa: E402
import tathya_avalokan.models.instance  # noqa: E402, F401  (registers DatabaseInstance)
import tathya_avalokan.models.project  # noqa: E402, F401   (registers Project)

# Alembic Config object — provides access to the .ini values
config = context.config

# Override the sqlalchemy.url from the environment variable if set,
# matching the precedence logic in database/session.py
db_url = os.getenv("METADATA_DATABASE_URL", "sqlite+aiosqlite:///./app_metadata.db")
config.set_main_option("sqlalchemy.url", db_url)

# Python logging configuration from alembic.ini
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# The metadata object containing all table definitions for autogenerate
target_metadata = Base.metadata


def run_migrations_offline() -> None:
    """
    Run migrations in 'offline' mode.

    Generates SQL scripts without an active DB connection.
    Useful for generating migration SQL for review or DBA application.
    """
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
        render_as_batch=True,  # Required for SQLite ALTER TABLE support
    )

    with context.begin_transaction():
        context.run_migrations()


def do_run_migrations(connection: Connection) -> None:
    context.configure(
        connection=connection,
        target_metadata=target_metadata,
        render_as_batch=True,  # Required for SQLite ALTER TABLE support
    )

    with context.begin_transaction():
        context.run_migrations()


async def run_async_migrations() -> None:
    """Run migrations using an async engine (required for aiosqlite)."""
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)

    await connectable.dispose()


def run_migrations_online() -> None:
    """Run migrations in 'online' mode against the live database."""
    asyncio.run(run_async_migrations())


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
