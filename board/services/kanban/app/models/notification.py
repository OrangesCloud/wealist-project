from sqlalchemy import Column, String, Text, Boolean, Enum as SQLEnum, DateTime
from sqlalchemy.dialects.postgresql import UUID as PG_UUID, JSONB
from app.models.base import BaseModel
from app.models.enums import NotificationType

class Notification(BaseModel):
    """알림 테이블 - 사용자에게 전달되는 알림 저장"""
    __tablename__ = "notifications"

    user_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="알림을 받을 사용자 ID (users.id 참조)"
    )
    notification_type = Column(
        SQLEnum(NotificationType),
        nullable=False,
        index=True,
        comment="알림 타입 (TICKET_CREATED, COMMENT_ADDED 등)"
    )
    title = Column(
        String(200),
        nullable=False,
        comment="알림 제목"
    )
    content = Column(
        Text,
        nullable=True,
        comment="알림 내용"
    )
    target_type = Column(
        String(50),
        nullable=True,
        comment="관련 대상 타입 (TICKET, COMMENT, PROJECT 등)"
    )
    target_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="관련 대상 ID (tickets.id, comments.id 등)"
    )
    extra_data = Column(
        JSONB,
        nullable=True,
        comment="추가 정보 (JSON 형식, 예: {\"old_status\": \"OPEN\", \"new_status\": \"IN_PROGRESS\"})"
    )
    is_read = Column(
        Boolean,
        default=False,
        nullable=False,
        index=True,
        comment="읽음 여부"
    )
    read_at = Column(
        DateTime,
        nullable=True,
        comment="읽은 시각"
    )

    def __repr__(self):
        return f"<Notification(id={self.id}, user_id={self.user_id}, type={self.notification_type}, is_read={self.is_read})>"
