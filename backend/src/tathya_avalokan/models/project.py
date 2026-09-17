import uuid
from typing import TYPE_CHECKING

from sqlalchemy import String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from tathya_avalokan.models.base import Base, TimestampMixin

if TYPE_CHECKING:
    from tathya_avalokan.models.instance import DatabaseInstance


class Project(Base, TimestampMixin):
    """Internal metadata entity representing a user's grouping project."""

    __tablename__ = "projects"

    id: Mapped[str] = mapped_column(
        String(36),
        primary_key=True,
        default=lambda: str(uuid.uuid4()),
    )
    name: Mapped[str] = mapped_column(
        String(100),
        nullable=False,
        index=True,
    )
    description: Mapped[str | None] = mapped_column(
        Text,
        nullable=True,
    )

    # 1-to-many relationship with database instances; cascade deletes
    instances: Mapped[list["DatabaseInstance"]] = relationship(
        "DatabaseInstance",
        back_populates="project",
        cascade="all, delete-orphan",
        lazy="selectin",
    )
