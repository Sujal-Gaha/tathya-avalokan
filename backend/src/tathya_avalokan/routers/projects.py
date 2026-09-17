from typing import Any

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from tathya_avalokan.database.session import get_db
from tathya_avalokan.models.project import Project
from tathya_avalokan.schemas.envelope import ResponseEnvelope
from tathya_avalokan.schemas.project import ProjectCreate, ProjectResponse, ProjectUpdate

router = APIRouter(prefix="/projects", tags=["Projects"])


@router.post("", response_model=ResponseEnvelope[ProjectResponse], status_code=status.HTTP_201_CREATED)
async def create_project(
    payload: ProjectCreate,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[ProjectResponse]:
    """Creates a new project."""
    project = Project(
        name=payload.name,
        description=payload.description,
    )
    db.add(project)
    await db.flush()
    await db.refresh(project)

    return ResponseEnvelope(
        data=ProjectResponse(
            id=project.id,
            name=project.name,
            description=project.description,
            instances_count=0,
            instances=[],
            created_at=project.created_at,
            updated_at=project.updated_at,
        ),
        error=None,
        metadata={},
    )


@router.get("", response_model=ResponseEnvelope[list[ProjectResponse]])
async def list_projects(
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[list[ProjectResponse]]:
    """Lists all projects with associated instance counts."""
    stmt = select(Project).order_by(Project.created_at.desc())
    result = await db.execute(stmt)
    projects = result.scalars().all()

    items = [
        ProjectResponse(
            id=p.id,
            name=p.name,
            description=p.description,
            instances_count=len(p.instances),
            instances=None,
            created_at=p.created_at,
            updated_at=p.updated_at,
        )
        for p in projects
    ]

    return ResponseEnvelope(
        data=items,
        error=None,
        metadata={"total_count": len(items)},
    )


@router.get("/{project_id}", response_model=ResponseEnvelope[ProjectResponse])
async def get_project(
    project_id: str,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[ProjectResponse]:
    """Retrieves a single project with its database instances."""
    stmt = select(Project).where(Project.id == project_id)
    result = await db.execute(stmt)
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project with id '{project_id}' was not found",
        )

    return ResponseEnvelope(
        data=ProjectResponse.model_validate(project),
        error=None,
        metadata={},
    )


@router.delete("/{project_id}", response_model=ResponseEnvelope[dict[str, Any]])
async def delete_project(
    project_id: str,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[dict[str, Any]]:
    """Deletes a project and all associated database instances."""
    stmt = select(Project).where(Project.id == project_id)
    result = await db.execute(stmt)
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project with id '{project_id}' was not found",
        )

    await db.delete(project)
    return ResponseEnvelope(
        data={"deleted": True, "id": project_id},
        error=None,
        metadata={},
    )


@router.patch("/{project_id}", response_model=ResponseEnvelope[ProjectResponse])
async def update_project(
    project_id: str,
    payload: ProjectUpdate,
    db: AsyncSession = Depends(get_db),
) -> ResponseEnvelope[ProjectResponse]:
    """Partially updates a project's name or description."""
    stmt = select(Project).where(Project.id == project_id)
    result = await db.execute(stmt)
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project with id '{project_id}' was not found",
        )

    updates = payload.model_dump(exclude_none=True)
    if not updates:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="No update fields provided in request body",
        )

    for field, value in updates.items():
        setattr(project, field, value)

    await db.flush()
    await db.refresh(project)

    return ResponseEnvelope(
        data=ProjectResponse(
            id=project.id,
            name=project.name,
            description=project.description,
            instances_count=len(project.instances),
            instances=None,
            created_at=project.created_at,
            updated_at=project.updated_at,
        ),
        error=None,
        metadata={},
    )
