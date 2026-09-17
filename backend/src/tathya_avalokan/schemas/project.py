from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field

from tathya_avalokan.schemas.instance import DatabaseInstanceResponse


class ProjectBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100, description="Project display name")
    description: str | None = Field(default=None, description="Optional description of the project")


class ProjectCreate(ProjectBase):
    pass


class ProjectUpdate(BaseModel):
    name: str | None = Field(default=None, min_length=1, max_length=100)
    description: str | None = Field(default=None)


class ProjectResponse(ProjectBase):
    id: str
    instances_count: int = Field(default=0, description="Total number of database instances in this project")
    instances: list[DatabaseInstanceResponse] | None = Field(
        default=None,
        description="Nested instances when requested",
    )
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)
