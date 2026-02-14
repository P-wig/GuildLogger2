from __future__ import annotations

from pydantic import BaseModel, Field


class ProjectCreate(BaseModel):
    """Schema for creating a new project."""

    projectId: str = Field(
        ..., min_length=1, max_length=100, description="Unique project identifier"
    )
    projectName: str = Field(
        ..., min_length=1, max_length=200, description="Display name of project"
    )
    description: str = Field("", max_length=1000, description="Project description")
    ownerUserId: str = Field(
        ..., min_length=1, description="User ID of project creator"
    )


class ProjectJoin(BaseModel):
    """Schema for joining a project."""

    userId: str = Field(..., min_length=1, description="User ID to add to project")


class ProjectLeave(BaseModel):
    """Schema for leaving a project."""

    userId: str = Field(..., min_length=1, description="User ID to remove from project")


class ProjectUpdate(BaseModel):
    """Schema for partially updating a project."""

    projectName: str | None = Field(None, min_length=1, max_length=200)
    description: str | None = Field(None, max_length=1000)
