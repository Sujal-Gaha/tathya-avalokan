import uuid
from typing import TYPE_CHECKING

from sqlalchemy import Boolean, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from tathya_avalokan.models.base import Base, TimestampMixin

if TYPE_CHECKING:
    from tathya_avalokan.models.project import Project


class DatabaseInstance(Base, TimestampMixin):
    """Internal metadata entity representing a configured database connection."""

    __tablename__ = "database_instances"

    id: Mapped[str] = mapped_column(
        String(36),
        primary_key=True,
        default=lambda: str(uuid.uuid4()),
    )
    project_id: Mapped[str] = mapped_column(
        String(36),
        ForeignKey("projects.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    name: Mapped[str] = mapped_column(
        String(100),
        nullable=False,
    )
    driver_type: Mapped[str] = mapped_column(
        String(20),
        nullable=False,
        default="postgresql",
    )
    host: Mapped[str | None] = mapped_column(
        String(255),
        nullable=True,
    )
    port: Mapped[int | None] = mapped_column(
        Integer,
        nullable=True,
    )
    database_name: Mapped[str] = mapped_column(
        String(100),
        nullable=False,
    )
    username: Mapped[str | None] = mapped_column(
        String(100),
        nullable=True,
    )
    # Symmetrically encrypted JSON string via Fernet
    encrypted_credentials: Mapped[str] = mapped_column(
        Text,
        nullable=False,
    )
    ssl_mode: Mapped[str] = mapped_column(
        String(30),
        nullable=False,
        default="prefer",
    )
    is_read_only: Mapped[bool] = mapped_column(
        Boolean,
        nullable=False,
        default=False,
    )

    # Relationship back to parent project
    project: Mapped["Project"] = relationship(
        "Project",
        back_populates="instances",
    )
