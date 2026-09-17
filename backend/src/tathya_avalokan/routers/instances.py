from typing import Any

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from tathya_avalokan.database.session import get_db
from tathya_avalokan.models.instance import DatabaseInstance
from tathya_avalokan.models.project import Project
from tathya_avalokan.schemas.envelope import ResponseEnvelope
from tathya_avalokan.schemas.instance import DatabaseInstanceCreate, DatabaseInstanceResponse, DatabaseInstanceUpdate
from tathya_avalokan.utils.security import decrypt_credentials, encrypt_credentials

router = APIRouter(tags=["Instances"])


@router.post(
    "/projects/{project_id}/instances",
    response_model=ResponseEnvelope[DatabaseInstanceResponse],
    status_code=status.HTTP_201_CREATED,
)
async def create_instance(
    project_id: str,
    payload: DatabaseInstanceCreate,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[DatabaseInstanceResponse]:
    """Registers a database instance under an existing project."""
    # Verify parent project exists
    project_stmt = select(Project).where(Project.id == project_id)
    project_res = await db.execute(project_stmt)
    if not project_res.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Parent project with id '{project_id}' does not exist",
        )

    # Encrypt credentials dictionary
    credentials = {
        "password": payload.password or "",
        "connection_url": payload.connection_url or "",
    }
    encrypted_blob = encrypt_credentials(credentials)

    instance = DatabaseInstance(
        project_id=project_id,
        name=payload.name,
        driver_type=payload.driver_type.value,
        host=payload.host,
        port=payload.port,
        database_name=payload.database_name,
        username=payload.username,
        encrypted_credentials=encrypted_blob,
        ssl_mode=payload.ssl_mode,
        is_read_only=payload.is_read_only,
    )
    db.add(instance)
    await db.flush()
    await db.refresh(instance)

    return ResponseEnvelope(
        data=DatabaseInstanceResponse(
            id=instance.id,
            project_id=instance.project_id,
            name=instance.name,
            driver_type=instance.driver_type,  # type: ignore[arg-type]
            host=instance.host,
            port=instance.port,
            database_name=instance.database_name,
            username=instance.username,
            ssl_mode=instance.ssl_mode,
            is_read_only=instance.is_read_only,
            is_password_set=bool(payload.password or payload.connection_url),
            created_at=instance.created_at,
            updated_at=instance.updated_at,
        ),
        error=None,
        metadata={},
    )


@router.get("/instances/{instance_id}", response_model=ResponseEnvelope[DatabaseInstanceResponse])
async def get_instance(
    instance_id: str,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[DatabaseInstanceResponse]:
    """Retrieves details for a database instance with masked credentials."""
    stmt = select(DatabaseInstance).where(DatabaseInstance.id == instance_id)
    res = await db.execute(stmt)
    instance = res.scalar_one_or_none()

    if not instance:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Database instance with id '{instance_id}' not found",
        )

    return ResponseEnvelope(
        data=DatabaseInstanceResponse(
            id=instance.id,
            project_id=instance.project_id,
            name=instance.name,
            driver_type=instance.driver_type,  # type: ignore[arg-type]
            host=instance.host,
            port=instance.port,
            database_name=instance.database_name,
            username=instance.username,
            ssl_mode=instance.ssl_mode,
            is_read_only=instance.is_read_only,
            is_password_set=bool(instance.encrypted_credentials),
            created_at=instance.created_at,
            updated_at=instance.updated_at,
        ),
        error=None,
        metadata={},
    )


@router.delete("/instances/{instance_id}", response_model=ResponseEnvelope[dict[str, Any]])
async def delete_instance(
    instance_id: str,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[dict[str, Any]]:
    """Deletes a database instance."""
    stmt = select(DatabaseInstance).where(DatabaseInstance.id == instance_id)
    res = await db.execute(stmt)
    instance = res.scalar_one_or_none()

    if not instance:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Database instance with id '{instance_id}' not found",
        )

    await db.delete(instance)
    return ResponseEnvelope(
        data={"deleted": True, "id": instance_id},
        error=None,
        metadata={},
    )


@router.patch("/instances/{instance_id}", response_model=ResponseEnvelope[DatabaseInstanceResponse])
async def update_instance(
    instance_id: str,
    payload: DatabaseInstanceUpdate,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[DatabaseInstanceResponse]:
    """Partially updates a database instance configuration. Re-encrypts credentials if a new password is provided."""
    stmt = select(DatabaseInstance).where(DatabaseInstance.id == instance_id)
    res = await db.execute(stmt)
    instance = res.scalar_one_or_none()

    if not instance:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Database instance with id '{instance_id}' not found",
        )

    updates = payload.model_dump(exclude_none=True)
    if not updates:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="No update fields provided in request body",
        )

    # Re-encrypt credentials if a new password or connection_url is being set
    new_password = updates.pop("password", None)
    new_connection_url = updates.pop("connection_url", None)
    if new_password is not None or new_connection_url is not None:
        existing = decrypt_credentials(instance.encrypted_credentials)
        if new_password is not None:
            existing["password"] = new_password
        if new_connection_url is not None:
            existing["connection_url"] = new_connection_url
        instance.encrypted_credentials = encrypt_credentials(existing)

    for field, value in updates.items():
        setattr(instance, field, value)

    await db.flush()
    await db.refresh(instance)

    return ResponseEnvelope(
        data=DatabaseInstanceResponse(
            id=instance.id,
            project_id=instance.project_id,
            name=instance.name,
            driver_type=instance.driver_type,  # type: ignore[arg-type]
            host=instance.host,
            port=instance.port,
            database_name=instance.database_name,
            username=instance.username,
            ssl_mode=instance.ssl_mode,
            is_read_only=instance.is_read_only,
            is_password_set=bool(instance.encrypted_credentials),
            created_at=instance.created_at,
            updated_at=instance.updated_at,
        ),
        error=None,
        metadata={},
    )
