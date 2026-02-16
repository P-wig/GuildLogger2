from __future__ import annotations

from pydantic import BaseModel, Field


class HardwareCreate(BaseModel):
    """Schema for creating a new hardware set."""

    hardwareName: str = Field(
        ..., min_length=1, max_length=200, description="Display name of hardware"
    )
    capacity: int = Field(
        ..., ge=1, description="Total stock / full capacity of this hardware set"
    )


class HardwareUpdate(BaseModel):
    """Schema for partially updating a hardware set."""

    hardwareName: str | None = Field(None, min_length=1, max_length=200)
    capacity: int | None = Field(None, ge=1)


class HardwareCheckout(BaseModel):
    """Schema for checking out hardware on behalf of a project."""

    projectId: str = Field(..., min_length=1, description="Project to checkout for")
    amount: int = Field(..., ge=1, description="Number of units to checkout")
    userId: str = Field(..., min_length=1, description="User performing the checkout")


class HardwareCheckin(BaseModel):
    """Schema for checking in hardware on behalf of a project."""

    projectId: str = Field(..., min_length=1, description="Project to checkin for")
    amount: int = Field(..., ge=1, description="Number of units to checkin")
    userId: str = Field(..., min_length=1, description="User performing the checkin")
