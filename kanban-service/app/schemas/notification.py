from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional
from uuid import UUID
from app.models.enums import NotificationType

class NotificationBase(BaseModel):
    notification_type: NotificationType = Field(..., description="알림 타입")
    title: str = Field(..., min_length=1, max_length=200, description="알림 제목")
    content: Optional[str] = Field(None, max_length=1000, description="알림 내용")
    target_type: Optional[str] = Field(None, max_length=50, description="대상 타입 (TICKET, COMMENT 등)")
    target_id: Optional[UUID] = Field(None, description="대상 ID")

class NotificationCreate(NotificationBase):
    """알림 생성 스키마"""
    user_id: UUID = Field(..., description="알림 받을 사용자 ID")
    extra_data: Optional[dict] = Field(None, description="추가 정보 (JSON)")

class NotificationResponse(NotificationBase):
    """알림 응답 스키마"""
    id: UUID
    user_id: UUID
    is_read: bool
    read_at: Optional[datetime] = None
    extra_data: Optional[dict] = None
    created_at: datetime
    updated_at: datetime
    created_by: UUID
    updated_by: Optional[UUID] = None

    class Config:
        from_attributes = True

class NotificationListResponse(BaseModel):
    """알림 목록 응답 스키마"""
    total: int
    unread_count: int
    items: list[NotificationResponse]
    limit: int
    offset: int

    class Config:
        from_attributes = True

class NotificationMarkReadResponse(BaseModel):
    """알림 읽음 처리 응답"""
    id: UUID
    is_read: bool
    read_at: datetime

    class Config:
        from_attributes = True

class NotificationUnreadCountResponse(BaseModel):
    """읽지 않은 알림 개수 응답"""
    unread_count: int
