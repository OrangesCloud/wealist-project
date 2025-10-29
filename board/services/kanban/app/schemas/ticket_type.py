from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional
from uuid import UUID

class TicketTypeBase(BaseModel):
    type_name: str = Field(..., min_length=1, max_length=100, description="티켓 타입 이름")
    description: Optional[str] = Field(None, max_length=500, description="타입 설명")
    color: Optional[str] = Field(None, pattern=r"^#[0-9A-Fa-f]{6}$", description="HEX 색상 코드 (예: #FF5733)")
    icon: Optional[str] = Field(None, max_length=50, description="아이콘 이름 (예: code, bug, palette)")
    display_order: Optional[int] = Field(None, ge=0, description="표시 순서")

class TicketTypeCreate(TicketTypeBase):
    """티켓 타입 생성 스키마"""
    pass

class TicketTypeUpdate(BaseModel):
    """티켓 타입 수정 스키마 (모든 필드 optional)"""
    type_name: Optional[str] = Field(None, min_length=1, max_length=100)
    description: Optional[str] = Field(None, max_length=500)
    color: Optional[str] = Field(None, pattern=r"^#[0-9A-Fa-f]{6}$")
    icon: Optional[str] = Field(None, max_length=50)
    display_order: Optional[int] = Field(None, ge=0)

class TicketTypeResponse(TicketTypeBase):
    """티켓 타입 응답 스키마"""
    id: UUID
    project_id: UUID
    is_deleted: bool
    created_at: datetime
    updated_at: datetime
    created_by: UUID
    updated_by: Optional[UUID] = None

    class Config:
        from_attributes = True

class TicketTypeListResponse(BaseModel):
    """티켓 타입 목록 응답 스키마"""
    total: int
    items: list[TicketTypeResponse]
    limit: int
    offset: int

    class Config:
        from_attributes = True
